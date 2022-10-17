package notarysigner

import (
	"context"
	"path"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/image"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	VolumeName           = "config"
	ConfigPath           = "/etc/notary-signer"
	HTTPSVolumeName      = "certificates"
	HTTPSCertificatePath = ConfigPath + "/certificates"
)

var (
	varFalse = false

	fsGroup    int64 = 10000
	runAsGroup int64 = 10000
	runAsUser  int64 = 10000
)

func (r *Reconciler) GetDeployment(ctx context.Context, notary *goharborv1.NotarySigner) (*appsv1.Deployment, error) { //nolint:funlen
	getImageOptions := []image.Option{
		image.WithImageFromSpec(notary.Spec.Image),
		image.WithHarborVersion(version.GetVersion(notary.Annotations)),
	}

	image, err := image.GetImage(ctx, harbormetav1.NotarySignerComponent.String(), getImageOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, notary.GetName())
	namespace := notary.GetNamespace()

	volumes := []corev1.Volume{{
		Name: VolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
			},
		},
	}, {
		Name: HTTPSVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: notary.Spec.Authentication.CertificateRef,
			},
		},
	}}

	volumeMounts := []corev1.VolumeMount{{
		Name:      VolumeName,
		MountPath: ConfigPath,
	}, {
		Name:      HTTPSVolumeName,
		MountPath: HTTPSCertificatePath,
	}}

	initContainers := []corev1.Container{}

	migrateCmd := ""
	migrationEnvs := []corev1.EnvVar{}

	if notary.Spec.MigrationEnabled == nil || *notary.Spec.MigrationEnabled {
		secretDatabaseVariable := ""

		if notary.Spec.Storage.Postgres.PasswordRef != "" {
			migrationEnvs = append(migrationEnvs, corev1.EnvVar{
				Name: "secretDatabase",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: notary.Spec.Storage.Postgres.PasswordRef,
						},
						Key: harbormetav1.PostgresqlPasswordKey,
					},
				},
			})

			secretDatabaseVariable = "$(secretDatabase)"
		}

		migrationDatabaseURL := notary.Spec.Storage.Postgres.GetDSNStringWithRawPassword(secretDatabaseVariable)
		migrateCmd = "migrate-patch -database=" + migrationDatabaseURL + " && /migrations/migrate.sh && "

		migrationEnvs = append(migrationEnvs, corev1.EnvVar{
			Name:  "DB_URL",
			Value: migrationDatabaseURL,
		}, corev1.EnvVar{
			Name:  "MIGRATIONS_PATH",
			Value: "/migrations/signer/postgresql",
		}, corev1.EnvVar{
			Name:  "SERVICE_NAME",
			Value: "notary_signer",
		})
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: version.NewVersionAnnotations(notary.Annotations),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"):      name,
					r.Label("namespace"): namespace,
				},
			},
			Replicas: notary.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: notary.Spec.ComponentSpec.TemplateAnnotations,
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup:    &fsGroup,
						RunAsGroup: &runAsGroup,
						RunAsUser:  &runAsUser,
					},
					InitContainers: initContainers,
					Containers: []corev1.Container{{
						Name:         controllers.NotarySigner.String(),
						Image:        image,
						Command:      []string{"/bin/sh"},
						Args:         []string{"-c", migrateCmd + "notary-signer -config " + path.Join(ConfigPath, ConfigName)},
						VolumeMounts: volumeMounts,
						Ports: []corev1.ContainerPort{{
							ContainerPort: goharborv1.NotarySignerAPIPort,
							Name:          harbormetav1.NotarySignerAPIPortName,
							Protocol:      corev1.ProtocolTCP,
						}},
						EnvFrom: []corev1.EnvFromSource{{
							Prefix: "NOTARY_SIGNER_",
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: notary.Spec.Storage.AliasesRef,
								},
								Optional: &varFalse,
							},
						}},
						Env: migrationEnvs,
					}},
				},
			},
		},
	}

	notary.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, nil
}
