/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package webserver

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type apiActionResult func()
type apiAction func(w http.ResponseWriter, r *http.Request, data map[string]interface{}) apiActionResult

type securedApiAction struct {
	action   apiAction
	needAuth bool
}

var apiActions = map[string]securedApiAction{
	"get_version":      {action: getVersion, needAuth: false},
	"get_configmaps":   {action: getConfigMaps, needAuth: true},
	"delete_configmap": {action: deleteConfigMap, needAuth: true},
	"get_configmap":    {action: getConfigMap, needAuth: true},
	"update_configmap": {action: updateConfigMap, needAuth: true},
}

func sendError(w http.ResponseWriter, errmsg string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	msg, _ := json.Marshal(map[string]string{"error": errmsg})
	_, err := w.Write([]byte(msg))

	if err != nil {
		slog.Error("Error writing response", "error", err)
	}
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}

	// if method get and data parameter exists in query string
	if (r.Method == "GET" || r.Method == "HEAD") && len(r.URL.Query().Get("data")) > 0 {
		err := json.Unmarshal([]byte(r.URL.Query().Get("data")), &data)

		if err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else if r.Method == "POST" {
		if r.Header.Get("Content-Type") == "application/json" {
			buf, err := io.ReadAll(r.Body)

			if err != nil {
				sendError(w, err.Error(), http.StatusBadRequest)
				return
			}

			err = json.Unmarshal(buf, &data)

			if err != nil {
				sendError(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			err := r.ParseForm()

			if err != nil {
				sendError(w, err.Error(), http.StatusBadRequest)
				return
			}

			data = make(map[string]interface{})

			for key, value := range r.PostForm {
				data[key] = value[0]
			}
		} else {
			sendError(w, "Content-Type is not allowed", http.StatusBadRequest)
			return
		}

	} else {
		sendError(w, "method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	action, ok := data["action"].(string)

	if !ok {
		sendError(w, "action parameter is missing", http.StatusBadRequest)
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// check if action exists
	if _, ok := apiActions[action]; !ok {
		sendError(w, "action parameter is invalid", http.StatusBadRequest)
		return
	}

	// call action
	if apiActions[action].needAuth {
		authHeader := r.Header.Get("Authorization")

		if len(authHeader) == 0 {
			slog.Error("Authorization header is missing")
			sendError(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 {
			slog.Error("Authorization header is invalid")
			sendError(w, "Authorization header is invalid", http.StatusUnauthorized)
			return
		}

		tokenType, token := parts[0], parts[1]

		switch tokenType {
		case "Bearer":
			valid, err := validateToken(token)
			if err != nil {
				slog.Error("Token validation failed", "error", err)
				sendError(w, "Token validation failed", http.StatusUnauthorized)
				return
			} else if valid {
				slog.Info("Token is valid and user is in 'admins' group.")
			} else {
				slog.Error("User is not in 'admins' group")
				sendError(w, "User is not in 'admins' group", http.StatusUnauthorized)
				return
			}
		default:
			slog.Error("Authorization header is invalid")
			sendError(w, "Authorization header is invalid", http.StatusUnauthorized)
			return
		}
	}

	result := apiActions[action].action(w, r, data)

	//chech if w has content type set if not set it to json
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	result()
}

func getConfigMaps(w http.ResponseWriter, r *http.Request, data map[string]interface{}) apiActionResult {
	return func() {
		var config *rest.Config
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				sendError(w, "Kubeconfig error", http.StatusInternalServerError)
				return
			}
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			sendError(w, "Clientset error", http.StatusInternalServerError)
			return
		}
		cms, err := clientset.CoreV1().ConfigMaps("").List(r.Context(), metav1.ListOptions{})
		if err != nil {
			sendError(w, "List error", http.StatusInternalServerError)
			return
		}
		result := []map[string]string{}
		for _, cm := range cms.Items {
			result = append(result, map[string]string{"name": cm.Name})
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			slog.Error("Error encoding response", "error", err)
		}
	}
}

func deleteConfigMap(w http.ResponseWriter, r *http.Request, data map[string]interface{}) apiActionResult {
	return func() {
		name, ok := data["name"].(string)
		if !ok || name == "" {
			sendError(w, "Missing configmap name", http.StatusBadRequest)
			return
		}
		var config *rest.Config
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				sendError(w, "Kubeconfig error", http.StatusInternalServerError)
				return
			}
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			sendError(w, "Clientset error", http.StatusInternalServerError)
			return
		}
		err = clientset.CoreV1().ConfigMaps("").Delete(r.Context(), name, metav1.DeleteOptions{})
		if err != nil {
			sendError(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"deleted"}`)); err != nil {
			slog.Error("Error writing response", "error", err)
		}
	}
}

func getConfigMap(w http.ResponseWriter, r *http.Request, data map[string]interface{}) apiActionResult {
	return func() {
		name, ok := data["name"].(string)
		if !ok || name == "" {
			sendError(w, "Missing configmap name", http.StatusBadRequest)
			return
		}
		var config *rest.Config
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				sendError(w, "Kubeconfig error", http.StatusInternalServerError)
				return
			}
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			sendError(w, "Clientset error", http.StatusInternalServerError)
			return
		}
		cm, err := clientset.CoreV1().ConfigMaps("").Get(r.Context(), name, metav1.GetOptions{})
		if err != nil {
			sendError(w, "Get error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"name": cm.Name,
			"data": cm.Data,
		}); err != nil {
			slog.Error("Error encoding response", "error", err)
		}
	}
}

func updateConfigMap(w http.ResponseWriter, r *http.Request, data map[string]interface{}) apiActionResult {
	return func() {
		name, ok := data["name"].(string)
		if !ok || name == "" {
			sendError(w, "Missing configmap name", http.StatusBadRequest)
			return
		}
		rawData, ok := data["data"]
		if !ok {
			sendError(w, "Missing configmap data", http.StatusBadRequest)
			return
		}
		// Convert rawData to map[string]string
		dataMap := map[string]string{}
		switch v := rawData.(type) {
		case map[string]interface{}:
			for k, val := range v {
				if str, ok := val.(string); ok {
					dataMap[k] = str
				}
			}
		default:
			sendError(w, "Invalid data format", http.StatusBadRequest)
			return
		}

		var config *rest.Config
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				sendError(w, "Kubeconfig error", http.StatusInternalServerError)
				return
			}
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			sendError(w, "Clientset error", http.StatusInternalServerError)
			return
		}
		cm, err := clientset.CoreV1().ConfigMaps("").Get(r.Context(), name, metav1.GetOptions{})
		if err != nil {
			sendError(w, "Get error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		cm.Data = dataMap
		_, err = clientset.CoreV1().ConfigMaps("").Update(r.Context(), cm, metav1.UpdateOptions{})
		if err != nil {
			sendError(w, "Update error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"updated"}`)); err != nil {
			slog.Error("Error writing response", "error", err)
		}
	}
}
