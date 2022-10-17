package core

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config/harbor"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/image"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/goharbor/harbor/src/common"
	registry "github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	varFalse = false

	fsGroup    int64 = 10000
	runAsGroup int64 = 10000
	runAsUser  int64 = 10000

	terminationGracePeriodSeconds int64 = 120
)

const (
	healthCheckPeriod                     = 90 * time.Second
	ConfigPath                            = "/etc/core"
	VolumeName                            = "configuration"
	InternalCertificatesVolumeName        = "internal-certificates"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
	InternalCertificatesPath              = ConfigPath + "/ssl"
	PublicCertificateVolumeName           = "ca-download"
	PublicCertificatePath                 = ConfigPath + "/ca"
	EncryptionKeyVolumeName               = "encryption"
	EncryptionKeyPath                     = "key"
	HealthPath                            = "/api/v2.0/ping"
	TokenStorageVolumeName                = "psc"
	TokenStoragePath                      = ConfigPath + "/token"
	ServiceTokenCertificateVolumeName     = "token-service-private-key"
	ServiceTokenCertificatePath           = ConfigPath + "/private_key.pem"
)

const (
	httpsPort = 8443 // https://github.com/goharbor/harbor/blob/46d7434d0b0e647d4638e69693d4eddf50841ccb/src/core/main.go#L215
	httpPort  = 8080 // https://github.com/goharbor/harbor/blob/2fb1cc89d9ef9313842cc68b4b7c36be73681505/src/common/const.go#L127
)

func getDefaultAllowedRegistryTypesForProxyCache() string {
	// TODO: only enable docker registry in harbor 2.3.x
	return strings.Join([]string{
		registry.RegistryTypeDockerHub,
		registry.RegistryTypeHarbor,
		registry.RegistryTypeAzureAcr,
		registry.RegistryTypeAwsEcr,
		registry.RegistryTypeGoogleGcr,
		registry.RegistryTypeQuay,
		registry.RegistryTypeDockerRegistry,
	}, ",")
}

func (r *Reconciler) GetDeployment(ctx context.Context, core *goharborv1.Core) (*appsv1.Deployment, error) { //nolint:funlen
	getImageOptions := []image.Option{
		image.WithImageFromSpec(core.Spec.Image),
		image.WithHarborVersion(version.GetVersion(core.Annotations)),
	}

	image, err := image.GetImage(ctx, harbormetav1.CoreComponent.String(), getImageOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, core.GetName())
	namespace := core.GetNamespace()

	volumes := []corev1.Volume{{
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
				Items: []corev1.KeyToPath{{
					Key:  harbormetav1.SharedSecretKey,
					Path: EncryptionKeyPath,
				}},
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
	}}

	volumeMounts := []corev1.VolumeMount{{
		Name:      VolumeName,
		MountPath: path.Join(ConfigPath, ConfigName),
		SubPath:   strings.TrimLeft(ConfigName, "/"),
		ReadOnly:  true,
	}, {
		Name:      EncryptionKeyVolumeName,
		ReadOnly:  true,
		MountPath: path.Join(ConfigPath, EncryptionKeyPath),
		SubPath:   strings.TrimLeft(EncryptionKeyPath, "/"),
	}, {
		Name:      TokenStorageVolumeName,
		ReadOnly:  false,
		MountPath: TokenStoragePath,
	}, {
		Name:      ServiceTokenCertificateVolumeName,
		ReadOnly:  true,
		MountPath: ServiceTokenCertificatePath,
		SubPath:   strings.TrimLeft(corev1.TLSPrivateKeyKey, "/"),
	}}

	// inject certs if need.
	if core.Spec.CertificateInjection.ShouldInject() {
		volumes = append(volumes, core.Spec.CertificateInjection.GenerateVolumes()...)
		volumeMounts = append(volumeMounts, core.Spec.CertificateInjection.GenerateVolumeMounts()...)
	}

	scheme := "http"
	if core.Spec.Components.TLS.Enabled() {
		scheme = "https"
	}

	host := r.NormalizeName(ctx, core.GetName())
	if core.Spec.Components.TLS.Enabled() {
		host += ":443"
	} else {
		host += ":80"
	}

	coreLocalURL := (&url.URL{
		Scheme: scheme,
		Host:   host,
	}).String()

	// Only one host is supported
	if len(core.Spec.Database.Hosts) == 0 {
		return nil, serrors.UnrecoverrableError(harbormetav1.NewErrPostgresNoHost(), serrors.InvalidSpecReason, "cannot get a database host")
	}

	firstDatabaseHost := core.Spec.Database.Hosts[0]

	envs, err := harbor.EnvVars(map[string]harbor.ConfigValue{
		common.ExtEndpoint:        harbor.Value(core.Spec.ExternalEndpoint),
		common.AUTHMode:           harbor.Value(core.Spec.AuthenticationMode),
		common.DatabaseType:       harbor.Value(goharborv1.CoreDatabaseType),
		common.PostGreSQLHOST:     harbor.Value(firstDatabaseHost.Host),
		common.PostGreSQLPort:     harbor.Value(fmt.Sprintf("%d", firstDatabaseHost.Port)),
		common.PostGreSQLUsername: harbor.Value(core.Spec.Database.Username),
		common.PostGreSQLPassword: harbor.ValueFrom(corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: harbormetav1.PostgresqlPasswordKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: core.Spec.Database.PasswordRef,
				},
			},
		}),
		common.PostGreSQLDatabase:    harbor.Value(core.Spec.Database.Database),
		common.CoreURL:               harbor.Value(coreLocalURL),
		common.CoreLocalURL:          harbor.Value(coreLocalURL),
		common.RegistryControllerURL: harbor.Value(core.Spec.Components.Registry.ControllerURL),
		common.RegistryURL:           harbor.Value(core.Spec.Components.Registry.RegistryURL),
		common.JobServiceURL:         harbor.Value(core.Spec.Components.JobService.URL),
		common.TokenServiceURL:       harbor.Value(core.Spec.Components.TokenService.URL),
		common.AdminInitialPassword: harbor.ValueFrom(corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:      harbormetav1.SharedSecretKey,
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: core.Spec.AdminInitialPasswordRef,
				},
			},
		}),
		common.WithChartMuseum: harbor.Value(strconv.FormatBool(core.Spec.Components.ChartRepository != nil)),
		common.WithNotary:      harbor.Value(strconv.FormatBool(core.Spec.Components.NotaryServer != nil)),
		common.WithTrivy:       harbor.Value(strconv.FormatBool(core.Spec.Components.Trivy != nil)),
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot configure environment variables")
	}

	metricsEnvs, err := core.Spec.Metrics.GetEnvVars(harbormetav1.CoreComponent.String())
	if err != nil {
		return nil, errors.Wrap(err, "get metrics environment variables")
	}

	envs = append(envs, metricsEnvs...)

	envs, err = core.Spec.Trace.AddEnvVars(harbormetav1.CoreComponent.String(), envs)
	if err != nil {
		return nil, errors.Wrap(err, "get trace environment variables")
	}

	envs = append(envs, core.Spec.Proxy.GetEnvVars()...)

	envs = append(envs, []corev1.EnvVar{{
		Name:  "LOG_LEVEL",
		Value: string(core.Spec.Log.Level),
	}, {
		Name: "CORE_SECRET",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:      harbormetav1.SharedSecretKey,
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: core.Spec.CoreConfig.SecretRef,
				},
			},
		},
	}, {
		Name: RedisDSNKey,
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
				Key:      harbormetav1.SharedSecretKey,
				Optional: &varFalse,
			},
		},
	}, {
		Name: "JOBSERVICE_SECRET",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:      harbormetav1.SharedSecretKey,
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
				Key:      harbormetav1.CSRFSecretKey,
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
		Name:  "RELOAD_KEY",
		Value: "true",
	}, {
		Name:  "SYNC_QUOTA",
		Value: "true", // TODO
	}, {
		Name:  "SYNC_REGISTRY",
		Value: strconv.FormatBool(core.Spec.Components.Registry.Sync),
	}, {
		Name:  "INTERNAL_TLS_ENABLED",
		Value: strconv.FormatBool(core.Spec.Components.TLS.Enabled()),
	}, {
		Name:  "PERMITTED_REGISTRY_TYPES_FOR_PROXY_CACHE",
		Value: getDefaultAllowedRegistryTypesForProxyCache(),
	}}...)

	envs, err = addDatabaseEnvs(core, envs)
	if err != nil {
		return nil, err
	}

	if core.Spec.Components.Registry.Redis != nil {
		envs = append(envs, corev1.EnvVar{
			Name: RegistryRedisDSNKey,
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
		urlConfig, err := harbor.EnvVar(common.ChartRepoURL, harbor.Value(core.Spec.Components.ChartRepository.URL))
		if err != nil {
			return nil, errors.Wrap(err, "cannot configure chartmuseum")
		}

		envs = append(envs, urlConfig, corev1.EnvVar{
			Name:  "CHART_CACHE_DRIVER",
			Value: core.Spec.Components.ChartRepository.CacheDriver,
		})
	}

	if core.Spec.Components.Trivy != nil {
		adapterURL, err := harbor.EnvVar(common.TrivyAdapterURL, harbor.Value(core.Spec.Components.Trivy.AdapterURL))
		if err != nil {
			return nil, errors.Wrap(err, "cannot configure trivy")
		}

		envs = append(envs, adapterURL)
	}

	if core.Spec.Components.NotaryServer != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "NOTARY_URL",
			Value: core.Spec.Components.NotaryServer.URL,
		})
	}

	if core.Spec.Components.TLS.Enabled() {
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
			MountPath: path.Join(InternalCertificateAuthorityDirectory, corev1.ServiceAccountRootCAKey),
			SubPath:   strings.TrimLeft(corev1.ServiceAccountRootCAKey, "/"),
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: InternalCertificatesPath,
			ReadOnly:  true,
		})

		volumes = append(volumes, corev1.Volume{
			Name: InternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: core.Spec.Components.TLS.CertificateRef,
				},
			},
		})
	} else {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: InternalCertificateAuthorityDirectory,
		})

		volumes = append(volumes, corev1.Volume{
			Name: InternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	port := httpPort
	portName := harbormetav1.CoreHTTPPortName

	if core.Spec.Components.TLS.Enabled() {
		port = httpsPort
		portName = harbormetav1.CoreHTTPSPortName
	}

	httpGET := &corev1.HTTPGetAction{
		Path:   HealthPath,
		Port:   intstr.FromString(portName),
		Scheme: core.Spec.Components.TLS.GetScheme(),
	}

	containerPorts := []corev1.ContainerPort{{
		Name:          portName,
		ContainerPort: int32(port),
		Protocol:      corev1.ProtocolTCP,
	}}

	if core.Spec.Metrics.IsEnabled() {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          harbormetav1.CoreMetricsPortName,
			ContainerPort: core.Spec.Metrics.Port,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: version.NewVersionAnnotations(core.Annotations),
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
					Annotations: core.Spec.ComponentSpec.TemplateAnnotations,
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken:  &varFalse,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Volumes:                       volumes,
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup:    &fsGroup,
						RunAsGroup: &runAsGroup,
						RunAsUser:  &runAsUser,
					},
					Containers: []corev1.Container{{
						Name:  controllers.Core.String(),
						Image: image,
						Ports: containerPorts,
						// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/core/env.jinja
						Env: envs,
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: httpGET,
							},
							PeriodSeconds: int32(healthCheckPeriod.Seconds()),
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: httpGET,
							},
						},
						VolumeMounts: volumeMounts,
					}},
				},
			},
		},
	}

	core.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, nil
}

func addDatabaseEnvs(core *goharborv1.Core, envs []corev1.EnvVar) ([]corev1.EnvVar, error) {
	if core.Spec.Database.MaxIdleConnections != nil {
		maxConns, err := harbor.EnvVar(common.PostGreSQLMaxIdleConns, harbor.Value(fmt.Sprintf("%d", *core.Spec.Database.MaxIdleConnections)))
		if err != nil {
			return nil, errors.Wrap(err, "cannot configure max idle connections")
		}

		envs = append(envs, maxConns)
	}

	if core.Spec.Database.MaxOpenConnections != nil {
		maxConns, err := harbor.EnvVar(common.PostGreSQLMaxOpenConns, harbor.Value(fmt.Sprintf("%d", *core.Spec.Database.MaxOpenConnections)))
		if err != nil {
			return nil, errors.Wrap(err, "cannot configure max open connections")
		}

		envs = append(envs, maxConns)
	}

	if sslMode, ok := core.Spec.Database.Parameters[harbormetav1.PostgresSSLModeKey]; ok {
		sslModeVar, err := harbor.EnvVar(common.PostGreSQLSSLMode, harbor.Value(sslMode))
		if err != nil {
			return nil, errors.Wrap(err, "cannot configure ssl mode")
		}

		envs = append(envs, sslModeVar)
	}

	return envs, nil
}
