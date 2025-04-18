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
	"strings"
)

type apiActionResult func()
type apiAction func(w http.ResponseWriter, r *http.Request, data map[string]interface{}) apiActionResult

type securedApiAction struct {
	action   apiAction
	needAuth bool
}

var apiActions = map[string]securedApiAction{
	"get_version": {action: getVersion, needAuth: false},
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
