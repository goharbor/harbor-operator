package v1alpha2

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
)

type ComponentSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#what-is-a-replicationcontroller
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// +kubebuilder:validation:Optional
	// ServiceAccountName is the name of the ServiceAccount to use to run this component.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,8,opt,name=serviceAccountName"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"Always","Never","IfNotPresent"}
	// +kubebuilder:default="IfNotPresent"
	// Image pull policy.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`

	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +kubebuilder:validation:Optional
	// NodeSelector is a selector which must be true for the component to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// +kubebuilder:validation:Optional
	// If specified, the pod's tolerations.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`

	// +kubebuilder:validation:Optional
	// Compute Resources required by this component.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// +kubebuilder:validation:Type=object
// ComponentStatus represents the current status of the resource.
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
type ComponentStatus struct {
	// +kubebuilder:validation:Optional
	Operator OperatorStatus `json:"operator,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current number of pods.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas"`

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
	// +kubebuilder:validation:Required
	// Type condition type
	Type status.ConditionType `json:"type"`

	// +kubebuilder:validation:Required
	// Status String that describes the condition status
	Status corev1.ConditionStatus `json:"status"`

	// +kubebuilder:validation:Optional
	// Reason one work CamelCase reason
	Reason string `json:"reason,omitempty"`

	// +kubebuilder:validation:Optional
	// Message Human readable reason string
	Message string `json:"message,omitempty"`
}
