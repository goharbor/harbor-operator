package lcm

import (
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This package container interface of harbor cluster service lifecycle manage.

type Controller interface {
	// Provision the new dependent service by the related section of cluster spec.
	Provision() (*CRStatus, error)

	// Delete the service
	Delete() (*CRStatus, error)

	// Scale will get the replicas of components, and update the crd.
	Scale() (*CRStatus, error)

	// Scale up
	ScaleUp(newReplicas uint64) (*CRStatus, error)

	// Scale down
	ScaleDown(newReplicas uint64) (*CRStatus, error)

	// Update the service
	Update(spec *goharborv1alpha2.HarborCluster) (*CRStatus, error)

	// More...
}

type CRStatus struct {
	Condition  goharborv1alpha2.HarborClusterCondition `json:"condition"`
	Properties Properties                              `json:"properties"`
}

// New returns new CRStatus
func New(conditionType goharborv1alpha2.HarborClusterConditionType) *CRStatus {
	return &CRStatus{
		Condition: goharborv1alpha2.HarborClusterCondition{
			LastTransitionTime: metav1.Now(),
			Type:               conditionType,
		},
	}
}

// WithStatus returns CRStatus with Condition status
func (cs *CRStatus) WithStatus(status corev1.ConditionStatus) *CRStatus {
	cs.Condition.Status = status
	return cs
}

// WithReason returns CRStatus with Condition reason
func (cs *CRStatus) WithReason(reason string) *CRStatus {
	cs.Condition.Reason = reason
	return cs
}

// WithMessage returns CRStatus with Condition message
func (cs *CRStatus) WithMessage(message string) *CRStatus {
	cs.Condition.Message = message
	return cs
}

// WithProperties returns CRStatus with Properties
func (cs *CRStatus) WithProperties(properties Properties) *CRStatus {
	cs.Properties = properties
	return cs
}
