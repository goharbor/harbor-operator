package lcm

import (
	"context"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Controller is designed to handle the lifecycle of the related incluster deployed services like psql, redis and minio.
type Controller interface {
	// Apply the changes to the cluster including:
	// - create new if the designed resource is not existing
	// - update the resource if the related spec has been changed
	// - scale the resources if the replica is changed
	//
	// Equal to the previous method "Reconcile()" of lcm Controller
	Apply(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*CRStatus, error)

	// Delete the related resources if the resource configuration is removed from the spec.
	// As we support connecting to the external or incluster provisioned dependent services,
	// the dependent service may switch from incluster to external mode and then the incluster
	// services may need to be unloaded.
	Delete(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*CRStatus, error)

	// Upgrade the specified resource to the given version.
	Upgrade(ctx context.Context, harborcluster *v1alpha2.HarborCluster) (*CRStatus, error)

	// HealthChecker returns a health checker implementation for checking the health status of the service managed
	// by this controller.
	HealthChecker() HealthChecker
}

type CRStatus struct {
	Condition  v1alpha2.HarborClusterCondition `json:"condition"`
	Properties Properties                      `json:"properties"`
}

// New returns new CRStatus.
func New(conditionType v1alpha2.HarborClusterConditionType) *CRStatus {
	return &CRStatus{
		Condition: v1alpha2.HarborClusterCondition{
			LastTransitionTime: metav1.Now(),
			Type:               conditionType,
		},
	}
}

// WithStatus returns CRStatus with Condition status.
func (cs *CRStatus) WithStatus(status corev1.ConditionStatus) *CRStatus {
	cs.Condition.Status = status

	return cs
}

// WithReason returns CRStatus with Condition reason.
func (cs *CRStatus) WithReason(reason string) *CRStatus {
	cs.Condition.Reason = reason

	return cs
}

// WithMessage returns CRStatus with Condition message.
func (cs *CRStatus) WithMessage(message string) *CRStatus {
	cs.Condition.Message = message

	return cs
}

// WithProperties returns CRStatus with Properties.
func (cs *CRStatus) WithProperties(properties Properties) *CRStatus {
	cs.Properties = properties

	return cs
}
