package registryctl

import (
	"context"
	"fmt"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/pkg/errors"
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

func (r *Reconciler) GetDeployment(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := registryCtl.Spec.GetImage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	cacheEnv := corev1.EnvVar{
		Name: "REDIS_URL",
	}
	if len(registryCtl.Spec.CacheSecret) > 0 {
		cacheEnv.ValueFrom = &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key:      goharborv1alpha2.RegistryCacheURLKey,
				Optional: &varTrue,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: registryCtl.Spec.CacheSecret,
				},
			},
		}
	}

	var storageVolumeSource corev1.VolumeSource
	if registryCtl.Spec.StorageSecret == "" {
		storageVolumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
	} else {
		storageVolumeSource.Secret = &corev1.SecretVolumeSource{
			SecretName: registryCtl.Spec.StorageSecret,
		}
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-registryctl", registryCtl.GetName()),
			Namespace: registryCtl.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"registryctl-name":      registryCtl.GetName(),
					"registryctl-namespace": registryCtl.GetNamespace(),
				},
			},
			Replicas: registryCtl.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"registryctl-name":      registryCtl.GetName(),
						"registryctl-namespace": registryCtl.GetNamespace(),
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 registryCtl.Spec.NodeSelector,
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
										Name: registryCtl.Name,
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
									SecretName: registryCtl.Name,
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
									Name: "CORE_HOSTNAME",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: registryCtl.Spec.ConfigName,
											},
											Key: goharborv1alpha2.CoreURLKey,
										},
									},
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
							Image: image,
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
											Key:      goharborv1alpha2.CoreSecretKey,
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: registryCtl.Spec.CoreSecret,
											},
										},
									},
								}, {
									Name: "JOBSERVICE_SECRET", // TODO check if it is necessary
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key:      goharborv1alpha2.JobServiceSecretKey,
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: registryCtl.Spec.JobServiceSecret,
											},
										},
									},
								}, {
									Name: "REGISTRY_HTTP_HOST",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: registryCtl.Spec.ConfigName,
											},
											Key:      goharborv1alpha2.RegistryCorePublicURLKey,
											Optional: &varFalse,
										},
									},
								}, {
									Name: "REGISTRY_AUTH_TOKEN_REALM",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: registryCtl.Spec.ConfigName,
											},
											Key:      goharborv1alpha2.RegistryAuthURLKey,
											Optional: &varFalse,
										},
									},
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
						},
					},
					Priority: registryCtl.Spec.Priority,
				},
			},
			RevisionHistoryLimit: &revisionHistoryLimit,
		},
	}, nil
}
