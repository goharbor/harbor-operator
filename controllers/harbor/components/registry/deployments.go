package registry

import (
	"context"
	"fmt"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	initImage   = "hairyhenderson/gomplate"
	apiPort     = 5000 // https://github.com/docker/distribution/blob/749f6afb4572201e3c37325d0ffedb6f32be8950/contrib/compose/docker-compose.yml#L15
	metricsPort = 5001 // https://github.com/docker/distribution/blob/b12bd4004afc203f1cbd2072317c8fda30b89710/cmd/registry/config-dev.yml#L34
	ctlAPIPort  = 8080 // https://github.com/goharbor/harbor/blob/2fb1cc89d9ef9313842cc68b4b7c36be73681505/src/common/const.go#L134
)

var (
	revisionHistoryLimit  int32 = 0 // nolint:golint
	registryConfigPath          = "/etc/registry/"
	registryCtlConfigPath       = "/etc/registryctl/"
	varFalse                    = false
	varTrue                     = true
)

func (r *Registry) GetDeployments(ctx context.Context) []*appsv1.Deployment { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := r.harbor.GetName()

	cacheEnv := corev1.EnvVar{
		Name: "REDIS_URL",
	}
	if len(r.harbor.Spec.Components.Registry.CacheSecret) > 0 {
		cacheEnv.ValueFrom = &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:      containerregistryv1alpha1.HarborRegistryURLKey,
				Optional: &varTrue,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: r.harbor.Spec.Components.Registry.CacheSecret,
				},
			},
		}
	}

	var storageVolumeSource corev1.VolumeSource
	if r.harbor.Spec.Components.Registry.StorageSecret == "" {
		storageVolumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
	} else {
		storageVolumeSource.Secret = &corev1.SecretVolumeSource{
			SecretName: r.harbor.Spec.Components.Registry.StorageSecret,
		}
	}

	return []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      r.harbor.NormalizeComponentName(containerregistryv1alpha1.RegistryName),
				Namespace: r.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.RegistryName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":      containerregistryv1alpha1.RegistryName,
						"harbor":   harborName,
						"operator": operatorName,
					},
				},
				Replicas: r.harbor.Spec.Components.Registry.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"checksum":         r.GetConfigCheckSum(),
							"operator/version": application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      containerregistryv1alpha1.RegistryName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						NodeSelector:                 r.harbor.Spec.Components.Registry.NodeSelector,
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
											Name: r.harbor.NormalizeComponentName(containerregistryv1alpha1.RegistryName),
										},
									},
								},
							}, {
								Name:         "config-storage",
								VolumeSource: storageVolumeSource,
							}, {
								Name: "certificate",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: r.harbor.NormalizeComponentName(containerregistryv1alpha1.CertificateName),
									},
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
										MountPath: path.Join("/workdir", registryConfigName),
										ReadOnly:  true,
										SubPath:   registryConfigName,
									}, {
										Name:      "config-template",
										MountPath: path.Join("/workdir", registryCtlConfigName),
										ReadOnly:  true,
										SubPath:   registryCtlConfigName,
									}, {
										Name:      "config-storage",
										MountPath: "/opt/configuration/storage",
										ReadOnly:  true,
									}, {
										Name:      "config",
										MountPath: "/processed",
										ReadOnly:  false,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "STORAGE_CONFIG",
										Value: "/opt/configuration/storage",
									}, {
										Name:  "CORE_HOSTNAME",
										Value: r.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
									}, {
										Name:  "METRICS_ADDRESS",
										Value: fmt.Sprintf(":%d", metricsPort),
									}, {
										Name:  "API_ADDRESS",
										Value: fmt.Sprintf(":%d", apiPort),
									}, {
										Name:  "REGISTRYCTL_PORT",
										Value: fmt.Sprintf("%d", ctlAPIPort),
									},
									cacheEnv,
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "registryctl",
								Image: r.harbor.Spec.Components.RegistryCtl.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: ctlAPIPort,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name: "CORE_SECRET",
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      "secret",
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: r.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
												},
											},
										},
									}, {
										Name: "JOBSERVICE_SECRET", // TODO check if it is necessary
										ValueFrom: &corev1.EnvVarSource{
											SecretKeyRef: &corev1.SecretKeySelector{
												Key:      "secret",
												Optional: &varFalse,
												LocalObjectReference: corev1.LocalObjectReference{
													Name: r.harbor.NormalizeComponentName(containerregistryv1alpha1.JobServiceName),
												},
											},
										},
									}, {
										Name:  "REGISTRY_HTTP_HOST",
										Value: r.harbor.Spec.PublicURL,
									}, {
										Name:  "REGISTRY_AUTH_TOKEN_REALM",
										Value: fmt.Sprintf("%s/service/token", r.harbor.Spec.PublicURL),
									}, {
										Name:  "REGISTRY_LOG_FIELDS_OPERATOR",
										Value: operatorName,
									}, {
										Name:  "REGISTRY_LOG_FIELDS_HARBOR",
										Value: harborName,
									},
								},
								ImagePullPolicy: corev1.PullAlways,
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/api/health",
											Port: intstr.FromInt(ctlAPIPort),
										},
									},
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/api/health",
											Port: intstr.FromInt(ctlAPIPort),
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: path.Join(registryConfigPath, defaultRegistryConfigName),
										Name:      "config",
										SubPath:   registryConfigName,
									}, {
										MountPath: path.Join(registryCtlConfigPath, registryCtlConfigName),
										Name:      "config",
										SubPath:   registryCtlConfigName,
									}, {
										MountPath: "/etc/registry/root.crt",
										Name:      "certificate",
										SubPath:   "tls.crt",
									},
								},
								Command: []string{"/home/harbor/harbor_registryctl"},
								Args:    []string{"-c", path.Join(registryCtlConfigPath, registryCtlConfigName)},
							}, {
								Name:  "registry",
								Image: r.harbor.Spec.Components.Registry.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: apiPort,
									}, {
										ContainerPort: metricsPort,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "REGISTRY_HTTP_HOST",
										Value: r.harbor.Spec.PublicURL,
									}, {
										Name:  "REGISTRY_AUTH_TOKEN_REALM",
										Value: fmt.Sprintf("%s/service/token", r.harbor.Spec.PublicURL),
									}, {
										Name:  "REGISTRY_LOG_FIELDS_OPERATOR",
										Value: operatorName,
									}, {
										Name:  "REGISTRY_LOG_FIELDS_HARBOR",
										Value: harborName,
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
								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: path.Join(registryConfigPath, registryConfigName),
										Name:      "config",
										SubPath:   registryConfigName,
									}, {
										MountPath: "/etc/registry/root.crt",
										Name:      "certificate",
										SubPath:   "tls.crt",
									},
								},
								Command: []string{"/usr/bin/registry"},
								Args:    []string{"serve", path.Join(registryConfigPath, registryConfigName)},
							},
						},
						Priority: r.Option.GetPriority(),
					},
				},
				RevisionHistoryLimit: &revisionHistoryLimit,
				Paused:               r.harbor.Spec.Paused,
			},
		},
	}
}
