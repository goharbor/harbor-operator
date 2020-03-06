package harbor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
)

const (
	HealthyStatus   = "healthy"
	UnhealthyStatus = "unhealthy"
)

const (
	HarborHealthEndpoint = "/api/health"
)

type ComponentHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type APIHealth struct {
	Status     string            `json:"status"`
	Components []ComponentHealth `json:"components"`
}

func (h *APIHealth) IsHealthy() bool {
	return h.Status == HealthyStatus
}

func (h *APIHealth) IsComponentHealthy(name string) (bool, error) {
	for _, component := range h.Components {
		if component.Name == name {
			return component.Status == HealthyStatus, nil
		}
	}

	return false, errors.New("component not found")
}

func (h *APIHealth) GetUnhealthyComponents() []string {
	var components []string

	for _, component := range h.Components {
		if component.Status != HealthyStatus {
			components = append(components, component.Name)
		}
	}

	return components
}

func (r *Reconciler) GetHealth(ctx context.Context, harbor *goharborv1alpha1.Harbor) (*APIHealth, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "check")
	defer span.Finish()

	config := rest.CopyConfig(r.RestConfig)
	config.APIPath = "api"
	config = rest.AddUserAgent(config, fmt.Sprintf("%s(%s)", r.GetName(), r.GetVersion()))
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
