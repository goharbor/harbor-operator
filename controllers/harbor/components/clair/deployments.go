package clair

import (
	"context"
	"encoding/json"
	"path"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
	"github.com/ovh/harbor-operator/pkg/factories/logger"
)

const (
	initImage   = "hairyhenderson/gomplate"
	apiPort     = 6060 // https://github.com/quay/clair/blob/c39101e9b8206401d8b9cb631f3aee47a24ab889/cmd/clair/config.go#L64
	healthPort  = 6061 // https://github.com/quay/clair/blob/c39101e9b8206401d8b9cb631f3aee47a24ab889/cmd/clair/config.go#L63
	adapterPort = 8080

	livenessProbeInitialDelay = 300 * time.Second
)

var (
	revisionHistoryLimit int32 = 0 // nolint:golint
	varFalse                   = false
)

func (c *Clair) GetDeployments(ctx context.Context) []*appsv1.Deployment { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := c.harbor.GetName()

	vulnsrc, err := json.Marshal(c.harbor.Spec.Components.Clair.VulnerabilitySources)
	if err != nil {
		logger.Get(ctx).Error(err, "invalid vulnerability sources")
	}

	return []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.ClairName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.ClairName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":      containerregistryv1alpha1.ClairName,
						"harbor":   harborName,
						"operator": operatorName,
					},
				},
				Replicas: c.harbor.Spec.Components.Clair.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"checksum":         c.GetConfigCheckSum(),
							"operator/version": application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      containerregistryv1alpha1.ClairName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						NodeSelector: c.harbor.Spec.Components.Clair.NodeSelector,
						Volumes: []corev1.Volume{
							{
								Name: "config-template",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.ClairName),
										},
										Items: []corev1.KeyToPath{
											{
												Key:  configKey,
												Path: configKey,
											},
										},
									},
								},
							}, {
								Name:         "config",
								VolumeSource: corev1.VolumeSource{},
							},
						},
						InitContainers: []corev1.Container{
							{
								Name:       "configuration",
								Image:      initImage,
								WorkingDir: "/workdir",
								Args:       []string{"--input-dir", "/workdir", "--output-dir", "/processed"},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "config-template",
										MountPath: "/workdir",
										ReadOnly:  true,
									}, {
										Name:      "config",
										MountPath: "/processed",
										ReadOnly:  false,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "vulnsrc",
										Value: string(vulnsrc),
									},
								},
								EnvFrom: []corev1.EnvFromSource{
									{
										SecretRef: &corev1.SecretEnvSource{
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: c.harbor.Spec.Components.Clair.DatabaseSecret,
											},
										},
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "clair",
								Image: c.harbor.Spec.Components.Clair.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: apiPort,
									}, {
										ContainerPort: healthPort,
									},
								},

								Env: []corev1.EnvVar{
									{ // // https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/clair/clair_env.jinja
										Name:  "HTTP_PROXY",
										Value: "",
									}, {
										Name:  "HTTPS_PROXY",
										Value: "",
									}, {
										Name:  "NO_PROXY",
										Value: "",
									}, { // https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/clair/postgres_env.jinja
										Name: "POSTGRES_PASSWORD",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborClairDatabasePasswordKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Clair.DatabaseSecret,
												},
											},
										},
									},
								},
								ImagePullPolicy: corev1.PullAlways,
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/health",
											Port: intstr.FromInt(healthPort),
										},
									},
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/health",
											Port: intstr.FromInt(healthPort),
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: path.Join("/etc/clair", configKey),
										Name:      "config",
										SubPath:   configKey,
									},
								},
							}, {
								Name:  "clair-adapter",
								Image: c.harbor.Spec.Components.Clair.Adapter.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: adapterPort,
									},
								},

								Env: []corev1.EnvVar{
									{
										Name: "SCANNER_STORE_REDIS_URL",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborClairAdapterBrokerURLKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Clair.Adapter.RedisSecret,
												},
											},
										},
									}, {
										Name: "SCANNER_STORE_REDIS_NAMESPACE",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborClairAdapterBrokerNamespaceKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Clair.Adapter.RedisSecret,
												},
											},
										},
									},
								},

								EnvFrom: []corev1.EnvFromSource{
									{
										Prefix: "clair_db_",
										SecretRef: &corev1.SecretEnvSource{
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: c.harbor.Spec.Components.Clair.DatabaseSecret,
											},
										},
									},
								},

								ImagePullPolicy: corev1.PullAlways,
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/probe/healthy",
											Port: intstr.FromInt(adapterPort),
										},
									},
									InitialDelaySeconds: int32(livenessProbeInitialDelay.Seconds()),
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/probe/healthy",
											Port: intstr.FromInt(adapterPort),
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: path.Join("/etc/clair", configKey),
										Name:      "config",
										SubPath:   configKey,
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
