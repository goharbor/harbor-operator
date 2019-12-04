package chartmuseum

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

var (
	revisionHistoryLimit int32 = 0 // nolint:golint
	varFalse                   = false
)

func (c *ChartMuseum) GetDeployments(ctx context.Context) []*appsv1.Deployment { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := c.harbor.GetName()

	return []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.ChartMuseumName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":      containerregistryv1alpha1.ChartMuseumName,
						"harbor":   harborName,
						"operator": operatorName,
					},
				},
				Replicas: c.harbor.Spec.Components.ChartMuseum.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"operator/version": application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      containerregistryv1alpha1.ChartMuseumName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "chartmuseum",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{
										Medium: corev1.StorageMediumMemory,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "chartmuseum",
								Image: c.harbor.Spec.Components.ChartMuseum.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: port,
									},
								},

								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: "/mnt/chartmuseum",
										Name:      "chartmuseum",
									},
								},

								Env: []corev1.EnvVar{
									{
										Name: "BASIC_AUTH_PASS",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      "secret",
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
												},
											},
										},
									},
								},

								EnvFrom: []corev1.EnvFromSource{
									{
										ConfigMapRef: &corev1.ConfigMapEnvSource{
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
											},
										},
									},
								},

								ImagePullPolicy: corev1.PullAlways,
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/health",
											Port: intstr.FromInt(8080),
										},
									},
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/health",
											Port: intstr.FromInt(8080),
										},
									},
								},
							},
						},
						Priority: c.Option.Priority,
					},
				},
				RevisionHistoryLimit: &revisionHistoryLimit,
				Paused:               c.harbor.Spec.Paused,
			},
		},
	}
}
