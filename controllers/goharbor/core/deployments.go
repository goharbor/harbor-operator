package core

import (
	"context"
	"fmt"
	"path"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/pkg/errors"
)

var (
	varFalse = false
)

const (
	port = 8080 // https://github.com/goharbor/harbor/blob/2fb1cc89d9ef9313842cc68b4b7c36be73681505/src/common/const.go#L127

	healthCheckPeriod                 = 90 * time.Second
	ConfigPath                        = "/etc/core"
	VolumeName                        = "configuration"
	InternalCertificatesVolumeName    = "internal-certificates"
	InternalCertificatesPath          = ConfigPath + "/internal-certificates"
	PublicCertificateVolumeName       = "ca-download"
	PublicCertificatePath             = ConfigPath + "/ca"
	CertificatesVolumeName            = "certificates"
	CertificatesPath                  = ConfigPath + "/certificates"
	EncryptionKeyVolumeName           = "encryption"
	EncryptionKeyPath                 = "key"
	HealthPath                        = "/api/v2.0/ping"
	TokenStorageVolumeName            = "psc"
	TokenStoragePath                  = ConfigPath + "/token"
	ServiceTokenCertificateVolumeName = "token-service-private-key"
	ServiceTokenCertificatePath       = ConfigPath + "private_key.pem"
)

func (r *Reconciler) GetDeployment(ctx context.Context, core *goharborv1alpha2.Core) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, core.GetName())
	namespace := core.GetNamespace()

	volumes := []corev1.Volume{
		{
			Name: VolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
					Optional: &varFalse,
				},
			},
		}, {
			Name: EncryptionKeyVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					Items: []corev1.KeyToPath{
						{
							Key:  goharborv1alpha2.SharedSecretKey,
							Path: EncryptionKeyPath,
						},
					},
					Optional:   &varFalse,
					SecretName: core.Spec.Database.EncryptionKeyRef,
				},
			},
		}, {
			Name: TokenStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}, {
			Name: ServiceTokenCertificateVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					Optional:   &varFalse,
					SecretName: core.Spec.ServiceToken.CertificateRef,
				},
			},
		},
	}

	volumesMount := []corev1.VolumeMount{
		{
			Name:      VolumeName,
			MountPath: path.Join(ConfigPath, ConfigName),
			SubPath:   ConfigName,
			ReadOnly:  true,
		}, {
			Name:      EncryptionKeyVolumeName,
			ReadOnly:  true,
			MountPath: path.Join(ConfigPath, EncryptionKeyPath),
			SubPath:   EncryptionKeyPath,
		}, {
			Name:      TokenStorageVolumeName,
			ReadOnly:  false,
			MountPath: TokenStoragePath,
		}, {
			Name:      ServiceTokenCertificateVolumeName,
			ReadOnly:  true,
			MountPath: ServiceTokenCertificatePath,
			SubPath:   corev1.TLSPrivateKeyKey,
		},
	}

	envs := []corev1.EnvVar{}

	if len(core.Spec.Components.Registry.Redis.DSN) > 0 {
		envs = append(envs, corev1.EnvVar{
			Name: "_REDIS_URL_REG",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key:      RegistryRedisDSNKey,
					Optional: &varFalse,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
				},
			},
		})
	}

	if core.Spec.PublicCertificateRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: PublicCertificateVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: core.Spec.PublicCertificateRef,
				},
			},
		})

		volumesMount = append(volumesMount, corev1.VolumeMount{
			Name:      PublicCertificateVolumeName,
			MountPath: PublicCertificatePath,
		})
	}

	if core.Spec.Components.Clair != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "CLAIR_DB_HOST",
			Value: core.Spec.Components.Clair.Database.Host,
		}, corev1.EnvVar{
			Name:  "CLAIR_DB_PORT",
			Value: fmt.Sprintf("%d", core.Spec.Components.Clair.Database.Port),
		}, corev1.EnvVar{
			Name:  "CLAIR_DB_USERNAME",
			Value: core.Spec.Components.Clair.Database.Username,
		}, corev1.EnvVar{
			Name:  "CLAIR_DB",
			Value: core.Spec.Components.Clair.Database.Name,
		}, corev1.EnvVar{
			Name:  "CLAIR_DB_SSLMODE",
			Value: core.Spec.Components.Clair.Database.SSLMode,
		}, corev1.EnvVar{
			Name: "CLAIR_DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key:      goharborv1alpha2.PostgresqlPasswordKey,
					Optional: &varFalse,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: core.Spec.Components.Clair.Database.PasswordRef,
					},
				},
			},
		}, corev1.EnvVar{
			Name:  "CLAIR_URL",
			Value: core.Spec.Components.Clair.URL,
		}, corev1.EnvVar{
			Name:  "CLAIR_ADAPTER_URL",
			Value: core.Spec.Components.Clair.AdapterURL,
		})
	}

	if core.Spec.Components.Trivy != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "TRIVY_ADAPTER_URL",
			Value: core.Spec.Components.Trivy.AdapterURL,
		})
	}

	if core.Spec.Components.NotaryServer != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "NOTARY_URL",
			Value: core.Spec.Components.NotaryServer.URL,
		})
	}

	if core.Spec.Components.NotaryServer != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "NOTARY_URL",
			Value: core.Spec.Components.NotaryServer.URL,
		})
	}

	if core.Spec.Components.TLS != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "INTERNAL_TLS_TRUST_CA_PATH",
			Value: path.Join(InternalCertificatesPath, corev1.ServiceAccountRootCAKey),
		}, corev1.EnvVar{
			Name:  "INTERNAL_TLS_CERT_PATH",
			Value: path.Join(InternalCertificatesPath, corev1.TLSCertKey),
		}, corev1.EnvVar{
			Name:  "INTERNAL_TLS_KEY_PATH",
			Value: path.Join(InternalCertificatesPath, corev1.TLSPrivateKeyKey),
		})

		volumesMount = append(volumesMount, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: path.Join(InternalCertificatesPath, corev1.ServiceAccountRootCAKey),
			SubPath:   corev1.ServiceAccountRootCAKey,
		}, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: path.Join(InternalCertificatesPath, corev1.TLSCertKey),
			SubPath:   corev1.TLSCertKey,
		}, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: path.Join(InternalCertificatesPath, corev1.TLSPrivateKeyKey),
			SubPath:   corev1.TLSPrivateKeyKey,
		})

		volumes = append(volumes, corev1.Volume{
			Name: InternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: core.Spec.Components.TLS.CertificateRef,
				},
			},
		})
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"):      name,
					r.Label("namespace"): namespace,
				},
			},
			Replicas: core.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 core.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					Containers: []corev1.Container{
						{
							Name:  "core",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
								},
							},

							// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/core/env.jinja
							Env: append(envs, corev1.EnvVar{
								Name:  "EXT_ENDPOINT",
								Value: core.Spec.ExternalEndpoint,
							}, corev1.EnvVar{
								Name:  "LOG_LEVEL",
								Value: string(core.Spec.Log.Level),
							}, corev1.EnvVar{
								Name:  "AUTH_MODE",
								Value: core.Spec.AuthenticationMode,
							}, corev1.EnvVar{
								Name:  "DATABASE_TYPE",
								Value: goharborv1alpha2.CoreDatabaseType,
							}, corev1.EnvVar{
								Name:  "POSTGRESQL_HOST",
								Value: core.Spec.Database.Host,
							}, corev1.EnvVar{
								Name:  "POSTGRESQL_PORT",
								Value: fmt.Sprintf("%d", core.Spec.Database.Port),
							}, corev1.EnvVar{
								Name:  "POSTGRESQL_USERNAME",
								Value: core.Spec.Database.Username,
							}, corev1.EnvVar{
								Name:  "POSTGRESQL_DATABASE",
								Value: core.Spec.Database.Name,
							}, corev1.EnvVar{
								Name: "POSTGRESQL_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key: goharborv1alpha2.PostgresqlPasswordKey,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.Database.PasswordRef,
										},
									},
								},
							}, corev1.EnvVar{
								Name:  "CORE_URL",
								Value: fmt.Sprintf("http://%s", core.GetName()),
							}, corev1.EnvVar{
								Name:  "CORE_LOCAL_URL",
								Value: fmt.Sprintf("http://%s", core.GetName()),
							}, corev1.EnvVar{
								Name: "CORE_SECRET",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key:      goharborv1alpha2.SharedSecretKey,
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.CoreConfig.SecretRef,
										},
									},
								},
							}, corev1.EnvVar{
								Name: "_REDIS_URL",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: name,
										},
										Key:      RedisDSNKey,
										Optional: &varFalse,
									},
								},
							}, corev1.EnvVar{
								Name:  "PORTAL_URL",
								Value: core.Spec.Components.Portal.URL,
							}, corev1.EnvVar{
								Name:  "REGISTRY_CONTROLLER_URL",
								Value: core.Spec.Components.Registry.ControllerURL,
							}, corev1.EnvVar{
								Name:  "REGISTRY_CREDENTIAL_USERNAME",
								Value: core.Spec.Components.Registry.Credentials.Username,
							}, corev1.EnvVar{
								Name: "REGISTRY_CREDENTIAL_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.Components.Registry.Credentials.PasswordRef,
										},
										Key:      goharborv1alpha2.SharedSecretKey,
										Optional: &varFalse,
									},
								},
							}, corev1.EnvVar{
								Name:  "JOBSERVICE_URL",
								Value: core.Spec.Components.JobService.URL,
							}, corev1.EnvVar{
								Name: "JOBSERVICE_SECRET",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key:      goharborv1alpha2.SharedSecretKey,
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.Components.JobService.SecretRef,
										},
									},
								},
							}, corev1.EnvVar{
								Name: "CSRF_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key:      goharborv1alpha2.CSRFSecretKey,
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.CSRFKeyRef,
										},
									},
								},
							}, corev1.EnvVar{
								Name:  "INTERNAL_TLS_KEY_PATH",
								Value: path.Join(CertificatesPath, corev1.TLSPrivateKeyKey),
							}, corev1.EnvVar{
								Name:  "INTERNAL_TLS_CERT_PATH",
								Value: path.Join(CertificatesPath, corev1.TLSCertKey),
							}, corev1.EnvVar{
								Name:  "INTERNAL_TLS_TRUST_CA_PATH",
								Value: path.Join(CertificatesPath, corev1.ServiceAccountRootCAKey),
							}, corev1.EnvVar{
								Name:  "REGISTRY_URL",
								Value: core.Spec.Components.Registry.URL,
							}, corev1.EnvVar{
								Name:  "REGISTRYCTL_URL",
								Value: core.Spec.Components.Registry.ControllerURL,
							}, corev1.EnvVar{
								Name:  "TOKEN_SERVICE_URL",
								Value: core.Spec.Components.TokenService.URL,
							}, corev1.EnvVar{
								Name:  "CONFIG_PATH",
								Value: path.Join(ConfigPath, ConfigName),
							}, corev1.EnvVar{
								Name:  "CFG_EXPIRATION",
								Value: fmt.Sprintf("%.0f", core.Spec.ConfigExpiration.Duration.Seconds()),
							}, corev1.EnvVar{
								Name:  "HTTP_PROXY",
								Value: "", // TODO
							}, corev1.EnvVar{
								Name:  "HTTPS_PROXY",
								Value: "", // TODO
							}, corev1.EnvVar{
								Name:  "RELOAD_KEY",
								Value: "true",
							}, corev1.EnvVar{
								Name:  "SYNC_QUOTA",
								Value: "true", // TODO
							}, corev1.EnvVar{
								Name:  "SYNC_REGISTRY",
								Value: fmt.Sprintf("%+v", core.Spec.Components.Registry.Sync),
							}, corev1.EnvVar{
								Name: "HARBOR_ADMIN_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key:      goharborv1alpha2.SharedSecretKey,
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.AdminInitialPasswordRef,
										},
									},
								},
							}, corev1.EnvVar{
								Name:  "INTERNAL_TLS_ENABLED",
								Value: fmt.Sprintf("%v", core.Spec.Components.TLS != nil),
							}, corev1.EnvVar{
								Name:  "WITH_CHARTMUSEUM",
								Value: fmt.Sprintf("%+v", core.Spec.Components.ChartRepository != nil),
							}, corev1.EnvVar{
								Name:  "WITH_CLAIR",
								Value: fmt.Sprintf("%+v", core.Spec.Components.Clair != nil),
							}, corev1.EnvVar{
								Name:  "WITH_NOTARY",
								Value: fmt.Sprintf("%+v", core.Spec.Components.NotaryServer != nil),
							}, corev1.EnvVar{
								Name:  "WITH_TRIVY",
								Value: fmt.Sprintf("%+v", core.Spec.Components.Trivy != nil),
							}),
							ImagePullPolicy: corev1.PullAlways,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: HealthPath,
										Port: intstr.FromInt(port),
									},
								},
								PeriodSeconds: int32(healthCheckPeriod.Seconds()),
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: HealthPath,
										Port: intstr.FromInt(port),
									},
								},
							},
							VolumeMounts: volumesMount,
						},
					},
				},
			},
		},
	}, nil
}
