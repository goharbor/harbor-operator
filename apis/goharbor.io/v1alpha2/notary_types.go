package v1alpha2

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	NotaryMigrationSourceConfigKey = "notary-migration-source"
)

type NotaryLoggingSpec struct {
	// +kubebuilder:validation:Optional
	Level NotaryLogLevel `json:"level,omitempty"`
}

type NotaryMigrationSpec struct {
	OpacifiedDSN `json:",inline"`
}

const migrationImage = "migrate/migrate"

var varFalse = false

func (r *NotaryMigrationSpec) GetMigrationContainer(ctx context.Context, storage *NotaryStorageSpec) (*corev1.Container, error) {
	migrationEnvs := []corev1.EnvVar{}
	secretDatabaseVariable := ""

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

	migrationDatabaseURL, err := storage.GetDSNStringWithRawPassword(secretDatabaseVariable)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get storage DSN")
	}

	migrationSourceURL, err := r.GetMigrationSourceURL(ctx)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			logger.Get(ctx).Info("no migration source url found, do not deploy migration container")

			return nil, nil
		}

		return nil, errors.Wrap(err, "cannot get migration source DSN")
	}

	if r != nil && r.PasswordRef != "" {
		migrationEnvs = append(migrationEnvs, corev1.EnvVar{
			Name: "secretSource",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.PasswordRef,
					},
					Key:      SharedSecretKey,
					Optional: &varFalse,
				},
			},
		})
	}

	return &corev1.Container{
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

func (r *NotaryMigrationSpec) GetMigrationSourceURL(ctx context.Context) (string, error) {
	if r == nil {
		return configstore.GetItemValue(NotaryMigrationSourceConfigKey)
	}

	secretSourceVariable := ""

	if r.PasswordRef != "" {
		secretSourceVariable = "$(secretSource)"
	}

	return r.GetDSNStringWithRawPassword(secretSourceVariable)
}

type NotaryHTTPSSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	CertificateRef string `json:"certificateRef"`
}

type NotaryStorageSpec struct {
	OpacifiedDSN `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"mysql","postgres","memory"}
	Type string `json:"type"`
}

var errNotImplemented = errors.New("not yet implemented")

func (n *NotaryStorageSpec) GetPasswordFieldKey() (string, error) {
	switch n.Type {
	case "postgres":
		return PostgresqlPasswordKey, nil
	default:
		return "", errNotImplemented
	}
}
