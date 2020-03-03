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

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/pkg/errors"
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

func (r *Reconciler) GetDeployment(ctx context.Context, core *goharborv1alpha2.Core) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := core.Spec.GetImage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	var envs []corev1.EnvVar

	if len(core.Spec.RegistryCacheSecret) > 0 {
		envs = append(envs, corev1.EnvVar{
			Name: "_REDIS_URL_REG",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key:      goharborv1alpha2.RegistryCacheURLKey,
					Optional: &varTrue,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: core.Spec.RegistryCacheSecret,
					},
				},
			},
		})
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-core", core.GetName()),
			Namespace: core.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"core-name":      core.GetName(),
					"core-namespace": core.GetNamespace(),
				},
			},
			Replicas: core.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"operator/version": application.GetVersion(ctx),
					},
					Labels: map[string]string{
						"core-name":      core.GetName(),
						"core-namespace": core.GetNamespace(),
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 core.Spec.NodeSelector,
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
										Name: core.Name,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  configName,
											Path: configName,
										},
									},
									Optional: &varFalse,
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
									SecretName: core.Name,
								},
							},
						}, {
							Name: "certificate",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: core.Name,
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
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "core",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(port),
								},
							},

							// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/core/env.jinja
							Env: append(envs, corev1.EnvVar{
								Name:  "EXT_ENDPOINT",
								Value: core.Spec.PublicURL,
							}, corev1.EnvVar{
								Name:  "LOG_LEVEL",
								Value: core.Spec.LogLevel,
							}, corev1.EnvVar{
								Name:  "AUTH_MODE",
								Value: "db_auth",
							}, corev1.EnvVar{
								Name:  "DATABASE_TYPE",
								Value: "postgresql",
							}, corev1.EnvVar{
								Name:  "CORE_URL",
								Value: fmt.Sprintf("http://%s", core.GetName()),
							}, corev1.EnvVar{
								Name:  "CORE_LOCAL_URL",
								Value: fmt.Sprintf("http://%s", core.GetName()),
							}, corev1.EnvVar{
								Name:  "READ_ONLY",
								Value: fmt.Sprintf("%+v", core.Spec.ReadOnly),
							}, corev1.EnvVar{
								Name: "CORE_SECRET",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key:      goharborv1alpha2.CoreSecretKey,
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Name,
										},
									},
								},
							}, corev1.EnvVar{
								Name: "JOBSERVICE_SECRET",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key:      goharborv1alpha2.JobServiceSecretKey,
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Name,
										},
									},
								},
							}, corev1.EnvVar{
								Name:  "CLAIR_ADAPTER_URL",
								Value: core.Spec.ClairAdapterURL,
							}, corev1.EnvVar{
								Name:  "CLAIR_URL",
								Value: core.Spec.ClairURL,
							}, corev1.EnvVar{
								Name:  "CHART_REPOSITORY_URL",
								Value: core.Spec.ChartRepositoryURL,
							}, corev1.EnvVar{
								Name:  "JOBSERVICE_URL",
								Value: core.Spec.JobServiceURL,
							}, corev1.EnvVar{
								Name:  "NOTARY_URL",
								Value: core.Spec.NotaryURL,
							}, corev1.EnvVar{
								Name:  "REGISTRY_URL",
								Value: core.Spec.RegistryURL,
							}, corev1.EnvVar{
								Name:  "REGISTRYCTL_URL",
								Value: core.Spec.RegistryControllerURL,
							}, corev1.EnvVar{
								Name:  "TOKEN_SERVICE_URL",
								Value: fmt.Sprintf("http://%s/service/token", core.GetName()),
							}, corev1.EnvVar{
								Name:  "CONFIG_PATH",
								Value: path.Join(coreConfigPath, configFileName),
							}, corev1.EnvVar{
								Name:  "CFG_EXPIRATION",
								Value: fmt.Sprintf("%.0f", core.Spec.ConfigExpiration.Seconds()),
							}, corev1.EnvVar{
								Name:  "RELOAD_KEY",
								Value: "true",
							}, corev1.EnvVar{
								Name:  "SYNC_QUOTA",
								Value: fmt.Sprintf("%+v", core.Spec.SyncQuota),
							}, corev1.EnvVar{
								Name:  "SYNC_REGISTRY",
								Value: fmt.Sprintf("%+v", core.Spec.SyncRegistry),
							}, corev1.EnvVar{
								Name: "HARBOR_ADMIN_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key:      goharborv1alpha2.CoreAdminPasswordKey,
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.AdminPasswordSecret,
										},
									},
								},
							}, corev1.EnvVar{
								Name:  "WITH_CHARTMUSEUM",
								Value: fmt.Sprintf("%+v", core.Spec.ChartRepositoryURL != ""),
							}, corev1.EnvVar{
								Name:  "WITH_CLAIR",
								Value: fmt.Sprintf("%+v", core.Spec.ClairURL != ""),
							}, corev1.EnvVar{
								Name:  "WITH_NOTARY",
								Value: fmt.Sprintf("%+v", core.Spec.NotaryURL != ""),
							}, corev1.EnvVar{
								// Not supported yet
								Name:  "WITH_TRIVY",
								Value: "false",
							}),
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.DatabaseSecret,
										},
										Optional: &varFalse,
									},
								}, {
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.ClairDatabaseSecret,
										},
										Optional: &varTrue,
									},
								}, {
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: core.Spec.SessionRedisSecret,
										},
										Optional: &varTrue,
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
					Priority: core.Spec.Priority,
				},
			},
			RevisionHistoryLimit: &revisionHistoryLimit,
		},
	}, nil
}
