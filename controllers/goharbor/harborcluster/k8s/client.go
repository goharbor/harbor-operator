package k8s

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// DefaultTimeout is a reasonable timeout to use with the Client.
const DefaultTimeout = 1 * time.Minute

// WrapClient returns a Client that performs requests within DefaultTimeout.
func WrapClient(ctx context.Context, client client.Client) Client {
	return &ClusterClient{
		crClient: client,
		ctx:      ctx,
	}
}

// Client wraps a controller-runtime client to use a
// default context with a timeout if no context is passed.
type Client interface {
	// WithContext returns a client configured to use the provided context on
	// subsequent requests, instead of one created from the preconfigured timeout.
	WithContext(ctx context.Context) Client

	// Get wraps a controller-runtime client.Get call with a context.
	Get(key client.ObjectKey, obj runtime.Object) error
	// List wraps a controller-runtime client.List call with a context.
	List(opts *client.ListOptions, list runtime.Object) error
	// Create wraps a controller-runtime client.Create call with a context.
	Create(obj runtime.Object) error
	// Delete wraps a controller-runtime client.Delete call with a context.
	Delete(obj runtime.Object, opts ...client.DeleteOption) error
	// Update wraps a controller-runtime client.Update call with a context.
	Update(obj runtime.Object) error
}

type ClusterClient struct {
	crClient client.Client
	ctx      context.Context
}

// WithContext returns a client configured to use the provided context on
// subsequent requests, instead of one created from the preconfigured timeout.
func (w *ClusterClient) WithContext(ctx context.Context) Client {
	w.ctx = ctx
	return w
}

// Get wraps a controller-runtime client.Get call with a context.
func (w *ClusterClient) Get(key client.ObjectKey, obj runtime.Object) error {
	return w.crClient.Get(w.ctx, key, obj)
}

// List wraps a controller-runtime client.List call with a context.
func (w *ClusterClient) List(opts *client.ListOptions, list runtime.Object) error {
	return w.crClient.List(w.ctx, list, opts)
}

// Create wraps a controller-runtime client.Create call with a context.
func (w *ClusterClient) Create(obj runtime.Object) error {
	return w.crClient.Create(w.ctx, obj)
}

// Update wraps a controller-runtime client.Update call with a context.
func (w *ClusterClient) Update(obj runtime.Object) error {
	return w.crClient.Update(w.ctx, obj)
}

// Delete wraps a controller-runtime client.Delete call with a context.
func (w *ClusterClient) Delete(obj runtime.Object, opts ...client.DeleteOption) error {
	return w.crClient.Delete(w.ctx, obj, opts...)
}
