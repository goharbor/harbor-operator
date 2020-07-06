package jobservice

import (
	"context"
	"fmt"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/pkg/errors"
)

var (
	varFalse = false
)

const (
	VolumeName       = "config"
	LogsVolumeName   = "logs"
	configPath       = "/etc/jobservice/"
	port             = 8080
	HealthPath       = "/api/v1/stats"
	JobLogsParentDir = "/mnt/joblogs"
	LogsParentDir    = "/mnt/logs"
)

func (r *Reconciler) GetDeployment(ctx context.Context, jobservice *goharborv1alpha2.JobService) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, jobservice.GetName())
	namespace := jobservice.GetNamespace()

	volumes := []corev1.Volume{
		{
			Name: VolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
				},
			},
		}, {
			Name: LogsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	volumeMounts := []corev1.VolumeMount{
		{
			MountPath: configPath,
			Name:      VolumeName,
		}, {
			MountPath: logsDirectory,
			Name:      LogsVolumeName,
		},
	}

	for i, fileLogger := range jobservice.Spec.Loggers.Files {
		source := corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}

		if fileLogger.Volume != nil {
			source = *fileLogger.Volume
		}

		name := fmt.Sprintf("logs-%d", i)

		volumes = append(volumes, corev1.Volume{
			Name:         name,
			VolumeSource: source,
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			MountPath: path.Join(LogsParentDir, fmt.Sprintf("%d", i)),
			Name:      name,
			ReadOnly:  false,
		})
	}

	for i, fileLogger := range jobservice.Spec.JobLoggers.Files {
		source := corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}

		if fileLogger.Volume != nil {
			source = *fileLogger.Volume
		}

		name := fmt.Sprintf("joblogs-%d", i)

		volumes = append(volumes, corev1.Volume{
			Name:         name,
			VolumeSource: source,
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			MountPath: path.Join(JobLogsParentDir, fmt.Sprintf("%d", i)),
			Name:      name,
			ReadOnly:  false,
		})
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"):      name,
					r.Label("namespace"): namespace,
				},
			},
			Replicas: jobservice.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 jobservice.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
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
												Name: jobservice.Spec.Core.SecretRef,
											},
											Key:      goharborv1alpha2.SharedSecretKey,
											Optional: &varFalse,
										},
									},
								}, {
									Name:  "REGISTRY_CREDENTIAL_USERNAME",
									Value: jobservice.Spec.Registry.Username,
								}, {
									Name: "REGISTRY_CREDENTIAL_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: jobservice.Spec.Registry.PasswordRef,
											},
											Key:      goharborv1alpha2.SharedSecretKey,
											Optional: &varFalse,
										},
									},
								}, {
									Name: "JOBSERVICE_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: jobservice.Spec.SecretRef,
											},
											Key:      goharborv1alpha2.SharedSecretKey,
											Optional: &varFalse,
										},
									},
								}, {
									Name:  "CORE_URL",
									Value: jobservice.Spec.Core.URL,
								},
							},
							Command:         []string{"/harbor/harbor_jobservice"},
							Args:            []string{"-c", path.Join(configPath, ConfigName)},
							ImagePullPolicy: corev1.PullAlways,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: HealthPath,
										Port: intstr.FromInt(port),
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: HealthPath,
										Port: intstr.FromInt(port),
									},
								},
							},
							VolumeMounts: volumeMounts,
						},
					},
				},
			},
		},
	}, nil
}
