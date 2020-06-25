package chartmuseum

import (
	"context"
	"path"

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
	ConfigPath = "/etc/chartmuseum"
	HealthPath = "/health"
	port       = 8080 // https://github.com/helm/chartmuseum/blob/969515a51413e1f1840fb99509401aa3c63deccd/pkg/config/vars.go#L135
	VolumeName = "config"
)

func (r *Reconciler) GetDeployment(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, chartMuseum.GetName())
	namespace := chartMuseum.GetNamespace()

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

	if chartMuseum.Spec.Auth.BasicAuthRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "BASIC_AUTH_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: chartMuseum.Spec.Auth.BasicAuthRef,
					},
					Key: corev1.BasicAuthUsernameKey,
				},
			},
		}, corev1.EnvVar{
			Name: "BASIC_AUTH_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: chartMuseum.Spec.Auth.BasicAuthRef,
					},
					Key: corev1.BasicAuthPasswordKey,
				},
			},
		})
	}

	envFroms := []corev1.EnvFromSource{}

	if chartMuseum.Spec.Chart.Storage.SecretRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "STORAGE",
			Value: chartMuseum.Spec.Chart.Storage.Name,
		})

		envFroms = append(envFroms, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: chartMuseum.Spec.Chart.Storage.SecretRef,
				},
			},
			Prefix: "STORAGE_",
		})
	}

	if chartMuseum.Spec.Cache.Redis.PasswordRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "CACHE_REDIS_DB",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: chartMuseum.Spec.Cache.Redis.PasswordRef,
					},
					Key:      goharborv1alpha2.RedisPasswordKey,
					Optional: &varFalse,
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

					Containers: []corev1.Container{
						{
							Name:  "chartmuseum",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
								},
							},
							Command: []string{"/home/chart/chartm"},
							Args:    []string{"-c", path.Join(ConfigPath, ConfigName)},

							EnvFrom: envFroms,
							Env:     envs,

							VolumeMounts: volumesMount,

							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: HealthPath,
										Port: intstr.FromInt(port),
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: HealthPath,
										Port: intstr.FromInt(port),
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
