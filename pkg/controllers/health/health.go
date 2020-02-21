package health

import (
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
