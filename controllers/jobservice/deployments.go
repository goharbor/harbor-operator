package jobservice

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
)

const (
	initImage  = "hairyhenderson/gomplate"
	configPath = "/etc/jobservice/"
	port       = 8080
)

func (r *Reconciler) GetDeployment(ctx context.Context, jobservice *goharborv1alpha2.JobService) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := jobservice.Spec.GetImage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jobservice", jobservice.GetName()),
			Namespace: jobservice.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"jobservice-name":      jobservice.GetName(),
					"jobservice-namespace": jobservice.GetNamespace(),
				},
			},
			Replicas: jobservice.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"jobservice-name":      jobservice.GetName(),
						"jobservice-namespace": jobservice.GetNamespace(),
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 jobservice.Spec.NodeSelector,
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
										Name: jobservice.Name,
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
								}, {
									Name:  "LOG_LEVEL",
									Value: jobservice.Spec.LogLevel,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "jobservice",
							Image: image,
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
											LocalObjectReference: corev1.LocalObjectReference{
												Name: jobservice.Spec.CoreSecret,
											},
											Key:      goharborv1alpha2.CoreSecretKey,
											Optional: &varFalse,
										},
									},
								}, {
									Name:  "JOBSERVICE_WEBHOOK_JOB_MAX_RETRY",
									Value: fmt.Sprintf("%d", jobservice.Spec.WebHookMaxRetry),
								}, {
									Name:  "JOB_SERVICE_POOL_WORKERS",
									Value: fmt.Sprintf("%d", jobservice.Spec.WorkerCount),
								}, {
									Name: "JOBSERVICE_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: jobservice.Name,
											},
											Key:      goharborv1alpha2.JobServiceSecretKey,
											Optional: &varFalse,
										},
									},
								}, {
									Name:  "CORE_URL",
									Value: jobservice.Spec.CoreURL,
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: jobservice.Spec.RedisSecret,
										},
										Optional: &varFalse,
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
					Priority: jobservice.Spec.Priority,
				},
			},
			RevisionHistoryLimit: &revisionHistoryLimit,
		},
	}, nil
}
