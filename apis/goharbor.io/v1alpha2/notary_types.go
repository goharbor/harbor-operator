package v1alpha2

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
)

const (
	NotaryMigrationSourceConfigKey = "notary-migration-source"
)

type NotaryLoggingSpec struct {
	// +kubebuilder:validation:Optional
	Level harbormetav1.NotaryLogLevel `json:"level,omitempty"`
}

type NotaryMigrationSpec struct {
	OpacifiedDSN `json:",inline"`
}

const migrationImage = "migrate/migrate"

var varFalse = false

func (r *NotaryMigrationSpec) GetMigrationContainer(ctx context.Context, storage *NotaryStorageSpec) (*corev1.Container, error) {
	migrationEnvs := []corev1.EnvVar{}
	secretDatabaseVariable := ""

	if storage.Postgres.PasswordRef != "" {
		migrationEnvs = append(migrationEnvs, corev1.EnvVar{
			Name: "secretDatabase",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: storage.Postgres.PasswordRef,
					},
					Key: harbormetav1.PostgresqlPasswordKey,
				},
			},
		})

		secretDatabaseVariable = "$(secretDatabase)"
	}

	migrationDatabaseURL := storage.Postgres.GetDSNStringWithRawPassword(secretDatabaseVariable)

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
					Key:      harbormetav1.SharedSecretKey,
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

type NotaryStorageSpec struct {
	// +kubebuilder:validation:Required
	Postgres harbormetav1.PostgresConnectionWithParameters `json:"postgres"`

	// TODO Add support for mysql and memory
}

func (n *NotaryStorageSpec) GetPasswordFieldKey() string {
	return harbormetav1.PostgresqlPasswordKey
}
