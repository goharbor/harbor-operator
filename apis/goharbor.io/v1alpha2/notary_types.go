package v1alpha2

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
)

type NotaryLoggingSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="info"
	Level harbormetav1.NotaryLogLevel `json:"level,omitempty"`
}

type NotaryMigrationSpec struct {
	// +kubebuilder:validation:Optional
	Github *NotaryMigrationGithubSpec `json:"github,omitempty"`

	// +kubebuilder:validation:Optional
	FileSystem *NotaryMigrationFileSystemSpec `json:"fileSystem,omitempty"`
}

func (r *NotaryMigrationSpec) DSNWithVariable(ctx context.Context) (string, error) {
	if r.Github != nil {
		return r.Github.DSNWithVariable(ctx)
	}

	if r.FileSystem != nil {
		return r.FileSystem.RelativeDSN(ctx)
	}

	return "", ErrNoMigrationConfiguration
}

func (r *NotaryMigrationSpec) Validate() error {
	if r == nil {
		return nil
	}

	found := 0

	if r.Github != nil {
		found++
	}

	if r.FileSystem != nil {
		found++
	}

	switch found {
	case 0:
		return ErrNoMigrationConfiguration
	case 1:
		return nil
	default:
		return Err2MigrationConfiguration
	}
}

func (r *NotaryMigrationSpec) Enabled() bool {
	return r != nil
}

// This driver is catered for those that want to source migrations from github.com.
// See https://github.com/golang-migrate/migrate/tree/master/source/github for more information.
type NotaryMigrationGithubSpec struct {
	// +kubebuilder:validation:Required
	CredentialsRef string `json:"credentialsRef"`

	// +kubebuilder:validation:Required
	Owner string `json:"owner"`

	// +kubebuilder:validation:Required
	RepositoryName string `json:"repositoryName"`

	// +kubebuilder:validation:Required
	Path string `json:"path"`

	// +kubebuilder:validation:Optional
	Reference string `json:"reference,omitempty"`
}

const (
	NotaryMigrationSourcePasswordVariableName = "passwordSource"
	NotaryMigrationSourceUserVariableName     = "userSource"
)

func (r *NotaryMigrationGithubSpec) DSNWithVariable(ctx context.Context) (string, error) {
	return fmt.Sprintf("github://$(%s):$(%s)@%s/%s/%s#%s", NotaryMigrationSourceUserVariableName, NotaryMigrationSourcePasswordVariableName, r.Owner, r.RepositoryName, r.Path, url.QueryEscape(r.Reference)), nil
}

// This driver is catered for those that want to source migrations from github.com.
// See https://github.com/golang-migrate/migrate/tree/master/source/github for more information.
type NotaryMigrationFileSystemSpec struct {
	// +kubebuilder:validation:Required
	VolumeSource corev1.VolumeSource `json:"volumeSource"`

	// +kubebuilder:validation:Optional
	SubPath string `json:"subPath,omitempty"`
}

const NotaryMigrationSourceVolumeMountVariableName = "mountPath"

func (r *NotaryMigrationFileSystemSpec) RelativeDSN(ctx context.Context) (string, error) {
	return fmt.Sprintf("file://$(%s)/%s", NotaryMigrationSourceVolumeMountVariableName, r.SubPath), nil
}

const migrationImage = "migrate/migrate"

var varFalse = false

func (r *NotaryMigrationSpec) GetMigrationContainer(ctx context.Context, storage *NotaryStorageSpec) (*corev1.Container, []corev1.Volume, error) {
	migrationEnvs := []corev1.EnvVar{}
	migrationVolumes := []corev1.Volume{}
	migrationVolumeMounts := []corev1.VolumeMount{}
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

	migrationSourceURL, err := r.DSNWithVariable(ctx)
	if err != nil {
		if err == ErrNoMigrationConfiguration {
			return nil, nil, serrors.UnrecoverrableError(err, serrors.InvalidSpecReason, "cannot get migration dsn")
		}

		return nil, nil, errors.Wrap(err, "cannot get migration source DSN")
	}

	if r.Github != nil {
		migrationEnvs = append(migrationEnvs, corev1.EnvVar{
			Name: NotaryMigrationSourcePasswordVariableName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.Github.CredentialsRef,
					},
					Key:      harbormetav1.GithubTokenPasswordKey,
					Optional: &varFalse,
				},
			},
		}, corev1.EnvVar{
			Name: NotaryMigrationSourceUserVariableName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.Github.CredentialsRef,
					},
					Key:      harbormetav1.GithubTokenUserKey,
					Optional: &varFalse,
				},
			},
		})
	}
	if r.FileSystem != nil {
		const mountPath = "/mnt/source"
		migrationEnvs = append(migrationEnvs, corev1.EnvVar{
			Name:  NotaryMigrationSourceVolumeMountVariableName,
			Value: mountPath,
		})
		migrationVolumes = append(migrationVolumes, corev1.Volume{
			Name:         "source",
			VolumeSource: r.FileSystem.VolumeSource,
		})
		migrationVolumeMounts = append(migrationVolumeMounts, corev1.VolumeMount{
			Name:      "source",
			MountPath: mountPath,
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
		Env:          migrationEnvs,
		VolumeMounts: migrationVolumeMounts,
	}, migrationVolumes, nil
}

type NotaryStorageSpec struct {
	// +kubebuilder:validation:Required
	Postgres harbormetav1.PostgresConnectionWithParameters `json:"postgres"`

	// TODO Add support for mysql and memory
}

func (n *NotaryStorageSpec) GetPasswordFieldKey() string {
	return harbormetav1.PostgresqlPasswordKey
}
