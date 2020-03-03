package clair

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/pkg/errors"
)

const (
	initImage       = "hairyhenderson/gomplate"
	apiPort         = 6060 // https://github.com/quay/clair/blob/c39101e9b8206401d8b9cb631f3aee47a24ab889/cmd/clair/config.go#L64
	healthPort      = 6061 // https://github.com/quay/clair/blob/c39101e9b8206401d8b9cb631f3aee47a24ab889/cmd/clair/config.go#L63
	adapterPort     = 8080
	clairConfigPath = "/etc/clair"

	livenessProbeInitialDelay = 300 * time.Second
)

var (
	revisionHistoryLimit int32 = 0 // nolint:golint
	varFalse                   = false
)

func (r *Reconciler) GetDeployment(ctx context.Context, clair *goharborv1alpha2.Clair) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := clair.Spec.GetImage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	adapterImage, err := clair.Spec.GetAdapterImage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get adapter image")
	}

	vulnsrc, err := json.Marshal(clair.Spec.VulnerabilitySources)
	if err != nil {
		logger.Get(ctx).Error(err, "invalid vulnerability sources")
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-clair", clair.GetName()),
			Namespace: clair.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"clair-name":      clair.GetName(),
					"clair-namespace": clair.GetNamespace(),
				},
			},
			Replicas: clair.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clair-name":      clair.GetName(),
						"clair-namespace": clair.GetNamespace(),
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 clair.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes: []corev1.Volume{
						{
							Name: "config-template",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: clair.Name,
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
											Name: clair.Spec.DatabaseSecret,
										},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "clair",
							Image: image,
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
											Key:      goharborv1alpha2.HarborClairDatabasePasswordKey,
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: clair.Spec.DatabaseSecret,
											},
										},
									},
								},
							},
							Command:         []string{"/home/clair/clair"},
							Args:            []string{"-config", path.Join(clairConfigPath, configKey)},
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
									MountPath: path.Join(clairConfigPath, configKey),
									Name:      "config",
									SubPath:   configKey,
								},
							},
						}, {
							Name:  "clair-adapter",
							Image: adapterImage,
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
											Key:      goharborv1alpha2.HarborClairAdapterBrokerURLKey,
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: clair.Spec.Adapter.RedisSecret,
											},
										},
									},
								}, {
									Name: "SCANNER_STORE_REDIS_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key:      goharborv1alpha2.HarborClairAdapterBrokerNamespaceKey,
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: clair.Spec.Adapter.RedisSecret,
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
											Name: clair.Spec.DatabaseSecret,
										},
									},
								}, {
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: clair.Name,
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
									MountPath: path.Join(clairConfigPath, configKey),
									Name:      "config",
									SubPath:   configKey,
								},
							},
						},
					},
					Priority: clair.Spec.Priority,
				},
			},
			RevisionHistoryLimit: &revisionHistoryLimit,
		},
	}, nil
}
