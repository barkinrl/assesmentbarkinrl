package webserver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	annotationKey = "example.org/postgres-cluster"
	templatePath  = "docs/templates/db.yaml"
	outputPath    = "/tmp/generated-db.yaml"
)

// Watches ConfigMaps and applies/removes resources based on their lifecycle events.
func WatchConfigMaps() error {
	var config *rest.Config
	var err error

	// Try in-cluster config first, if it fails use kubeconfig from local
	config, err = rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to get kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	// List existing ConfigMaps at startup
	cms, err := clientset.CoreV1().ConfigMaps("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("ConfigMap list error:", err)
	} else {
		fmt.Println("Existing ConfigMaps:")
		for _, cm := range cms.Items {
			fmt.Println("-", cm.Name)
		}
	}

	// Start watching ConfigMaps
	w, err := clientset.CoreV1().ConfigMaps("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to watch configmaps: %w", err)
	}
	fmt.Println("Watching ConfigMaps...")

	for event := range w.ResultChan() {
		cm, ok := event.Object.(*corev1.ConfigMap)
		if !ok {
			continue
		}
		switch event.Type {
		case watch.Added, watch.Modified:
			// If the ConfigMap has the required annotation, process it
			if val, ok := cm.GetAnnotations()[annotationKey]; ok && val == "true" {
				// Load the template from a specific ConfigMap
				cfgMap, err := clientset.CoreV1().ConfigMaps("template-namespace").Get(context.TODO(), "db-template", metav1.GetOptions{})
				if err != nil {
					fmt.Println("Template ConfigMap get error:", err)
					continue
				}
				templateStr, ok := cfgMap.Data["db.yaml"]
				if !ok {
					fmt.Println("Template not found in ConfigMap")
					continue
				}
				// Parse the template
				tpl, err := template.New("resource").Parse(templateStr)
				if err != nil {
					fmt.Println("Template parse error:", err)
					continue
				}
				data := map[string]interface{}{
					"CLUSTERNAME": "mycluster",
					"NAMESPACE":   "default",
					"SANAME":      "my-service-account",
				}
				var buf bytes.Buffer
				if err := tpl.Execute(&buf, data); err != nil {
					fmt.Println("Template execute error:", err)
					continue
				}
				content := buf.String()
				// Write the processed template to a file
				if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
					fmt.Println("Write error:", err)
					continue
				}
				// Parse YAML to resource object(s)
				decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(content), 4096)
				for {
					var statefulSet appsv1.StatefulSet
					if err := decoder.Decode(&statefulSet); err != nil {
						if err.Error() == "EOF" {
							break
						}
						fmt.Println("YAML decode error:", err)
						continue
					}
					// Apply resource: check if it exists, create if not, update if exists
					_, err := clientset.AppsV1().StatefulSets(cm.Namespace).Get(context.TODO(), statefulSet.Name, metav1.GetOptions{})
					if err != nil {
						// Create if not exists
						_, err = clientset.AppsV1().StatefulSets(cm.Namespace).Create(context.TODO(), &statefulSet, metav1.CreateOptions{})
						if err != nil {
							fmt.Println("StatefulSet create error:", err)
							continue
						}
						fmt.Println("StatefulSet created:", statefulSet.Name)
					} else {
						// Update if exists
						_, err = clientset.AppsV1().StatefulSets(cm.Namespace).Update(context.TODO(), &statefulSet, metav1.UpdateOptions{})
						if err != nil {
							fmt.Println("StatefulSet update error:", err)
							continue
						}
						fmt.Println("StatefulSet updated:", statefulSet.Name)
					}
				}
			}
		case watch.Deleted:
			// When a ConfigMap is deleted, remove related resources
			name := cm.Name
			namespace := cm.Namespace
			fmt.Printf("ConfigMap deleted: %s/%s\n", namespace, name)

			// Scale down the StatefulSet to 0 replicas before deleting
			_, err := clientset.AppsV1().StatefulSets(namespace).UpdateScale(context.TODO(), name, &autoscalingv1.Scale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: autoscalingv1.ScaleSpec{
					Replicas: 0,
				},
			}, metav1.UpdateOptions{})
			if err != nil {
				fmt.Println("StatefulSet scale error:", err)
			}

			// Wait until all pods are terminated
			waitForPodsTerminated := func() {
				for {
					podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
						LabelSelector: fmt.Sprintf("statefulset.kubernetes.io/pod-name in (%s-0)", name),
					})
					if err != nil {
						fmt.Println("Pod list error:", err)
						break
					}
					if len(podList.Items) == 0 {
						fmt.Println("All pods terminated.")
						break
					}
					fmt.Printf("Waiting for pods to terminate... (%d remaining)\n", len(podList.Items))
					// wait for 2 seconds before checking again
					time.Sleep(2 * time.Second)
				}
			}
			waitForPodsTerminated()

			// Delete StatefulSet, Service, and PVC with the same name, retry if error
			maxRetries := 3
			deleteWithRetry := func(delFunc func() error, desc string) {
				for i := 0; i < maxRetries; i++ {
					if err := delFunc(); err != nil {
						fmt.Printf("%s delete error (try %d/%d): %v\n", desc, i+1, maxRetries, err)
						if i == maxRetries-1 {
							fmt.Printf("%s delete failed after retries, not correctable.\n", desc)
						}
					} else {
						fmt.Printf("%s deleted successfully.\n", desc)
						break
					}
				}
			}
			deleteWithRetry(func() error {
				return clientset.AppsV1().StatefulSets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
			}, "StatefulSet")
			deleteWithRetry(func() error {
				return clientset.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
			}, "Service")
			deleteWithRetry(func() error {
				return clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
			}, "PVC")
		}
	}
	return nil
}
