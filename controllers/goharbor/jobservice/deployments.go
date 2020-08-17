package jobservice

import (
	"context"
	"fmt"
	"path"
	"strconv"

	"github.com/goharbor/harbor/src/common"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config/harbor"
)

const (
	ConfigVolumeName                      = "config"
	LogsVolumeName                        = "logs"
	ConfigPath                            = "/etc/jobservice"
	HealthPath                            = "/api/v1/stats"
	JobLogsParentDir                      = "/mnt/joblogs"
	LogsParentDir                         = "/mnt/logs"
	InternalCertificatesVolumeName        = "internal-certificates"
	InternalCertificatesPath              = ConfigPath + "/ssl"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
)

var varFalse = false

const (
	httpsPort = 8443
	httpPort  = 8080
)

func (r *Reconciler) GetDeployment(ctx context.Context, jobservice *goharborv1alpha2.JobService) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, jobservice.GetName())
	namespace := jobservice.GetNamespace()

	envs, err := harbor.EnvVars(map[string]harbor.ConfigValue{
		common.RegistryControllerURL: harbor.Value(jobservice.Spec.Registry.ControllerURL),
		common.RegistryURL:           harbor.Value(jobservice.Spec.Registry.RegistryURL),
		common.CoreURL:               harbor.Value(jobservice.Spec.Core.URL),
		common.TokenServiceURL:       harbor.Value(jobservice.Spec.TokenService.URL),
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot configure environment variables")
	}

	envs = append(envs, corev1.EnvVar{
		Name: "CORE_SECRET",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: jobservice.Spec.Core.SecretRef,
				},
				Key:      harbormetav1.SharedSecretKey,
				Optional: &varFalse,
			},
		},
	}, corev1.EnvVar{
		Name:  "REGISTRY_CREDENTIAL_USERNAME",
		Value: jobservice.Spec.Registry.Credentials.Username,
	}, corev1.EnvVar{
		Name: "REGISTRY_CREDENTIAL_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: jobservice.Spec.Registry.Credentials.PasswordRef,
				},
				Key:      harbormetav1.SharedSecretKey,
				Optional: &varFalse,
			},
		},
	}, corev1.EnvVar{
		Name: "JOBSERVICE_SECRET",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: jobservice.Spec.SecretRef,
				},
				Key:      harbormetav1.SharedSecretKey,
				Optional: &varFalse,
			},
		},
	}, corev1.EnvVar{
		Name:  "INTERNAL_TLS_ENABLED",
		Value: strconv.FormatBool(jobservice.Spec.TLS.Enabled()),
	})
	volumes := []corev1.Volume{{
		Name: ConfigVolumeName,
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
	}}
	volumeMounts := []corev1.VolumeMount{{
		MountPath: ConfigPath,
		Name:      ConfigVolumeName,
	}, {
		MountPath: logsDirectory,
		Name:      LogsVolumeName,
	}}

	if jobservice.Spec.TLS.Enabled() {
		envs = append(envs, corev1.EnvVar{
			Name:  "INTERNAL_TLS_TRUST_CA_PATH",
			Value: path.Join(InternalCertificateAuthorityDirectory, corev1.ServiceAccountRootCAKey),
		}, corev1.EnvVar{
			Name:  "INTERNAL_TLS_CERT_PATH",
			Value: path.Join(InternalCertificatesPath, corev1.TLSCertKey),
		}, corev1.EnvVar{
			Name:  "INTERNAL_TLS_KEY_PATH",
			Value: path.Join(InternalCertificatesPath, corev1.TLSPrivateKeyKey),
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: path.Join(InternalCertificateAuthorityDirectory, corev1.ServiceAccountRootCAKey),
			SubPath:   corev1.ServiceAccountRootCAKey,
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: InternalCertificatesPath,
			ReadOnly:  true,
		})

		volumes = append(volumes, corev1.Volume{
			Name: InternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: jobservice.Spec.TLS.CertificateRef,
				},
			},
		})
	} else {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: InternalCertificateAuthorityDirectory,
		})

		volumes = append(volumes, corev1.Volume{
			Name: InternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
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

	port := harbormetav1.JobServiceHTTPPortName
	if jobservice.Spec.TLS.Enabled() {
		port = harbormetav1.JobServiceHTTPSPortName
	}

	httpGET := &corev1.HTTPGetAction{
		Path:   HealthPath,
		Port:   intstr.FromString(port),
		Scheme: jobservice.Spec.TLS.GetScheme(),
	}

	deploy := &appsv1.Deployment{
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
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					Containers: []corev1.Container{{
						Name:  controllers.JobService.String(),
						Image: image,
						Ports: []corev1.ContainerPort{{
							Name:          harbormetav1.JobServiceHTTPPortName,
							ContainerPort: httpPort,
						}, {
							Name:          harbormetav1.JobServiceHTTPSPortName,
							ContainerPort: httpsPort,
						}},

						// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/jobservice/env.jinja
						Env: envs,
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: httpGET,
							},
						},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: httpGET,
							},
						},
						VolumeMounts: volumeMounts,
					}},
				},
			},
		},
	}

	jobservice.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, nil
}
