package v1alpha2

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type NotaryLoggingSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal","panic"}
	// +kubebuilder:default="info"
	Level string `json:"level"`
}

type NotaryMigrationSpec struct {
	// +kubebuilder:validation:Optional
	Disabled bool `json:"disabled"`

	// +kubebuilder:validation:Optional
	// Source of the migration.
	Source OpacifiedDSN `json:"source"`
}

func (r *NotaryMigrationSpec) Validate() error {
	if r.Disabled {
		return nil
	}

	return nil
}

const migrationImage = "migrate/migrate"

var varFalse = false

func (r *NotaryMigrationSpec) GetMigrationContainer(ctx context.Context, storage *NotaryStorageSpec) (corev1.Container, error) {
	migrationEnvs := []corev1.EnvVar{}
	secretDatabaseVariable, secretSourceVariable := "", ""

	if storage.PasswordRef != "" {
		migrationEnvs = append(migrationEnvs, corev1.EnvVar{
			Name: "secretDatabase",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: storage.PasswordRef,
					},
					Key: PostgresqlPasswordKey,
				},
			},
		})

		secretDatabaseVariable = "$(secretDatabase)"
	}

	if r.Source.PasswordRef != "" {
		migrationEnvs = append(migrationEnvs, corev1.EnvVar{
			Name: "secretSource",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.Source.PasswordRef,
					},
					Key:      SharedSecretKey,
					Optional: &varFalse,
				},
			},
		})

		secretSourceVariable = "$(secretSource)"
	}

	migrationDatabaseURL, err := storage.GetDSNStringWithRawPassword(secretDatabaseVariable)
	if err != nil {
		return corev1.Container{}, errors.Wrap(err, "cannot get storage DSN")
	}

	migrationSourceURL, err := r.Source.GetDSNStringWithRawPassword(secretSourceVariable)
	if err != nil {
		return corev1.Container{}, errors.Wrap(err, "cannot get migration source DSN")
	}

	return corev1.Container{
		Name:  "init-db",
		Image: migrationImage,
		Args: []string{
			"-source", migrationSourceURL,
			"-database", migrationDatabaseURL,
			"up",
		},
		Env: migrationEnvs,
	}, nil
}

type NotaryHTTPSSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	CertificateRef string `json:"certificateRef"`
}

type NotaryStorageSpec struct {
	OpacifiedDSN `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"mysql","postgres","memory"}
	Type string `json:"type"`
}

var (
	errNotImplemented = errors.New("not yet implemented")
)

func (n *NotaryStorageSpec) GetPasswordFieldKey() (string, error) {
	switch n.Type {
	case "postgres":
		return PostgresqlPasswordKey, nil
	default:
		return "", errNotImplemented
	}
}
