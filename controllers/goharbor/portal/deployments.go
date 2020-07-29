package portal

import (
	"context"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/pkg/errors"
)

const (
	ConfigVolumeName                      = "config"
	LogsVolumeName                        = "logs"
	ConfigPath                            = "/etc/nginx"
	HealthPath                            = "/"
	JobLogsParentDir                      = "/mnt/joblogs"
	LogsParentDir                         = "/mnt/logs"
	InternalCertificatesVolumeName        = "internal-certificates"
	InternalCertificatesPath              = "/etc/portal/ssl"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
)

const (
	httpsPort = 8443
	httpPort  = 8080
)

var varFalse = false

func (r *Reconciler) GetDeployment(ctx context.Context, portal *goharborv1alpha2.Portal) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, portal.GetName())
	namespace := portal.GetNamespace()

	envs := []corev1.EnvVar{}
	volumes := []corev1.Volume{{
		Name: ConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Optional: &varFalse,
			},
		},
	}}
	volumeMounts := []corev1.VolumeMount{{
		Name:      ConfigVolumeName,
		MountPath: path.Join(ConfigPath, ConfigName),
		SubPath:   ConfigName,
	}}

	if portal.Spec.TLS.Enabled() {
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
					SecretName: portal.Spec.TLS.CertificateRef,
				},
			},
		})
	}

	port := goharborv1alpha2.JobServiceHTTPPortName
	if portal.Spec.TLS.Enabled() {
		port = goharborv1alpha2.JobServiceHTTPSPortName
	}

	httpGET := &corev1.HTTPGetAction{
		Path:   HealthPath,
		Port:   intstr.FromString(port),
		Scheme: portal.Spec.TLS.GetScheme(),
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
			Replicas: portal.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 portal.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					Containers: []corev1.Container{
						{
							Name:  "portal",
							Image: image,
							Ports: []corev1.ContainerPort{{
								Name:          goharborv1alpha2.JobServiceHTTPPortName,
								ContainerPort: httpPort,
							}, {
								Name:          goharborv1alpha2.JobServiceHTTPSPortName,
								ContainerPort: httpsPort,
							}},

							Env:          envs,
							VolumeMounts: volumeMounts,

							ImagePullPolicy: corev1.PullAlways,
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
						},
					},
				},
			},
			Paused: false,
		},
	}, nil
}
