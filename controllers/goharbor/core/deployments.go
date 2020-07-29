package core

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/config/harbor"
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

	database, dbName, err := goharborv1alpha2.FromOpacifiedDSN(core.Spec.Database.OpacifiedDSN)
	if err != nil {
		return nil, errors.Wrap(err, "database")
	}

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
					SecretName: core.Spec.Components.TokenService.CertificateRef,
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
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

	envs, err := harbor.EnvVars(map[string]harbor.ConfigValue{
		common.ExtEndpoint:        harbor.Value(core.Spec.ExternalEndpoint),
		common.AUTHMode:           harbor.Value(core.Spec.AuthenticationMode),
		common.DatabaseType:       harbor.Value(goharborv1alpha2.CoreDatabaseType),
		common.PostGreSQLHOST:     harbor.Value(database.Host),
		common.PostGreSQLPort:     harbor.Value(fmt.Sprintf("%d", database.Port)),
		common.PostGreSQLUsername: harbor.Value(database.Username),
		common.PostGreSQLPassword: harbor.ValueFrom(corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: goharborv1alpha2.PostgresqlPasswordKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: database.PasswordRef,
				},
			},
		}),
		common.PostGreSQLDatabase:    harbor.Value(dbName),
		common.CoreURL:               harbor.Value(core.Spec.ExternalEndpoint),
		common.CoreLocalURL:          harbor.Value(fmt.Sprintf("http://%s", core.GetName())),
		common.RegistryControllerURL: harbor.Value(core.Spec.Components.Registry.ControllerURL),
		common.RegistryURL:           harbor.Value(core.Spec.Components.Registry.URL),
		common.JobServiceURL:         harbor.Value(core.Spec.Components.JobService.URL),
		common.TokenServiceURL:       harbor.Value(core.Spec.Components.TokenService.URL),
		common.AdminInitialPassword: harbor.ValueFrom(corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:      goharborv1alpha2.SharedSecretKey,
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: core.Spec.AdminInitialPasswordRef,
				},
			},
		}),
		common.WithChartMuseum: harbor.Value(fmt.Sprintf("%+v", core.Spec.Components.ChartRepository != nil)),
		common.WithClair:       harbor.Value(fmt.Sprintf("%+v", core.Spec.Components.Clair != nil)),
		common.WithNotary:      harbor.Value(fmt.Sprintf("%+v", core.Spec.Components.NotaryServer != nil)),
		common.WithTrivy:       harbor.Value(fmt.Sprintf("%+v", core.Spec.Components.Trivy != nil)),
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot configure environment variables")
	}

	envs = append(envs, []corev1.EnvVar{{
		Name:  "LOG_LEVEL",
		Value: string(core.Spec.Log.Level),
	}, {
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
	}, {
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
	}, {
		Name:  "PORTAL_URL",
		Value: core.Spec.Components.Portal.URL,
	}, {
		Name:  "REGISTRY_CREDENTIAL_USERNAME",
		Value: core.Spec.Components.Registry.Credentials.Username,
	}, {
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
	}, {
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
	}, {
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
	}, {
		Name:  "CONFIG_PATH",
		Value: path.Join(ConfigPath, ConfigName),
	}, {
		Name:  "CFG_EXPIRATION",
		Value: fmt.Sprintf("%.0f", core.Spec.ConfigExpiration.Duration.Seconds()),
	}, {
		Name:  "HTTP_PROXY",
		Value: "", // TODO
	}, {
		Name:  "HTTPS_PROXY",
		Value: "", // TODO
	}, {
		Name:  "RELOAD_KEY",
		Value: "true",
	}, {
		Name:  "SYNC_QUOTA",
		Value: "true", // TODO
	}, {
		Name:  "SYNC_REGISTRY",
		Value: fmt.Sprintf("%+v", core.Spec.Components.Registry.Sync),
	}, {
		Name:  "INTERNAL_TLS_ENABLED",
		Value: fmt.Sprintf("%v", core.Spec.Components.TLS != nil),
	}}...)

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

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      PublicCertificateVolumeName,
			MountPath: PublicCertificatePath,
		})
	}

	if core.Spec.Components.ChartRepository != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "CHART_REPOSITORY_URL",
			Value: core.Spec.Components.ChartRepository.URL,
		}, corev1.EnvVar{
			Name:  "CHART_CACHE_DRIVER",
			Value: core.Spec.Components.ChartRepository.CacheDriver,
		})
	}

	if core.Spec.Components.Clair != nil {
		database, dbName, err := goharborv1alpha2.FromOpacifiedDSN(core.Spec.Database.OpacifiedDSN)
		if err != nil {
			return nil, errors.Wrap(err, "database")
		}

		adapterURLConfig, err := harbor.EnvVar(common.ClairAdapterURL, harbor.Value(core.Spec.Components.Clair.AdapterURL))
		if err != nil {
			return nil, errors.Wrap(err, "cannot configure clair")
		}

		envs = append(envs, adapterURLConfig, corev1.EnvVar{
			Name:  "CLAIR_DB_HOST",
			Value: database.Host,
		}, corev1.EnvVar{
			Name:  "CLAIR_DB_PORT",
			Value: fmt.Sprintf("%d", database.Port),
		}, corev1.EnvVar{
			Name:  "CLAIR_DB_USERNAME",
			Value: database.Username,
		}, corev1.EnvVar{
			Name:  "CLAIR_DB",
			Value: dbName,
		}, corev1.EnvVar{
			Name:  "CLAIR_DB_SSLMODE",
			Value: database.SSLMode,
		}, corev1.EnvVar{
			Name: "CLAIR_DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key:      goharborv1alpha2.PostgresqlPasswordKey,
					Optional: &varFalse,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: database.PasswordRef,
					},
				},
			},
		}, corev1.EnvVar{
			Name:  "CLAIR_URL",
			Value: core.Spec.Components.Clair.URL,
		})
	}

	if core.Spec.Components.Trivy != nil {
		adapterURLConfig, err := harbor.EnvVar(common.TrivyAdapterURL, harbor.Value(core.Spec.Components.Trivy.AdapterURL))
		if err != nil {
			return nil, errors.Wrap(err, "cannot configure trivy")
		}

		envs = append(envs, adapterURLConfig)
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

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
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
							Env:             envs,
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
