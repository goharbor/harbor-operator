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

// DynamicClientOptions provide options for initializing ClusterDynamicClient.
type DynamicClientOptions struct {
	// Resource namespace
	Namespace string
	// Resource schema
	Resource schema.GroupVersionResource
}

// DynamicClientOption is option template.
type DynamicClientOption func(opt *DynamicClientOptions)

// WithResource option.
func WithResource(resource schema.GroupVersionResource) DynamicClientOption {
	return func(opt *DynamicClientOptions) {
		opt.Resource = resource
	}
}

// WithNamespace option.
func WithNamespace(namespace string) DynamicClientOption {
	return func(opt *DynamicClientOptions) {
		opt.Namespace = namespace
	}
}

// DynamicClientWrapper wraps the dynamic client to DClient.
type DynamicClientWrapper struct {
	dClient dynamic.Interface
}

// DynamicClient returns a DClient copy.
// Required options: WithResource and WithNamespace.
func (d *DynamicClientWrapper) DynamicClient(ctx context.Context, options ...DynamicClientOption) DClient {
	clientOptions := &DynamicClientOptions{}

	for _, op := range options {
		op(clientOptions)
	}

	return &ClusterDynamicClient{
		ctx:       ctx,
		dClient:   d.dClient,
		resource:  clientOptions.Resource,
		namespace: clientOptions.Namespace,
	}
}

// RawClient returns the used dynamic.Interface.
func (d *DynamicClientWrapper) RawClient() dynamic.Interface {
	return d.dClient
}

// DynamicClient returns a dynamic client wrapper.
func DynamicClient() (*DynamicClientWrapper, error) {
	client, err := newDynamicClient()
	if err != nil {
		return nil, err
	}

	return &DynamicClientWrapper{
		dClient: client,
	}, nil
}

// DClient wraps a client-go dynamic.
type DClient interface {
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

// Create wraps a client-go dynamic.Create call with a context.
func (w *ClusterDynamicClient) Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Create(w.ctx, obj, options, subresources...)
}

// Update wraps a client-go dynamic.Update call with a context.
func (w *ClusterDynamicClient) Update(obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Update(w.ctx, obj, options, subresources...)
}

// UpdateStatus wraps a client-go dynamic.UpdateStatus call with a context.
func (w *ClusterDynamicClient) UpdateStatus(obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).UpdateStatus(w.ctx, obj, options)
}

// Delete wraps a client-go dynamic.Delete call with a context.
func (w *ClusterDynamicClient) Delete(name string, options metav1.DeleteOptions, subresources ...string) error {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Delete(w.ctx, name, options, subresources...)
}

// DeleteCollection wraps a client-go dynamic.DeleteCollection call with a context.
func (w *ClusterDynamicClient) DeleteCollection(options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).DeleteCollection(w.ctx, options, listOptions)
}

// Get wraps a client-go dynamic.Get call with a context.
func (w *ClusterDynamicClient) Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Get(w.ctx, name, options, subresources...)
}

// Get wraps a client-go dynamic.Get call with a context.
func (w *ClusterDynamicClient) List(opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).List(w.ctx, opts)
}

// Watch wraps a client-go dynamic.Watch call with a context.
func (w *ClusterDynamicClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Watch(w.ctx, opts)
}

// Patch wraps a client-go dynamic.Patch call with a context.
func (w *ClusterDynamicClient) Patch(name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return w.dClient.Resource(w.resource).Namespace(w.namespace).Patch(w.ctx, name, pt, data, options, subresources...)
}

// newDynamicClient returns the dynamic interface.
func newDynamicClient() (dynamic.Interface, error) {
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

// HomeDir returns home dir.
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
