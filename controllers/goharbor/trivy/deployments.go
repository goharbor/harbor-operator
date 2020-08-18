package trivy

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

var (
	varFalse = false
)

const (
	ContainerName                  = "trivy"
	LivenessProbe                  = "/probe/healthy"
	ReadinessProbe                 = "/probe/ready"
	CacheVolumeName                = "cache"
	CacheVolumePath                = "/home/scanner/.cache/trivy"
	ReportsVolumeName              = "reports"
	ReportsVolumePath              = "/home/scanner/.cache/reports"
	InternalCertificatesVolumeName = "internal-certificates"
	InternalCertificatesPath       = "/etc/harbor/ssl"
)

const (
	httpsPort = 8443
	httpPort  = 8080
)

func (r *Reconciler) AddDeployment(ctx context.Context, trivy *goharborv1alpha2.Trivy, dependencies ...graph.Resource) error {
	// Forge the deploy resource
	deploy, err := r.GetDeployment(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "cannot get deployment")
	}

	// Add deploy to reconciler controller
	_, err = r.Controller.AddDeploymentToManage(ctx, deploy, dependencies...)
	if err != nil {
		return errors.Wrapf(err, "cannot manage deploy %s", deploy.GetName())
	}

	return nil
}

func (r *Reconciler) GetDeployment(ctx context.Context, trivy *goharborv1alpha2.Trivy) (*appsv1.Deployment, error) { // nolint:funlen
	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()

	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get image for deploy: %s", name)
	}

	volumes := []corev1.Volume{{
		Name:         ReportsVolumeName,
		VolumeSource: trivy.Spec.Storage.Reports.VolumeSource,
	}, {
		Name:         CacheVolumeName,
		VolumeSource: trivy.Spec.Storage.Cache.VolumeSource,
	}}
	volumesMount := []corev1.VolumeMount{{
		Name:      ReportsVolumeName,
		MountPath: ReportsVolumePath,
		ReadOnly:  false,
	}, {
		Name:      CacheVolumeName,
		MountPath: CacheVolumePath,
		ReadOnly:  false,
	}}
	envs := []corev1.EnvVar{}
	envFroms := []corev1.EnvFromSource{{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}, {
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}}

	if trivy.Spec.Update.GithubTokenRef != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "SCANNER_TRIVY_GITHUB_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: trivy.Spec.Update.GithubTokenRef,
					},
					Key: harbormetav1.GithubTokenPasswordKey,
				},
			},
		})
	}

	address := fmt.Sprintf(":%d", httpPort)

	if trivy.Spec.Server.TLS.Enabled() {
		volumes = append(volumes, corev1.Volume{
			Name: InternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: trivy.Spec.Server.TLS.CertificateRef,
				},
			},
		})

		volumesMount = append(volumesMount, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: InternalCertificatesPath,
			ReadOnly:  true,
		})

		envs = append(envs, corev1.EnvVar{
			Name:  "SCANNER_API_SERVER_TLS_CERTIFICATE",
			Value: path.Join(InternalCertificatesPath, corev1.TLSCertKey),
		}, corev1.EnvVar{
			Name:  "SCANNER_API_SERVER_TLS_KEY",
			Value: path.Join(InternalCertificatesPath, corev1.TLSPrivateKeyKey),
		}, corev1.EnvVar{
			Name:  "SCANNER_API_SERVER_CLIENT_CAS",
			Value: strings.Join(trivy.Spec.Server.ClientCertificateAuthorityRefs, ","),
		})

		address = fmt.Sprintf(":%d", httpsPort)
	}

	envs = append(envs, corev1.EnvVar{
		Name:  "SCANNER_API_SERVER_ADDR",
		Value: address,
	})

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
			Replicas: trivy.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 trivy.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,

					Containers: []corev1.Container{{
						Name:  ContainerName,
						Image: image,
						Ports: []corev1.ContainerPort{{
							Name:          harbormetav1.TrivyHTTPPortName,
							ContainerPort: httpPort,
						}, {
							Name:          harbormetav1.TrivyHTTPSPortName,
							ContainerPort: httpsPort,
						}},

						Env:          envs,
						EnvFrom:      envFroms,
						VolumeMounts: volumesMount,

						LivenessProbe:  r.getProbe(ctx, trivy, LivenessProbe),
						ReadinessProbe: r.getProbe(ctx, trivy, ReadinessProbe),
					}},
				},
			},
		},
	}, nil
}

func (r *Reconciler) getProbe(ctx context.Context, trivy *goharborv1alpha2.Trivy, probePath string) *corev1.Probe {
	if trivy.Spec.Server.TLS.Enabled() {
		hostname := r.NormalizeName(ctx, trivy.GetName())

		return &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"curl", "-sSL",
						"--resolve", fmt.Sprintf("%s:%d:%s", hostname, httpsPort, "127.0.0.1"),
						"--cacert", path.Join(InternalCertificatesPath, corev1.ServiceAccountRootCAKey),
						"--key", path.Join(InternalCertificatesPath, corev1.TLSPrivateKeyKey),
						"--cert", path.Join(InternalCertificatesPath, corev1.TLSCertKey),
						(&url.URL{
							Scheme: "https",
							Host:   fmt.Sprintf("%s:%d", hostname, httpsPort),
							Path:   probePath,
						}).String(),
					},
				},
			},
		}
	}

	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: probePath,
				Port: intstr.FromInt(httpPort),
			},
		},
	}
}
