package chartmuseum

import (
	"context"
	"path"

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
	varTrue                    = true
)

const (
	initImage  = "hairyhenderson/gomplate"
	configPath = "/etc/chartmuseum/"
	port       = 8080 // https://github.com/helm/chartmuseum/blob/969515a51413e1f1840fb99509401aa3c63deccd/pkg/config/vars.go#L135
)

func (c *ChartMuseum) GetDeployments(ctx context.Context) []*appsv1.Deployment { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := c.harbor.GetName()

	volumes := []corev1.Volume{{
		Name: "chartmuseum",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	}}
	volumeMounts := []corev1.VolumeMount{{
		MountPath: "/mnt/chartmuseum",
		Name:      "chartmuseum",
	}}
	envs := []corev1.EnvVar{{
		Name:  "STORAGE",
		Value: "local",
	}, {
		Name:  "STORAGE_LOCAL_ROOTDIR",
		Value: "/mnt/chartmuseum",
	}}
	envFroms := []corev1.EnvFromSource{}

	if c.harbor.Spec.Components.ChartMuseum.StorageSecret != "" {
		volumes = []corev1.Volume{}
		volumeMounts = []corev1.VolumeMount{}

		envs = []corev1.EnvVar{{
			Name: "STORAGE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: c.harbor.Spec.Components.ChartMuseum.StorageSecret,
					},
					Key: containerregistryv1alpha1.HarborChartMuseumStorageKindKey,
				},
			},
		}}

		envFroms = []corev1.EnvFromSource{{
			SecretRef: &corev1.SecretEnvSource{
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: c.harbor.Spec.Components.ChartMuseum.StorageSecret,
				},
			},
			Prefix: "STORAGE_",
		}, {
			// Some storage driver requires environment variable, add it from secret data
			// See https://chartmuseum.com/docs/#using-with-openstack-object-storage
			SecretRef: &corev1.SecretEnvSource{
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: c.harbor.Spec.Components.ChartMuseum.StorageSecret,
				},
			},
		}}
	}

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
							"configuration/checksum": c.GetConfigMapsCheckSum(),
							"secret/checksum":        c.GetSecretsCheckSum(),
							"operator/version":       application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      containerregistryv1alpha1.ChartMuseumName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						NodeSelector:                 c.harbor.Spec.Components.ChartMuseum.NodeSelector,
						AutomountServiceAccountToken: &varFalse,
						Volumes: append([]corev1.Volume{
							{
								Name: "config",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							}, {
								Name: "config-template",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
										},
									},
								},
							},
						}, volumes...),
						InitContainers: []corev1.Container{
							{
								Name:            "configuration",
								Image:           initImage,
								WorkingDir:      "/workdir",
								Args:            []string{"--input-dir", "/workdir", "--output-dir", "/processed"},
								SecurityContext: &corev1.SecurityContext{},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "config-template",
										MountPath: path.Join("/workdir", configName),
										ReadOnly:  true,
										SubPath:   configName,
									}, {
										Name:      "config",
										MountPath: "/processed",
										ReadOnly:  false,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name: "CACHE_URL",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.ChartMuseum.CacheSecret,
												},
												Key:      containerregistryv1alpha1.HarborChartMuseumCacheURLKey,
												Optional: &varTrue,
											},
										},
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "chartmuseum",
								Image: c.harbor.Spec.Components.ChartMuseum.GetImage(),
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: port,
									},
								},
								Command: []string{"/home/chart/chartm"},
								Args:    []string{"-c", path.Join(configPath, configName)},

								VolumeMounts: append(volumeMounts, corev1.VolumeMount{
									MountPath: path.Join(configPath, configName),
									Name:      "config",
									SubPath:   configName,
								}),

								Env: append([]corev1.EnvVar{
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
								}, envs...),

								EnvFrom: append(envFroms, corev1.EnvFromSource{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
										},
									},
								}),

								ImagePullPolicy: corev1.PullAlways,
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/health",
											Port: intstr.FromInt(port),
										},
									},
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/health",
											Port: intstr.FromInt(port),
										},
									},
								},
							},
						},
						Priority: c.Option.GetPriority(),
					},
				},
				RevisionHistoryLimit: &revisionHistoryLimit,
				Paused:               c.harbor.Spec.Paused,
			},
		},
	}
}
