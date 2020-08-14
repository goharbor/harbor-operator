package registryctl

import (
	"context"
	"path"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
)

const (
	VolumeName                            = "registry-config"
	ConfigPath                            = "/etc/registryctl"
	CompatibilitySchema1Path              = ConfigPath + "/compatibility-schema1"
	CompatibilitySchema1VolumeName        = "compatibility-schema1-certificate"
	AuthenticationHTPasswdPath            = ConfigPath + "/auth"
	AuthenticationHTPasswdVolumeName      = "authentication-htpasswd"
	InternalCertificatesVolumeName        = "internal-certificates"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
	InternalCertificatesPath              = ConfigPath + "/ssl"
	StorageName                           = "storage"
	StoragePath                           = "/var/lib/registry"
	HealthPath                            = "/api/health"
)

var (
	varFalse = false
	varTrue  = true
)

const (
	httpsPort = 8443
	httpPort  = 8080
)

func (r *Reconciler) GetDeployment(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, registryCtl.GetName())
	namespace := registryCtl.GetNamespace()

	registry, err := r.GetRegistry(ctx, registryCtl)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get registry")
	}

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

	volumeMounts := []corev1.VolumeMount{{
		Name:      VolumeName,
		MountPath: ConfigPath,
	}}

	if registry.Spec.HTTP.SecretRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "REGISTRY_HTTP_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: harbormetav1.SharedSecretKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registry.Spec.HTTP.SecretRef,
					},
					Optional: &varFalse,
				},
			},
		})
	}

	if registry.Spec.Redis != nil {
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

	if registry.Spec.Proxy != nil {
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
					Items: []corev1.KeyToPath{
						{
							Key:  harbormetav1.HTPasswdFileName,
							Path: harbormetav1.HTPasswdFileName,
						},
					},
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			MountPath: AuthenticationHTPasswdPath,
			Name:      AuthenticationHTPasswdVolumeName,
			ReadOnly:  true,
		})
	}

	if registryCtl.Spec.TLS.Enabled() {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: path.Join(InternalCertificateAuthorityDirectory, corev1.ServiceAccountRootCAKey),
			SubPath:   corev1.ServiceAccountRootCAKey,
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
					SecretName: registryCtl.Spec.TLS.CertificateRef,
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

	port := harbormetav1.RegistryControllerHTTPPortName
	if registryCtl.Spec.TLS.Enabled() {
		port = harbormetav1.RegistryControllerHTTPSPortName
	}

	httpGET := &corev1.HTTPGetAction{
		Path:   HealthPath,
		Port:   intstr.FromString(port),
		Scheme: registryCtl.Spec.TLS.GetScheme(),
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
			Replicas: registryCtl.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 registryCtl.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					Containers: []corev1.Container{
						{
							Name:  "registryctl",
							Image: image,
							Ports: []corev1.ContainerPort{{
								Name:          harbormetav1.RegistryControllerHTTPPortName,
								ContainerPort: httpPort,
							}, {
								Name:          harbormetav1.RegistryControllerHTTPSPortName,
								ContainerPort: httpsPort,
							}},
							ImagePullPolicy: corev1.PullAlways,
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
							Env:          envs,
						},
					},
				},
			},
		},
	}

	err = r.ApplyStorageConfiguration(ctx, registry, deploy)

	return deploy, errors.Wrap(err, "cannot apply storage configuration")
}
