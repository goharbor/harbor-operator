package k8s

import (
	"context"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// WrapDClient returns a Dynamic Client.
func WrapDClient(client dynamic.Interface) DClient {
	return &ClusterDynamicClient{
		dClient: client,
	}
}

// DClient wraps a client-go dynamic.
type DClient interface {
	WithResource(resource schema.GroupVersionResource) DClient
	WithNamespace(namespace string) DClient
	WithContext(ctx context.Context) DClient
	Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error)
	Update(obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error)
	Delete(name string, options metav1.DeleteOptions, subresources ...string) error
	Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
	List(opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

type ClusterDynamicClient struct {
	dClient   dynamic.Interface
	namespace string
	resource  schema.GroupVersionResource
	ctx       context.Context
}

// WithResource returns a client with resource.
func (w *ClusterDynamicClient) WithResource(resource schema.GroupVersionResource) DClient {
	w.resource = resource
	return w
}

// WithNamespace returns a client with namespace.
func (w *ClusterDynamicClient) WithNamespace(namespace string) DClient {
	w.namespace = namespace
	return w
}

// WithContext returns a client with context.
func (w *ClusterDynamicClient) WithContext(ctx context.Context) DClient {
	w.ctx = ctx
	return w
}

// Create wraps a client-go dynamic.Create call with a context.
func (w *ClusterDynamicClient) Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Create(obj, options, subresources...)
}

// Update wraps a client-go dynamic.Update call with a context.
func (w *ClusterDynamicClient) Update(obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Update(obj, options, subresources...)
}

// UpdateStatus wraps a client-go dynamic.UpdateStatus call with a context.
func (w *ClusterDynamicClient) UpdateStatus(obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).UpdateStatus(obj, options)
}

// Delete wraps a client-go dynamic.Delete call with a context.
func (w *ClusterDynamicClient) Delete(name string, options metav1.DeleteOptions, subresources ...string) error {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Delete(name, &options, subresources...)
}

// DeleteCollection wraps a client-go dynamic.DeleteCollection call with a context.
func (w *ClusterDynamicClient) DeleteCollection(options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).DeleteCollection(&options, listOptions)
}

// Get wraps a client-go dynamic.Get call with a context.
func (w *ClusterDynamicClient) Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Get(name, options, subresources...)
}

// Get wraps a client-go dynamic.Get call with a context.
func (w *ClusterDynamicClient) List(opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).List(opts)
}

// Watch wraps a client-go dynamic.Watch call with a context.
func (w *ClusterDynamicClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Watch(opts)
}

// Patch wraps a client-go dynamic.Patch call with a context.
func (w *ClusterDynamicClient) Patch(name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Patch(name, pt, data, options, subresources...)
}

//NewDynamicClient returns the dynamic interface.
func NewDynamicClient() (dynamic.Interface, error) {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		config, err = ExternalConfig()
		if err != nil {
			return nil, err
		}
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return dynamicClient, nil
}

//HomeDir returns home dir
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

// ExternalConfig returns a config object which uses the service account
// kubernetes gives to pods.
func ExternalConfig() (*rest.Config, error) {
	home := HomeDir()
	kubeConfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}
