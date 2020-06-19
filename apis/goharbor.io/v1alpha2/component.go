package v1alpha2

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
)

type NodeSelector map[string]string

type ComponentSpec struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas"`

	// +kubebuilder:validation:Optional
	Priority *int32 `json:"priority,omitempty"`

	// +kubebuilder:validation:Optional
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// +kubebuilder:validation:Type=object
// ComponentStatus represents the current status of the resource.
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
type ComponentStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Status
	Status status.Status `json:"status"`

	// Current number of pods.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas"`

	// Message
	Message string `json:"message"`

	// Conditions list of extracted conditions from Resource
	// +listType:map
	// +listMapKey:type
	Conditions []Condition `json:"conditions"`
}

func (s ComponentStatus) MarshalJSON() ([]byte, error) {
	var data struct {
		ObservedGeneration int64         `json:"observedGeneration,omitempty"`
		Status             status.Status `json:"status"`
		Message            string        `json:"message"`
		Conditions         []Condition   `json:"conditions"`
	}

	data.ObservedGeneration = s.ObservedGeneration
	data.Status = s.Status
	data.Message = s.Message

	if s.Conditions == nil {
		data.Conditions = []Condition{}
	} else {
		data.Conditions = s.Conditions
	}

	return json.Marshal(data)
}

// Condition defines the general format for conditions on Kubernetes resources.
// In practice, each kubernetes resource defines their own format for conditions, but
// most (maybe all) follows this structure.
type Condition struct {
	// Type condition type
	Type status.ConditionType `json:"type"`

	// Status String that describes the condition status
	Status corev1.ConditionStatus `json:"status"`

	// Reason one work CamelCase reason
	Reason string `json:"reason"`

	// Message Human readable reason string
	Message string `json:"message"`
}
