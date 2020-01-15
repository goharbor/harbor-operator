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
	initImage = "hairyhenderson/gomplate"
)

var (
	revisionHistoryLimit  int32 = 0 // nolint:golint
	registryConfigPath          = "/etc/registry/"
	registryCtlConfigPath       = "/etc/registryctl/"
	varFalse                    = false
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
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: r.harbor.Spec.Components.Registry.CacheSecret,
				},
			},
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
				Replicas: r.harbor.Spec.Components.Core.Replicas,
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
						NodeSelector: r.harbor.Spec.Components.Registry.NodeSelector,
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
								Name: "config-storage",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: r.harbor.Spec.Components.Registry.StorageSecret,
									},
								},
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
								Name:            "registry-configuration",
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
										ContainerPort: 8080,
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
									},
									{
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
										Name:  "REGISTRY_STORAGE_INMEMORY",
										Value: "",
									}, {
										Name:  "REGISTRY_HTTP_HOST",
										Value: fmt.Sprintf("https://%s", r.harbor.Spec.PublicURL),
									}, {
										Name:  "REGISTRY_AUTH_TOKEN_REALM",
										Value: fmt.Sprintf("https://%s/service/token", r.harbor.Spec.PublicURL),
									}, {
										Name:  "REGISTRY_NOTIFICATION_ENDPOINTS_0_URL",
										Value: r.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
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
											Port: intstr.FromInt(8080),
										},
									},
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/api/health",
											Port: intstr.FromInt(8080),
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: path.Join(registryCtlConfigPath, registryConfigName),
										Name:      "config",
										SubPath:   registryCtlConfigName,
									}, {
										MountPath: "/etc/registry/root.crt",
										Name:      "certificate",
										SubPath:   "tls.crt",
									}, {
										MountPath: "/etc/ssl/certs/harbor.pem",
										Name:      "certificate",
										SubPath:   "ca.crt",
									},
								},
							},
							{
								Name:  "registry",
								Image: r.harbor.Spec.Components.Registry.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 5000,
									}, {
										ContainerPort: 5001,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "REGISTRY_HTTP_HOST",
										Value: fmt.Sprintf("https://%s", r.harbor.Spec.PublicURL),
									}, {
										Name:  "REGISTRY_AUTH_TOKEN_REALM",
										Value: fmt.Sprintf("https://%s/service/token", r.harbor.Spec.PublicURL),
									}, {
										Name:  "REGISTRY_NOTIFICATION_ENDPOINTS_0_URL",
										Value: r.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
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
											Port:   intstr.FromInt(5000),
											Scheme: corev1.URISchemeHTTP,
										},
									},
									InitialDelaySeconds: 20,
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path:   "/",
											Port:   intstr.FromInt(5000),
											Scheme: corev1.URISchemeHTTP,
										},
									},
									InitialDelaySeconds: 20,
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
									}, {
										MountPath: "/etc/ssl/certs/harbor.pem",
										Name:      "certificate",
										SubPath:   "ca.crt",
									},
								},
								Args: []string{"serve", path.Join(registryConfigPath, registryConfigName)},
							},
						},
						Priority: r.Option.Priority,
					},
				},
				RevisionHistoryLimit: &revisionHistoryLimit,
				Paused:               r.harbor.Spec.Paused,
			},
		},
	}
}
