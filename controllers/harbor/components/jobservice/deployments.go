package jobservice

import (
	"context"
	"fmt"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
)

var (
	revisionHistoryLimit int32 = 0 // nolint:golint
	varFalse                   = false
)

const (
	initImage  = "hairyhenderson/gomplate"
	configPath = "/etc/jobservice/"
	port       = 8080
)

func (j *JobService) GetDeployments(ctx context.Context) []*appsv1.Deployment { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := j.harbor.GetName()

	return []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      j.harbor.NormalizeComponentName(goharborv1alpha1.JobServiceName),
				Namespace: j.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.JobServiceName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":      goharborv1alpha1.JobServiceName,
						"harbor":   harborName,
						"operator": operatorName,
					},
				},
				Replicas: j.harbor.Spec.Components.JobService.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"configuration/checksum": j.GetConfigMapsCheckSum(),
							"secret/checksum":        j.GetSecretsCheckSum(),
							"operator/version":       application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      goharborv1alpha1.JobServiceName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						NodeSelector:                 j.harbor.Spec.Components.JobService.NodeSelector,
						AutomountServiceAccountToken: &varFalse,
						Volumes: []corev1.Volume{
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
											Name: j.harbor.NormalizeComponentName(goharborv1alpha1.JobServiceName),
										},
									},
								},
							}, {
								Name: "logs",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
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
										Name:  "PORT",
										Value: fmt.Sprintf("%d", port),
									}, {
										Name:  "LOGS_DIR",
										Value: logsDirectory,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "jobservice",
								Image: j.harbor.Spec.Components.JobService.GetImage(),
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: port,
									},
								},

								// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/jobservice/env.jinja
								Env: []corev1.EnvVar{
									{
										Name: "CORE_SECRET",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      "secret",
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: j.harbor.NormalizeComponentName(goharborv1alpha1.CoreName),
												},
											},
										},
									}, {
										Name: "JOBSERVICE_SECRET",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      "secret",
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: j.harbor.NormalizeComponentName(goharborv1alpha1.JobServiceName),
												},
											},
										},
									}, {
										Name: "CORE_URL",
										ValueFrom: &corev1.EnvVarSource{
											ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
												Key:      "CORE_URL",
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: j.harbor.NormalizeComponentName(goharborv1alpha1.CoreName),
												},
											},
										},
									}, {
										Name: "JOB_SERVICE_POOL_REDIS_URL",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      goharborv1alpha1.HarborJobServiceBrokerURLKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: j.harbor.Spec.Components.JobService.RedisSecret,
												},
											},
										},
									}, {
										Name: "JOB_SERVICE_POOL_REDIS_NAMESPACE",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      goharborv1alpha1.HarborJobServiceBrokerNamespaceKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: j.harbor.Spec.Components.JobService.RedisSecret,
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
												Name: j.harbor.NormalizeComponentName(goharborv1alpha1.JobServiceName),
											},
										},
									},
								},
								Command:         []string{"/harbor/harbor_jobservice"},
								Args:            []string{"-c", path.Join(configPath, configName)},
								ImagePullPolicy: corev1.PullAlways,
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/api/v1/stats",
											Port: intstr.FromInt(port),
										},
									},
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/api/v1/stats",
											Port: intstr.FromInt(port),
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: path.Join(configPath, configName),
										Name:      "config",
										SubPath:   configName,
									}, {
										MountPath: logsDirectory,
										Name:      "logs",
									},
								},
							},
						},
						Priority: j.Option.GetPriority(),
					},
				},
				RevisionHistoryLimit: &revisionHistoryLimit,
				Paused:               j.harbor.Spec.Paused,
			},
		},
	}
}
