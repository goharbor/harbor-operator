package trivy

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/pkg/errors"
)

var (
	varFalse = false
	varTrue  = true
)

const (
	HealthPath        = "/health"
	port              = 8080 // https://github.com/helm/chartmuseum/blob/969515a51413e1f1840fb99509401aa3c63deccd/pkg/config/vars.go#L135
	CacheVolumeName   = "cache"
	CacheVolumePath   = "/home/scanner/.cache/trivy"
	ReportsVolumeName = "reports"
	ReportsVolumePath = "/home/scanner/.cache/reports"
)

func (r *Reconciler) GetDeployment(ctx context.Context, trivy *goharborv1alpha2.Trivy) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()

	volumes := []corev1.Volume{
		{
			Name: CacheVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: ReportsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	volumesMount := []corev1.VolumeMount{
		{
			Name:      CacheVolumeName,
			MountPath: CacheVolumePath,
		},
		{
			Name:      ReportsVolumeName,
			MountPath: ReportsVolumePath,
		},
	}

	envFroms := []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
			},
		},
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

					Containers: []corev1.Container{
						{
							Name:  "trivy",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
								},
							},

							EnvFrom: envFroms,

							VolumeMounts: volumesMount,

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
						},
					},
				},
			},
		},
	}, nil
}
