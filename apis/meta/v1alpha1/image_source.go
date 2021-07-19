package v1alpha1

import (
	"github.com/goharbor/harbor-operator/pkg/image"
	corev1 "k8s.io/api/core/v1"
)

type ImageSourceSpec struct {
	// +kubebuilder:validation:Required
	// The default repository for the images of the components. eg docker.io/goharbor/
	Repository string `json:"repository,omitempty"`

	// +kubebuilder:validation:Optional
	// The tag suffix for the images of the images of the components. eg '-patch1'
	TagSuffix string `json:"tagSuffix,omitempty"`

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

func (spec *ImageSourceSpec) AddRepositoryAndTagSuffixOptions(options ...image.Option) []image.Option {
	if spec == nil {
		return options
	}

	if spec.Repository != "" || spec.TagSuffix != "" {
		options = append(options,
			image.WithRepository(spec.Repository),
			image.WithTagSuffix(spec.TagSuffix),
		)
	}

	return options
}
