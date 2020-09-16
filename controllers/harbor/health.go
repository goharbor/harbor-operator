package harbor

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
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
	span, _ := opentracing.StartSpanFromContext(ctx, "check")
	defer span.Finish()

	// access in-cluster service
	resp, err := http.Get(fmt.Sprintf("http://%s.%s%s", harbor.NormalizeComponentName(goharborv1alpha1.CoreName), harbor.GetNamespace(), HarborHealthEndpoint))
	if err != nil {
		return nil, errors.Wrap(err, "cannot get health response")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get health response body.")
	}

	health := &APIHealth{}
	err = json.Unmarshal(body, health)

	return health, errors.Wrap(err, "unexpected health response")
}
