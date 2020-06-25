package registry

import (
	"context"
	"path"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	apiPort                          = 5000 // https://github.com/docker/distribution/blob/749f6afb4572201e3c37325d0ffedb6f32be8950/contrib/compose/docker-compose.yml#L15
	metricsPort                      = 5001 // https://github.com/docker/distribution/blob/b12bd4004afc203f1cbd2072317c8fda30b89710/cmd/registry/config-dev.yml#L34
	VolumeName                       = "registry-config"
	ConfigPath                       = "/etc/registry"
	CompatibilitySchema1Path         = ConfigPath + "/compatibility-schema1"
	CompatibilitySchema1VolumeName   = "compatibility-schema1-certificate"
	AuthenticationHTPasswdPath       = ConfigPath + "/auth"
	AuthenticationHTPasswdVolumeName = "authentication-htpasswd"
	StorageName                      = "storage"
	StoragePath                      = "/var/lib/registry"
)

var (
	varFalse = false
	varTrue  = true
)

func (r *Reconciler) GetDeployment(ctx context.Context, registry *goharborv1alpha2.Registry) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, registry.GetName())
	namespace := registry.GetNamespace()

	envs := []corev1.EnvVar{}

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
		},
	}

	volumesMount := []corev1.VolumeMount{
		{
			Name:      VolumeName,
			MountPath: ConfigPath,
		},
	}

	if registry.Spec.HTTP.SecretRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "REGISTRY_HTTP_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: goharborv1alpha2.SharedSecretKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registry.Spec.HTTP.SecretRef,
					},
					Optional: &varFalse,
				},
			},
		})
	}

	if registry.Spec.Redis.PasswordRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "REGISTRY_REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: goharborv1alpha2.RedisPasswordKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registry.Spec.Redis.PasswordRef,
					},
					Optional: &varTrue,
				},
			},
		})
	}

	if registry.Spec.Proxy.BasicAuthRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "REGISTRY_PROXY_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: corev1.BasicAuthPasswordKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registry.Spec.Proxy.BasicAuthRef,
					},
					Optional: &varTrue,
				},
			},
		}, corev1.EnvVar{
			Name: "REGISTRY_PROXY_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: corev1.BasicAuthUsernameKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registry.Spec.Proxy.BasicAuthRef,
					},
					Optional: &varTrue,
				},
			},
		})
	}

	if registry.Spec.Compatibility.Schema1.CertificateRef != "" {
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

		volumesMount = append(volumesMount, corev1.VolumeMount{
			MountPath: CompatibilitySchema1Path,
			Name:      CompatibilitySchema1VolumeName,
			ReadOnly:  true,
		})
	}

	if registry.Spec.Authentication.HTPasswd.SecretRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "REGISTRY_AUTH_HTPASSWD_PATH",
			Value: path.Join(AuthenticationHTPasswdPath, goharborv1alpha2.HTPasswdFileName),
		})

		volumes = append(volumes, corev1.Volume{
			Name: AuthenticationHTPasswdVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registry.Spec.Authentication.HTPasswd.SecretRef,
					Optional:   &varFalse,
					Items: []corev1.KeyToPath{
						{
							Key:  goharborv1alpha2.HTPasswdFileName,
							Path: goharborv1alpha2.HTPasswdFileName,
						},
					},
				},
			},
		})

		volumesMount = append(volumesMount, corev1.VolumeMount{
			MountPath: AuthenticationHTPasswdPath,
			Name:      AuthenticationHTPasswdVolumeName,
			ReadOnly:  true,
		})
	}

	deploy := &appsv1.Deployment{
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
			Replicas: registry.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 registry.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					Containers: []corev1.Container{
						{
							Name:  "registry",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: apiPort,
								}, {
									ContainerPort: metricsPort,
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/",
										Port:   intstr.FromInt(apiPort),
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/",
										Port:   intstr.FromInt(apiPort),
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							VolumeMounts: volumesMount,
							Args:         []string{"serve", path.Join(ConfigPath, ConfigName)},
							Env:          envs,
						},
					},
					Priority: registry.Spec.Priority,
				},
			},
		},
	}

	err = r.ApplyStorageConfiguration(ctx, registry, deploy)

	return deploy, errors.Wrap(err, "cannot apply storage configuration")
}

const registryContainerIndex = 0

func (r *Reconciler) GetFilesystemStorageEnvs(ctx context.Context, registry *goharborv1alpha2.Registry, deploy *appsv1.Deployment) error {
	regContainer := deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
		Name:         StorageName,
		VolumeSource: registry.Spec.Storage.Driver.FileSystem.VolumeSource,
	})

	regContainer.VolumeMounts = append(regContainer.VolumeMounts, corev1.VolumeMount{
		Name:      StorageName,
		MountPath: StoragePath,
		ReadOnly:  false,
	})

	return nil
}

func (r *Reconciler) GetS3StorageEnvs(ctx context.Context, registry *goharborv1alpha2.Registry, deploy *appsv1.Deployment) error {
	regContainer := deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	regContainer.Env = append(regContainer.Env, corev1.EnvVar{
		Name: "REGISTRY_STORAGE_S3_SECRETKEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.S3.SecretKeyRef,
				},
			},
		},
	})

	return nil
}

func (r *Reconciler) GetSwiftStorageEnvs(ctx context.Context, registry *goharborv1alpha2.Registry, deploy *appsv1.Deployment) error {
	regContainer := deploy.Spec.Template.Spec.Containers[registryContainerIndex]

	regContainer.Env = append(regContainer.Env, corev1.EnvVar{
		Name: "REGISTRY_STORAGE_SWIFT_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.Swift.PasswordRef,
				},
			},
		},
	}, corev1.EnvVar{
		Name: "REGISTRY_STORAGE_SWIFT_SECRETKEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registry.Spec.Storage.Driver.Swift.SecretKeyRef,
				},
			},
		},
	})

	return nil
}

func (r *Reconciler) ApplyStorageConfiguration(ctx context.Context, registry *goharborv1alpha2.Registry, deploy *appsv1.Deployment) error {
	if registry.Spec.Storage.Driver.FileSystem != nil {
		return r.GetFilesystemStorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.S3 != nil {
		return r.GetS3StorageEnvs(ctx, registry, deploy)
	}

	if registry.Spec.Storage.Driver.Swift != nil {
		return r.GetSwiftStorageEnvs(ctx, registry, deploy)
	}

	return nil
}
