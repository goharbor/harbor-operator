package core

import (
	"context"
	"fmt"
	"path"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

var (
	revisionHistoryLimit int32 = 0 // nolint:golint
	maxIdleConns               = 0
	maxOpenConns               = 1
	varFalse                   = false
	varTrue                    = true
)

const (
	initImage      = "hairyhenderson/gomplate"
	coreConfigPath = "/etc/core"
	keyFileName    = "key"
	configFileName = "app.conf"
	port           = 8080 // https://github.com/goharbor/harbor/blob/2fb1cc89d9ef9313842cc68b4b7c36be73681505/src/common/const.go#L127

	healthCheckPeriod = 90 * time.Second
)

func (c *HarborCore) GetDeployments(ctx context.Context) []*appsv1.Deployment { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := c.harbor.GetName()

	cacheEnv := corev1.EnvVar{
		Name: "_REDIS_URL_REG",
	}
	if len(c.harbor.Spec.Components.Registry.CacheSecret) > 0 {
		cacheEnv.ValueFrom = &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:      containerregistryv1alpha1.HarborRegistryURLKey,
				Optional: &varTrue,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: c.harbor.Spec.Components.Registry.CacheSecret,
				},
			},
		}
	}

	return []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.CoreName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":      containerregistryv1alpha1.CoreName,
						"harbor":   harborName,
						"operator": operatorName,
					},
				},
				Replicas: c.harbor.Spec.Components.Core.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"checksum":         c.GetConfigCheckSum(),
							"operator/version": application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      containerregistryv1alpha1.CoreName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						NodeSelector: c.harbor.Spec.Components.Core.NodeSelector,
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
											Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
										},
									},
								},
							}, {
								Name: "secret-key",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										Items: []corev1.KeyToPath{
											{
												Key:  secretKey,
												Path: keyFileName,
											},
										},
										Optional:   &varFalse,
										SecretName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
									},
								},
							}, {
								Name: "certificate",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CertificateName),
									},
								},
							}, {
								Name: "psc",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
						InitContainers: []corev1.Container{
							{
								Name:            "registry-configuration",
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
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "core",
								Image: c.harbor.Spec.Components.Core.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: int32(port),
									},
								},

								// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/core/env.jinja
								Env: []corev1.EnvVar{
									{
										Name: "CORE_SECRET",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      "secret",
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
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
													Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.JobServiceName),
												},
											},
										},
									}, {
										Name: "HARBOR_ADMIN_PASSWORD",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborAdminPasswordKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.AdminPasswordSecret,
												},
											},
										},
									}, {
										Name: "POSTGRESQL_HOST",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabaseHostKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "POSTGRESQL_PORT",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabasePortKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "POSTGRESQL_DATABASE",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabaseNameKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "POSTGRESQL_USERNAME",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabaseUserKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "POSTGRESQL_PASSWORD",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabasePasswordKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "CLAIR_DB_HOST",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabaseHostKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "CLAIR_DB_PORT",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabasePortKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "CLAIR_DB_DATABASE",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabaseNameKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "CLAIR_DB_USERNAME",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabaseUserKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									}, {
										Name: "CLAIR_DB_PASSWORD",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      containerregistryv1alpha1.HarborCoreDatabasePasswordKey,
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: c.harbor.Spec.Components.Core.DatabaseSecret,
												},
											},
										},
									},
									cacheEnv,
								},
								EnvFrom: []corev1.EnvFromSource{
									{
										ConfigMapRef: &corev1.ConfigMapEnvSource{
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
											},
										},
									},
								},
								ImagePullPolicy: corev1.PullAlways,
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/api/ping",
											Port: intstr.FromInt(port),
										},
									},
									PeriodSeconds: int32(healthCheckPeriod.Seconds()),
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/api/ping",
											Port: intstr.FromInt(port),
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "config",
										ReadOnly:  true,
										MountPath: path.Join(coreConfigPath, configFileName),
										SubPath:   configFileName,
									}, {
										Name:      "secret-key",
										ReadOnly:  true,
										MountPath: path.Join(coreConfigPath, keyFileName),
										SubPath:   keyFileName,
									}, {
										Name:      "certificate",
										ReadOnly:  true,
										MountPath: path.Join(coreConfigPath, "private_key.pem"),
										SubPath:   "tls.key",
									}, {
										Name:      "psc",
										ReadOnly:  false,
										MountPath: path.Join(coreConfigPath, "token"),
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
