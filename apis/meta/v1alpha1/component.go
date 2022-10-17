package v1alpha1

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/goharbor/harbor-operator/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
)

//go:generate stringer -type=Component -linecomment
type Component int

const (
	CoreComponent               Component = iota // core
	JobServiceComponent                          // jobservice
	PortalComponent                              // portal
	RegistryComponent                            // registry
	RegistryControllerComponent                  // registryctl
	ChartMuseumComponent                         // chartmuseum
	ExporterComponent                            // exporter
	NotaryServerComponent                        // notaryserver
	NotarySignerComponent                        // notarysigner
	TrivyComponent                               // trivy

	componentCount
)

var ErrUnsupportedComponent = errors.New("component not supported")

func GetLargestComponentNameSize() int {
	max := len(Component(math.MaxInt64).String())

	size := len(Component(math.MinInt64).String())
	if size > max {
		max = size
	}

	for i := 0; i < int(componentCount); i++ {
		size := len(Component(i).String())
		if size > max {
			max = size
		}
	}

	return max
}

type ImageSpec struct {
	// +kubebuilder:validation:Optional
	// Image name for the component.
	Image string `json:"image,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"Always","Never","IfNotPresent"}
	// Image pull policy.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// +kubebuilder:validation:Optional
	// +listType:map
	// +listMapKey:name
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

type ComponentSpec struct {
	// +kubebuilder:validation:Optional
	// Custom annotations to be added into the pods
	TemplateAnnotations map[string]string `json:"templateAnnotations,omitempty"`

	ImageSpec `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#what-is-a-replicationcontroller
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	// ServiceAccountName is the name of the ServiceAccount to use to run this component.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// +kubebuilder:validation:Optional
	// NodeSelector is a selector which must be true for the component to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// If specified, the pod's tolerations.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// +kubebuilder:validation:Optional
	// Compute Resources required by this component.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

func (c *ComponentSpec) ApplyToDeployment(deploy *appsv1.Deployment) {
	deploy.Spec.Replicas = c.Replicas
	deploy.Spec.Template.Spec.ServiceAccountName = c.ServiceAccountName

	for i := range deploy.Spec.Template.Spec.Containers {
		if c.ImagePullPolicy != nil {
			deploy.Spec.Template.Spec.Containers[i].ImagePullPolicy = *c.ImagePullPolicy
		} else {
			deploy.Spec.Template.Spec.Containers[i].ImagePullPolicy = config.DefaultImagePullPolicy
		}

		deploy.Spec.Template.Spec.Containers[i].Resources = c.Resources
	}

	deploy.Spec.Template.Spec.ImagePullSecrets = c.ImagePullSecrets
	deploy.Spec.Template.Spec.NodeSelector = c.NodeSelector
	deploy.Spec.Template.Spec.Tolerations = c.Tolerations
}

// ComponentStatus represents the current status of the resource.
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
type ComponentStatus struct {
	// +kubebuilder:validation:Optional
	Operator OperatorStatus `json:"operator,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current number of pods.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas,omitempty"`

	// Conditions list of extracted conditions from Resource
	// +listType:map
	// +listMapKey:type
	Conditions []Condition `json:"conditions"`
}

func (s ComponentStatus) MarshalJSON() ([]byte, error) {
	var data struct {
		ObservedGeneration int64          `json:"observedGeneration,omitempty"`
		Operator           OperatorStatus `json:"operator,omitempty"`
		Replicas           *int32         `json:"replicas,omitempty"`
		Conditions         []Condition    `json:"conditions"`
	}

	data.Operator = s.Operator
	data.Replicas = s.Replicas
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
