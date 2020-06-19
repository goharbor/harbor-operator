package notaryserver

import (
	"context"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/pkg/errors"
)

const (
	notaryServerPort     = 4443
	VolumeName           = "config"
	ConfigPath           = "/etc/notary-server"
	HTTPSVolumeName      = "certificates"
	HTTPSCertificatePath = ConfigPath + "/certificates"
	TrustVolumeName      = "trust-certificates"
	TrustCertificatePath = ConfigPath + "/trust-certificates"
	AuthVolumeName       = "auth-certificates"
	AuthCertificatePath  = ConfigPath + "/auth-certificates"
)

var (
	varFalse = false
)

func (r *Reconciler) GetDeployment(ctx context.Context, notary *goharborv1alpha2.NotaryServer) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	dbMigratorImage, err := r.GetDBMigratorImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get dbMigrator image")
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
	}}

	volumeMounts := []corev1.VolumeMount{{
		Name:      VolumeName,
		MountPath: ConfigPath,
	}}

	if notary.Spec.TrustService.CertificateRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: TrustVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: notary.Spec.TrustService.CertificateRef,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      TrustVolumeName,
			MountPath: TrustCertificatePath,
		})
	}

	if notary.Spec.HTTPS.CertificateRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: HTTPSVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: notary.Spec.HTTPS.CertificateRef,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      HTTPSVolumeName,
			MountPath: HTTPSCertificatePath,
		})
	}

	if notary.Spec.Auth.Token.CertificateRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: AuthVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: notary.Spec.Auth.Token.CertificateRef,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      AuthVolumeName,
			MountPath: AuthCertificatePath,
		})
	}

	migrationDatabaseURL, err := notary.Spec.Storage.GetDSNStringWithRawPassword("$(secret)")
	if err != nil {
		return nil, errors.Wrap(err, "cannot get storage DSN")
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"notaryserver.goharbor.io/name":      name,
					"notaryserver.goharbor.io/namespace": namespace,
				},
			},
			Replicas: notary.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"notaryserver.goharbor.io/name":      name,
						"notaryserver.goharbor.io/namespace": namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 notary.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					InitContainers: []corev1.Container{
						{
							Name:  "init-db",
							Image: dbMigratorImage,
							Args: []string{
								"-source", "$(source)",
								"-database", migrationDatabaseURL,
								"up",
							},
							Env: []corev1.EnvVar{
								{
									Name: "secret",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: notary.Spec.Storage.PasswordRef,
											},
											Key: goharborv1alpha2.PostgresqlPasswordKey,
										},
									},
								}, {
									Name: "source",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: name,
											},
											Key:      MigrationSourceSecretKey,
											Optional: &varFalse,
										},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "notary-server",
							Image:           image,
							Args:            []string{"notary-server", "-config", path.Join(ConfigPath, ConfigName)},
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts:    volumeMounts,
						},
					},
					Priority: notary.Spec.Priority,
				},
			},
		},
	}, nil
}
