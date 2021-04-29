package registry

import (
	"context"
	"path"
	"strings"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/image"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	VolumeName                            = "registry-config"
	ConfigPath                            = "/etc/registry"
	CompatibilitySchema1Path              = ConfigPath + "/compatibility-schema1"
	CompatibilitySchema1VolumeName        = "compatibility-schema1-certificate"
	AuthenticationHTPasswdPath            = ConfigPath + "/auth"
	AuthenticationHTPasswdVolumeName      = "authentication-htpasswd"
	InternalCertificatesVolumeName        = "internal-certificates"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
	InternalCertificatesPath              = ConfigPath + "/ssl"
	StorageName                           = "storage"
	StoragePath                           = "/var/lib/registry"
	HealthPath                            = "/"
	StorageServiceCAName                  = "storage-service-ca"
	StorageServiceCAMountPath             = "/harbor_cust_cert/custom-ca-bundle.crt"
)

var (
	varFalse = false
	varTrue  = true

	fsGroup    int64 = 10000
	runAsGroup int64 = 10000
	runAsUser  int64 = 10000
)

const (
	apiPort     = 5000 // https://github.com/docker/distribution/blob/749f6afb4572201e3c37325d0ffedb6f32be8950/contrib/compose/docker-compose.yml#L15
	metricsPort = 5001 // https://github.com/docker/distribution/blob/b12bd4004afc203f1cbd2072317c8fda30b89710/cmd/registry/config-dev.yml#L34
)

func (r *Reconciler) GetDeployment(ctx context.Context, registry *goharborv1.Registry) (*appsv1.Deployment, error) { // nolint:funlen
	getImageOptions := []image.Option{
		image.WithConfigstore(r.ConfigStore),
		image.WithImageFromSpec(registry.Spec.Image),
		image.WithHarborVersion(version.GetVersion(registry.Annotations)),
	}

	image, err := image.GetImage(ctx, harbormetav1.RegistryComponent.String(), getImageOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, registry.GetName())
	namespace := registry.GetNamespace()

	envs := []corev1.EnvVar{}

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
	}}

	volumeMounts := []corev1.VolumeMount{{
		Name:      VolumeName,
		MountPath: ConfigPath,
	}}

	if registry.Spec.HTTP.SecretRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: harbormetav1.RegistryHTTPSecret,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: harbormetav1.RegistryHTTPSecret,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registry.Spec.HTTP.SecretRef,
					},
					Optional: &varFalse,
				},
			},
		})
	}

	if registry.Spec.Redis != nil && len(registry.Spec.Redis.PasswordRef) > 0 {
		envs = append(envs, corev1.EnvVar{
			Name: "REGISTRY_REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: harbormetav1.RedisPasswordKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registry.Spec.Redis.PasswordRef,
					},
					Optional: &varTrue,
				},
			},
		})
	}

	envs = append(envs, registry.Spec.Proxy.GetEnvVars()...)

	if registry.Spec.Compatibility.Schema1.Enabled {
		envs = append(envs, corev1.EnvVar{
			Name:  "REGISTRY_COMPATIBILITY_SCHEMA1_SIGNINGKEYFILE",
			Value: path.Join(CompatibilitySchema1Path, corev1.TLSPrivateKeyKey),
		})

		volumes = append(volumes, corev1.Volume{
			Name: CompatibilitySchema1VolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registry.Spec.Compatibility.Schema1.CertificateRef,
					Optional:   &varFalse,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			MountPath: CompatibilitySchema1Path,
			Name:      CompatibilitySchema1VolumeName,
			ReadOnly:  true,
		})
	}

	if registry.Spec.Authentication.HTPasswd != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "REGISTRY_AUTH_HTPASSWD_PATH",
			Value: path.Join(AuthenticationHTPasswdPath, harbormetav1.HTPasswdFileName),
		})

		volumes = append(volumes, corev1.Volume{
			Name: AuthenticationHTPasswdVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registry.Spec.Authentication.HTPasswd.SecretRef,
					Optional:   &varFalse,
					Items: []corev1.KeyToPath{{
						Key:  harbormetav1.HTPasswdFileName,
						Path: harbormetav1.HTPasswdFileName,
					}},
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			MountPath: AuthenticationHTPasswdPath,
			Name:      AuthenticationHTPasswdVolumeName,
			ReadOnly:  true,
		})
	}

	if registry.Spec.Storage.Driver.S3 != nil && registry.Spec.Storage.Driver.S3.CertificateRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: StorageServiceCAName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registry.Spec.Storage.Driver.S3.CertificateRef,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      StorageServiceCAName,
			MountPath: StorageServiceCAMountPath,
			ReadOnly:  true,
			SubPath:   corev1.ServiceAccountRootCAKey,
		})
	}

	if registry.Spec.HTTP.TLS.Enabled() {
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
					SecretName: registry.Spec.HTTP.TLS.CertificateRef,
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

	// inject certs if need.
	if registry.Spec.CertificateInjection.ShouldInject() {
		volumes = append(volumes, registry.Spec.CertificateInjection.GenerateVolumes()...)
		volumeMounts = append(volumeMounts, registry.Spec.CertificateInjection.GenerateVolumeMounts()...)
	}

	httpGET := &corev1.HTTPGetAction{
		Path:   HealthPath,
		Port:   intstr.FromString(harbormetav1.RegistryAPIPortName),
		Scheme: registry.Spec.HTTP.TLS.GetScheme(),
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: version.NewVersionAnnotations(registry.Annotations),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"):      name,
					r.Label("namespace"): namespace,
				},
			},
			Replicas: registry.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
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
					Containers: []corev1.Container{{
						Name:  controllers.Registry.String(),
						Image: image,
						Ports: []corev1.ContainerPort{{
							ContainerPort: apiPort,
							Name:          harbormetav1.RegistryAPIPortName,
							Protocol:      corev1.ProtocolTCP,
						}, {
							ContainerPort: metricsPort,
							Name:          harbormetav1.RegistryMetricsPortName,
							Protocol:      corev1.ProtocolTCP,
						}},
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: httpGET,
							},
						},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: httpGET,
							},
						},
						VolumeMounts: volumeMounts,
						Args:         []string{"serve", path.Join(ConfigPath, ConfigName)},
						Env:          envs,
					}},
				},
			},
		},
	}

	err = r.ApplyStorageConfiguration(ctx, registry, deploy)

	registry.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, errors.Wrap(err, "cannot apply storage configuration")
}

const registryContainerIndex = 0

func (r *Reconciler) ApplyFilesystemStorageEnvs(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error {
	regContainer := &deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
		Name:         StorageName,
		VolumeSource: registry.Spec.Storage.Driver.FileSystem.VolumeSource,
	})

	regContainer.VolumeMounts = append(regContainer.VolumeMounts, corev1.VolumeMount{
		Name:      StorageName,
		MountPath: StoragePath,
		SubPath:   strings.TrimLeft(registry.Spec.Storage.Driver.FileSystem.Prefix, "/"),
		ReadOnly:  false,
	})

	return nil
}

func (r *Reconciler) ApplyS3StorageEnvs(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error {
	regContainer := &deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	regContainer.Env = append(regContainer.Env, corev1.EnvVar{
		Name: "REGISTRY_STORAGE_S3_SECRETKEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: harbormetav1.SharedSecretKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.S3.SecretKeyRef,
				},
			},
		},
	})

	return nil
}

func (r *Reconciler) ApplySwiftStorageEnvs(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error {
	regContainer := &deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	regContainer.Env = append(regContainer.Env, corev1.EnvVar{
		Name: "REGISTRY_STORAGE_SWIFT_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: harbormetav1.SharedSecretKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.Swift.PasswordRef,
				},
			},
		},
	}, corev1.EnvVar{
		Name: "REGISTRY_STORAGE_SWIFT_SECRETKEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: harbormetav1.SharedSecretKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.Swift.SecretKeyRef,
				},
			},
		},
	})

	return nil
}

func (r *Reconciler) ApplyInMemoryStorageEnvs(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error {
	regContainer := &deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	regContainer.Env = append(regContainer.Env, corev1.EnvVar{
		Name:  "REGISTRY_STORAGE",
		Value: "inmemory",
	})

	return nil
}

var errNoStorageDriverFound = errors.New("no storage driver found")

func (r *Reconciler) ApplyStorageConfiguration(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error {
	if registry.Spec.Storage.Driver.S3 != nil {
		return r.ApplyS3StorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.Swift != nil {
		return r.ApplySwiftStorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.FileSystem != nil {
		return r.ApplyFilesystemStorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.InMemory != nil {
		return r.ApplyInMemoryStorageEnvs(ctx, registry, deploy)
	}

	return serrors.UnrecoverrableError(errNoStorageDriverFound, serrors.InvalidSpecReason, "unable to configure storage")
}
