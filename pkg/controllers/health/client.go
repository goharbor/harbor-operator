package health

import (
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
)

type Client struct {
	RestConfig *rest.Config

	Scheme *runtime.Scheme
}

func (r *Client) GetByProxy(ctx context.Context, harbor *goharborv1alpha1.Harbor) (*APIHealth, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "check")
	defer span.Finish()

	config := rest.CopyConfig(r.RestConfig)
	config.APIPath = "api"
	config.NegotiatedSerializer = serializer.NewCodecFactory(r.Scheme)
	config.GroupVersion = &corev1.SchemeGroupVersion

	client, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get rest client")
	}

	// https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-services/#manually-constructing-apiserver-proxy-urls

	result, err := client.Get().
		Context(ctx).
		Resource("services").
		Namespace(harbor.GetNamespace()).
		Name(harbor.NormalizeComponentName(goharborv1alpha1.CoreName)).
		SubResource("proxy").
		Suffix(HarborHealthEndpoint).
		DoRaw()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get health response")
	}

	health := &APIHealth{}
	err = json.Unmarshal(result, health)

	return health, errors.Wrap(err, "unexpected health response")
}
