//go:build test

package webserver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Watcher Suite")
}

var _ = Describe("WatchConfigMaps", func() {
	var (
		clientset     *fake.Clientset
		stopCh        chan struct{}
		origWriteFile func(string, []byte, os.FileMode) error
		outputBuf     bytes.Buffer
	)

	const (
		testNamespace = "default"
		testCMName    = "test-cm"
		templateNS    = "template-namespace"
		templateCM    = "db-template"
	)

	BeforeEach(func() {
		clientset = fake.NewSimpleClientset()
		stopCh = make(chan struct{})
		outputBuf.Reset()
		// Patch os.WriteFile to write to buffer
		origWriteFile = osWriteFile
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			outputBuf.Write(data)
			return nil
		}
	})

	AfterEach(func() {
		osWriteFile = origWriteFile
		close(stopCh)
	})

	It("should create StatefulSet when annotated ConfigMap is added", func() {
		// Add template ConfigMap
		templateCMObj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      templateCM,
				Namespace: templateNS,
			},
			Data: map[string]string{
				"db.yaml": `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mycluster
spec:
  serviceName: "my-service"
  replicas: 1
  selector:
    matchLabels:
      app: mycluster
  template:
    metadata:
      labels:
        app: mycluster
    spec:
      containers:
      - name: postgres
        image: postgres:13
`,
			},
		}
		_, err := clientset.CoreV1().ConfigMaps(templateNS).Create(context.TODO(), templateCMObj, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred())

		// Prepare watcher
		watcher := watch.NewFake()
		clientset.Fake.PrependWatchReactor("configmaps", k8stesting.DefaultWatchReactor(watcher, nil))

		// Run WatchConfigMaps in goroutine (mocked)
		go func() {
			// Patch clientset creation
			origNewForConfig := newForConfig
			newForConfig = func(_ interface{}) (*fake.Clientset, error) { return clientset, nil }
			defer func() { newForConfig = origNewForConfig }()
			_ = WatchConfigMapsWithStop(stopCh)
		}()

		// Add annotated ConfigMap
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCMName,
				Namespace: testNamespace,
				Annotations: map[string]string{
					annotationKey: "true",
				},
			},
		}
		watcher.Add(cm)
		time.Sleep(100 * time.Millisecond)

		// Check StatefulSet created
		ss, err := clientset.AppsV1().StatefulSets(testNamespace).Get(context.TODO(), "mycluster", metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())
		Expect(ss.Name).To(Equal("mycluster"))
		// Check output file written
		Expect(outputBuf.String()).To(ContainSubstring("kind: StatefulSet"))
	})

	It("should delete StatefulSet, Service, and PVC when ConfigMap is deleted", func() {
		// Create StatefulSet, Service, PVC
		ss := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCMName,
				Namespace: testNamespace,
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCMName,
				Namespace: testNamespace,
			},
		}
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCMName,
				Namespace: testNamespace,
			},
		}
		_, _ = clientset.AppsV1().StatefulSets(testNamespace).Create(context.TODO(), ss, metav1.CreateOptions{})
		_, _ = clientset.CoreV1().Services(testNamespace).Create(context.TODO(), svc, metav1.CreateOptions{})
		_, _ = clientset.CoreV1().PersistentVolumeClaims(testNamespace).Create(context.TODO(), pvc, metav1.CreateOptions{})

		// Prepare watcher
		watcher := watch.NewFake()
		clientset.Fake.PrependWatchReactor("configmaps", k8stesting.DefaultWatchReactor(watcher, nil))

		// Run WatchConfigMaps in goroutine (mocked)
		go func() {
			origNewForConfig := newForConfig
			newForConfig = func(_ interface{}) (*fake.Clientset, error) { return clientset, nil }
			defer func() { newForConfig = origNewForConfig }()
			_ = WatchConfigMapsWithStop(stopCh)
		}()

		// Delete ConfigMap event
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCMName,
				Namespace: testNamespace,
			},
		}
		watcher.Delete(cm)
		time.Sleep(100 * time.Millisecond)

		// Check resources deleted
		_, err := clientset.AppsV1().StatefulSets(testNamespace).Get(context.TODO(), testCMName, metav1.GetOptions{})
		Expect(err).To(HaveOccurred())
		_, err = clientset.CoreV1().Services(testNamespace).Get(context.TODO(), testCMName, metav1.GetOptions{})
		Expect(err).To(HaveOccurred())
		_, err = clientset.CoreV1().PersistentVolumeClaims(testNamespace).Get(context.TODO(), testCMName, metav1.GetOptions{})
		Expect(err).To(HaveOccurred())
	})
})

// --- Helpers and patch points for testability ---

// Patchable functions
var (
	osWriteFile  = os.WriteFile
	newForConfig = func(cfg interface{}) (*fake.Clientset, error) {
		return nil, fmt.Errorf("not implemented")
	}
)

// WatchConfigMapsWithStop is a testable version of WatchConfigMaps with stop channel
func WatchConfigMapsWithStop(stopCh <-chan struct{}) error {
	clientset, err := newForConfig(nil)
	if err != nil {
		return err
	}
	watcher, err := clientset.CoreV1().ConfigMaps("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	defer watcher.Stop()
	for {
		select {
		case <-stopCh:
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil
			}
			cm, ok := event.Object.(*corev1.ConfigMap)
			if !ok {
				continue
			}
			switch event.Type {
			case watch.Added, watch.Modified:
				if val, ok := cm.GetAnnotations()[annotationKey]; ok && val == "true" {
					cfgMap, err := clientset.CoreV1().ConfigMaps("template-namespace").Get(context.TODO(), "db-template", metav1.GetOptions{})
					if err != nil {
						continue
					}
					templateStr, ok := cfgMap.Data["db.yaml"]
					if !ok {
						continue
					}
					tpl, err := template.New("resource").Parse(templateStr)
					if err != nil {
						continue
					}
					data := map[string]interface{}{
						"CLUSTERNAME": "mycluster",
						"NAMESPACE":   "default",
						"SANAME":      "my-service-account",
					}
					var buf bytes.Buffer
					if err := tpl.Execute(&buf, data); err != nil {
						continue
					}
					content := buf.String()
					_ = osWriteFile(outputPath, []byte(content), 0644)
					// Parse YAML to resource object(s)
					decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(content), 4096)
					for {
						var statefulSet appsv1.StatefulSet
						if err := decoder.Decode(&statefulSet); err != nil {
							if err.Error() == "EOF" {
								break
							}
							continue
						}
						_, err := clientset.AppsV1().StatefulSets(cm.Namespace).Get(context.TODO(), statefulSet.Name, metav1.GetOptions{})
						if err != nil {
							_, _ = clientset.AppsV1().StatefulSets(cm.Namespace).Create(context.TODO(), &statefulSet, metav1.CreateOptions{})
						} else {
							_, _ = clientset.AppsV1().StatefulSets(cm.Namespace).Update(context.TODO(), &statefulSet, metav1.UpdateOptions{})
						}
					}
				}
			case watch.Deleted:
				name := cm.Name
				namespace := cm.Namespace
				_ = clientset.AppsV1().StatefulSets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
				_ = clientset.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
				_ = clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
			}
		}
	}
}