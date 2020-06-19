package v1alpha2

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type NotaryLoggingSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal","panic"}
	// +kubebuilder:default=info
	Level string `json:"level"`
}

type NotaryMigrationSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	// Reference of the source of the migration.
	SourceRef *corev1.SecretKeySelector `json:"sourceRef"`

	// +kubebuilder:validation:Optional
	// Source of the migration.
	// Cannot be used if SourceRef is not empty.
	Source string `json:"source"`

	// +kubebuilder:validation:Optional
	// +nullable
	Volume corev1.VolumeSource `json:"volume"`

	// +kubebuilder:validation:Optional
	// +nullable
	Mount corev1.VolumeMount `json:"volumeMount"`
}

func (r *NotaryMigrationSpec) ValidateCreate() error {
	return r.Validate()
}

func (r *NotaryMigrationSpec) ValidateUpdate(old runtime.Object) error {
	return r.Validate()
}

func (r *NotaryMigrationSpec) Validate() error {
	if !r.Enabled {
		return nil
	}

	err := r.ValidateSource()
	if err != nil {
		return err
	}

	return r.ValidateVolume()
}

var errTwoSource = errors.New("only one source and sourceRef can be specified")

func (r *NotaryMigrationSpec) ValidateSource() error {
	if r.Source != "" && r.SourceRef != nil {
		return errTwoSource
	}

	return nil
}

var errVolume = errors.New("both or neither of volume and volumeMount should be specified")

func (r *NotaryMigrationSpec) ValidateVolume() error {
	if r.Volume.Size() == 0 && r.Mount.Size() == 0 {
		return nil
	}

	if r.Volume.Size() == 0 || r.Mount.Size() == 0 {
		return errVolume
	}

	return nil
}
