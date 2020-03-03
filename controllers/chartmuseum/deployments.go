package chartmuseum

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

func (r *Reconciler) GetDeployment(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := chartMuseum.Spec.GetImage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

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

	if chartMuseum.Spec.StorageSecret != "" {
		volumes = []corev1.Volume{}
		volumeMounts = []corev1.VolumeMount{}

		envs = []corev1.EnvVar{{
			Name: "STORAGE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: chartMuseum.Spec.StorageSecret,
					},
					Key: goharborv1alpha2.ChartMuseumStorageKindKey,
				},
			},
		}}

		envFroms = []corev1.EnvFromSource{{
			SecretRef: &corev1.SecretEnvSource{
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: chartMuseum.Spec.StorageSecret,
				},
			},
			Prefix: "STORAGE_",
		}, {
			// Some storage driver requires environment variable, add it from secret data
			// See https://chartmuseum.com/docs/#using-with-openstack-object-storage
			SecretRef: &corev1.SecretEnvSource{
				Optional: &varFalse,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: chartMuseum.Spec.StorageSecret,
				},
			},
		}}
	}

	initEnv := []corev1.EnvVar{}

	if chartMuseum.Spec.CacheSecret != "" {
		initEnv = []corev1.EnvVar{
			{
				Name: "CACHE_URL",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: chartMuseum.Spec.CacheSecret,
						},
						Key:      goharborv1alpha2.ChartMuseumCacheURLKey,
						Optional: &varTrue,
					},
				},
			},
		}
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-chartmuseum", chartMuseum.GetName()),
			Namespace: chartMuseum.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"chartmuseum-name":      chartMuseum.GetName(),
					"chartmuseum-namespace": chartMuseum.GetNamespace(),
				},
			},
			Replicas: chartMuseum.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"chartmuseum-name":      chartMuseum.GetName(),
						"chartmuseum-namespace": chartMuseum.GetNamespace(),
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 chartMuseum.Spec.NodeSelector,
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
										Name: chartMuseum.Name,
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
							Env: initEnv,
						},
					},
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
											Key:      goharborv1alpha2.ChartMuseumBasicAuthKey,
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: chartMuseum.Spec.SecretName,
											},
										},
									},
								}, {
									Name:  "PORT",
									Value: fmt.Sprintf("%d", port),
								}, {
									Name:  "CHART_URL",
									Value: chartMuseum.Spec.PublicURL,
								},
							}, envs...),

							EnvFrom: envFroms,

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
					Priority: chartMuseum.Spec.Priority,
				},
			},
			RevisionHistoryLimit: &revisionHistoryLimit,
		},
	}, nil
}
