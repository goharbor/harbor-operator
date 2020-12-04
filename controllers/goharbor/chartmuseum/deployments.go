package chartmuseum

import (
	"context"
	"path"
	"strings"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ConfigPath                            = "/etc/chartmuseum"
	HealthPath                            = "/health"
	VolumeName                            = "config"
	InternalCertificatesVolumeName        = "internal-certificates"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
	InternalCertificatesPath              = ConfigPath + "/ssl"
	LocalStorageVolume                    = "storage"
	DefaultLocalStoragePath               = "/mnt/chartstorage"
)

var varFalse = false

const (
	httpsPort = 8443
	httpPort  = 8080
)

func (r *Reconciler) GetDeployment(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, chartMuseum.GetName())
	namespace := chartMuseum.GetNamespace()

	envs := []corev1.EnvVar{{
		Name:  "CONFIG",
		Value: path.Join(ConfigPath, ConfigName),
	}}

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

	if chartMuseum.Spec.Authentication.BasicAuthRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "BASIC_AUTH_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: chartMuseum.Spec.Authentication.BasicAuthRef,
					},
					Key: corev1.BasicAuthUsernameKey,
				},
			},
		}, corev1.EnvVar{
			Name: "BASIC_AUTH_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: chartMuseum.Spec.Authentication.BasicAuthRef,
					},
					Key: corev1.BasicAuthPasswordKey,
				},
			},
		})
	}

	envFroms := []corev1.EnvFromSource{}

	if chartMuseum.Spec.Chart.URL != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "CHART_URL",
			Value: chartMuseum.Spec.Chart.URL,
		})
	}

	if chartMuseum.Spec.Chart.Storage.Amazon != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "STORAGE",
			Value: "amazon",
		}, corev1.EnvVar{
			Name:  "AWS_ACCESS_KEY_ID",
			Value: chartMuseum.Spec.Chart.Storage.Amazon.AccessKeyID,
		}, corev1.EnvVar{
			Name:  "STORAGE_AMAZON_BUCKET",
			Value: chartMuseum.Spec.Chart.Storage.Amazon.Bucket,
		}, corev1.EnvVar{
			Name:  "STORAGE_AMAZON_PREFIX",
			Value: chartMuseum.Spec.Chart.Storage.Amazon.Prefix,
		}, corev1.EnvVar{
			Name:  "STORAGE_AMAZON_REGION",
			Value: chartMuseum.Spec.Chart.Storage.Amazon.Region,
		}, corev1.EnvVar{
			Name:  "STORAGE_AMAZON_ENDPOINT",
			Value: chartMuseum.Spec.Chart.Storage.Amazon.Endpoint,
		})

		if chartMuseum.Spec.Chart.Storage.Amazon.AccessSecretRef != "" {
			envs = append(envs, corev1.EnvVar{
				Name: "AWS_SECRET_ACCESS_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: chartMuseum.Spec.Chart.Storage.Amazon.AccessSecretRef,
						},
						Key: harbormetav1.SharedSecretKey,
					},
				},
			})
		}
	}

	if chartMuseum.Spec.Chart.Storage.OpenStack != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "STORAGE",
			Value: "openstack",
		}, corev1.EnvVar{
			Name:  "OS_AUTH_URL",
			Value: chartMuseum.Spec.Chart.Storage.OpenStack.AuthenticationURL,
		}, corev1.EnvVar{
			Name: "OS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: chartMuseum.Spec.Chart.Storage.OpenStack.PasswordRef,
					},
					Key: harbormetav1.SharedSecretKey,
				},
			},
		}, corev1.EnvVar{
			Name:  "STORAGE_OPENSTACK_PREFIX",
			Value: chartMuseum.Spec.Chart.Storage.OpenStack.Prefix,
		}, corev1.EnvVar{
			Name:  "STORAGE_OPENSTACK_REGION",
			Value: chartMuseum.Spec.Chart.Storage.OpenStack.Region,
		}, corev1.EnvVar{
			Name:  "STORAGE_OPENSTACK_CONTAINER",
			Value: chartMuseum.Spec.Chart.Storage.OpenStack.Container,
		})

		if chartMuseum.Spec.Chart.Storage.OpenStack.Username != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "OS_USERNAME",
				Value: chartMuseum.Spec.Chart.Storage.OpenStack.Username,
			})
		} else {
			envs = append(envs, corev1.EnvVar{
				Name:  "OS_USERID",
				Value: chartMuseum.Spec.Chart.Storage.OpenStack.UserID,
			})
		}

		if chartMuseum.Spec.Chart.Storage.OpenStack.Tenant != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "OS_TENANT_NAME",
				Value: chartMuseum.Spec.Chart.Storage.OpenStack.Tenant,
			})
		} else {
			envs = append(envs, corev1.EnvVar{
				Name:  "OS_TENANT_ID",
				Value: chartMuseum.Spec.Chart.Storage.OpenStack.TenantID,
			})
		}
	}

	if chartMuseum.Spec.Chart.Storage.FileSystem != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "STORAGE",
			Value: "local",
		}, corev1.EnvVar{
			Name:  "STORAGE_LOCAL_ROOTDIR",
			Value: path.Join(DefaultLocalStoragePath, chartMuseum.Spec.Chart.Storage.FileSystem.Prefix),
		})

		volumes = append(volumes, corev1.Volume{
			Name:         LocalStorageVolume,
			VolumeSource: chartMuseum.Spec.Chart.Storage.FileSystem.VolumeSource,
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      LocalStorageVolume,
			MountPath: DefaultLocalStoragePath,
			SubPath:   strings.TrimLeft(chartMuseum.Spec.Chart.Storage.FileSystem.Prefix, "/"),
			ReadOnly:  false,
		})
	}

	if chartMuseum.Spec.Cache.Redis != nil {
		envs = append(envs, corev1.EnvVar{
			Name: "CACHE_REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: chartMuseum.Spec.Cache.Redis.PasswordRef,
					},
					Key:      harbormetav1.RedisPasswordKey,
					Optional: &varFalse,
				},
			},
		})
	}

	if chartMuseum.Spec.Server.TLS.Enabled() {
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
					SecretName: chartMuseum.Spec.Server.TLS.CertificateRef,
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

	port := harbormetav1.ChartMuseumHTTPPortName
	if chartMuseum.Spec.Server.TLS.Enabled() {
		port = harbormetav1.ChartMuseumHTTPSPortName
	}

	httpGET := &corev1.HTTPGetAction{
		Path:   HealthPath,
		Port:   intstr.FromString(port),
		Scheme: chartMuseum.Spec.Server.TLS.GetScheme(),
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
			Replicas: chartMuseum.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 chartMuseum.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,

					Containers: []corev1.Container{{
						Name:  controllers.ChartMuseum.String(),
						Image: image,
						Ports: []corev1.ContainerPort{{
							Name:          harbormetav1.ChartMuseumHTTPPortName,
							ContainerPort: httpPort,
						}, {
							Name:          harbormetav1.ChartMuseumHTTPSPortName,
							ContainerPort: httpsPort,
						}},

						EnvFrom: envFroms,
						Env:     envs,

						VolumeMounts: volumeMounts,

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
					}},
				},
			},
		},
	}

	chartMuseum.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, nil
}
