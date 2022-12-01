package registry

import (
	"context"
	"path"
	"strings"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/image"
	utilStrings "github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	VolumeName                            = "registry-config"
	CtlVolumeName                         = "registryctl-config"
	ConfigPath                            = "/etc/registry"
	CtlConfigPath                         = "/etc/registryctl"
	CompatibilitySchema1Path              = ConfigPath + "/compatibility-schema1"
	CompatibilitySchema1VolumeName        = "compatibility-schema1-certificate"
	AuthenticationHTPasswdPath            = ConfigPath + "/auth"
	AuthenticationHTPasswdVolumeName      = "authentication-htpasswd"
	InternalCertificatesVolumeName        = "internal-certificates"
	CtlInternalCertificatesVolumeName     = "ctl-internal-certificates"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
	InternalCertificatesPath              = ConfigPath + "/ssl"
	CtlInternalCertificatesPath           = CtlConfigPath + "/ssl"
	StorageName                           = "storage"
	StoragePath                           = "/var/lib/registry"
	HealthPath                            = "/"
	CtlHealthPath                         = "/api/health"
	StorageServiceCAName                  = "storage-service-ca"
	StorageServiceCAMountPath             = "/harbor_cust_cert/custom-ca-bundle.crt"
	GcsJSONKeyFilePath                    = "/etc/gcs/gcs-key.json"
)

var (
	varFalse = false
	varTrue  = true

	fsGroup    int64 = 10000
	runAsGroup int64 = 10000
	runAsUser  int64 = 10000
)

const (
	apiPort = 5000 // https://github.com/docker/distribution/blob/749f6afb4572201e3c37325d0ffedb6f32be8950/contrib/compose/docker-compose.yml#L15
	// registry controller port.
	httpsPort = 8443
	httpPort  = 8080
)

func (r *Reconciler) GetDeployment(ctx context.Context, registry *goharborv1.Registry) (*appsv1.Deployment, error) { //nolint:funlen
	getImageOptions := []image.Option{
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

	if registry.Spec.Storage.Driver.Gcs != nil && registry.Spec.Storage.Driver.Gcs.KeyDataRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "gcs-key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registry.Spec.Storage.Driver.Gcs.KeyDataRef,
					Items: []corev1.KeyToPath{
						{
							Key:  "GCS_KEY_DATA",
							Path: "gcs-key.json",
						},
					},
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "gcs-key",
			MountPath: GcsJSONKeyFilePath,
			SubPath:   "gcs-key.json",
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

	containerPorts := []corev1.ContainerPort{
		{
			ContainerPort: apiPort,
			Name:          harbormetav1.RegistryAPIPortName,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	if registry.Spec.HTTP.Debug != nil {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: registry.Spec.HTTP.Debug.Port,
			Name:          harbormetav1.RegistryMetricsPortName,
			Protocol:      corev1.ProtocolTCP,
		})
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
					Annotations: registry.Spec.ComponentSpec.TemplateAnnotations,
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
						Ports: containerPorts,
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: httpGET,
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
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

	if err = r.ApplyStorageConfiguration(ctx, registry, deploy); err != nil {
		return nil, errors.Wrap(err, "cannot apply storage configuration")
	}

	// attach registry controller container
	if err = r.attachRegistryCtlContainer(ctx, registry, deploy); err != nil {
		return nil, errors.Wrap(err, "cannot attach registryctl container")
	}

	registry.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, nil
}

func (r *Reconciler) attachRegistryCtlContainer(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error { //nolint:funlen
	registryCtl, err := r.GetRegistryCtl(ctx, registry)
	if err != nil {
		return errors.Wrap(err, "can not get registryctl from registry")
	}

	getImageOptions := []image.Option{
		image.WithImageFromSpec(registryCtl.Spec.Image),
		image.WithHarborVersion(version.GetVersion(registryCtl.Annotations)),
	}

	image, err := image.GetImage(ctx, harbormetav1.RegistryControllerComponent.String(), getImageOptions...)
	if err != nil {
		return errors.Wrap(err, "cannot get registryctl image")
	}

	name := controllers.RegistryController.String()

	volumeMounts := deploy.Spec.Template.Spec.Containers[0].VolumeMounts
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      CtlVolumeName,
		MountPath: CtlConfigPath,
	})

	if registryCtl.Spec.TLS.Enabled() {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      CtlInternalCertificatesVolumeName,
			MountPath: path.Join(InternalCertificateAuthorityDirectory, "ctl-"+corev1.ServiceAccountRootCAKey),
			SubPath:   strings.TrimLeft(corev1.ServiceAccountRootCAKey, "/"),
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      CtlInternalCertificatesVolumeName,
			MountPath: CtlInternalCertificatesPath,
			ReadOnly:  true,
		})

		deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: CtlInternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registryCtl.Spec.TLS.CertificateRef,
				},
			},
		})
	}

	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: CtlVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: utilStrings.NormalizeName(registryCtl.GetName(), RegistryCtlName),
				},
				Optional: &varFalse,
			},
		},
	})

	envs := deploy.Spec.Template.Spec.Containers[0].Env
	if registryCtl.Spec.Authentication.JobServiceSecretRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "JOBSERVICE_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registryCtl.Spec.Authentication.JobServiceSecretRef,
					},
					Key: harbormetav1.SharedSecretKey,
				},
			},
		})
	}

	if registryCtl.Spec.Authentication.CoreSecretRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "CORE_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registryCtl.Spec.Authentication.CoreSecretRef,
					},
					Key: harbormetav1.SharedSecretKey,
				},
			},
		})
	}

	envs, err = registry.Spec.Trace.AddEnvVars(harbormetav1.RegistryControllerComponent.String(), envs)
	if err != nil {
		return errors.Wrap(err, "get trace environment variables")
	}

	ports := []corev1.ContainerPort{{
		Name:          harbormetav1.RegistryControllerHTTPPortName,
		ContainerPort: httpPort,
		Protocol:      corev1.ProtocolTCP,
	}, {
		Name:          harbormetav1.RegistryControllerHTTPSPortName,
		ContainerPort: httpsPort,
		Protocol:      corev1.ProtocolTCP,
	}}

	port := harbormetav1.RegistryControllerHTTPPortName
	if registryCtl.Spec.TLS.Enabled() {
		port = harbormetav1.RegistryControllerHTTPSPortName
	}

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   CtlHealthPath,
				Port:   intstr.FromString(port),
				Scheme: registryCtl.Spec.TLS.GetScheme(),
			},
		},
	}

	container := &corev1.Container{
		Name:           name,
		Image:          image,
		Env:            envs,
		Ports:          ports,
		LivenessProbe:  probe,
		ReadinessProbe: probe,
		VolumeMounts:   volumeMounts,
	}
	deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, *container)

	return nil
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

	if registry.Spec.Storage.Driver.S3.SecretKeyRef != "" {
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
	}

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

func (r *Reconciler) ApplyGcsStorageEnvs(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error {
	regContainer := &deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	regContainer.Env = append(regContainer.Env, corev1.EnvVar{
		Name: "GCS_KEY_DATA",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: "GCS_KEY_DATA",
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.Gcs.KeyDataRef,
				},
			},
		},
	})

	return nil
}

func (r *Reconciler) ApplyOssStorageEnvs(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error {
	regContainer := &deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	regContainer.Env = append(regContainer.Env, corev1.EnvVar{
		Name: "REGISTRY_STORAGE_OSS_ACCESSKEYSECRET",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: harbormetav1.SharedSecretKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.Oss.AccessSecretRef,
				},
			},
		},
	})

	return nil
}

func (r *Reconciler) ApplyAzureStorageEnvs(ctx context.Context, registry *goharborv1.Registry, deploy *appsv1.Deployment) error {
	regContainer := &deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	regContainer.Env = append(regContainer.Env, corev1.EnvVar{
		Name: "REGISTRY_STORAGE_AZURE_ACCOUNTKEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: harbormetav1.SharedSecretKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.Azure.AccountKeyRef,
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

	if registry.Spec.Storage.Driver.Azure != nil {
		return r.ApplyAzureStorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.Oss != nil {
		return r.ApplyOssStorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.Gcs != nil {
		return r.ApplyGcsStorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.FileSystem != nil {
		return r.ApplyFilesystemStorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.InMemory != nil {
		return r.ApplyInMemoryStorageEnvs(ctx, registry, deploy)
	}

	return serrors.UnrecoverrableError(errNoStorageDriverFound, serrors.InvalidSpecReason, "unable to configure storage")
}
