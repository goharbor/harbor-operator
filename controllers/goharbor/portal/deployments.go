package portal

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/pkg/errors"
)

const (
	port = 8080
)

var (
	varFalse = false
)

func (r *Reconciler) GetDeployment(ctx context.Context, portal *goharborv1alpha2.Portal) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, portal.GetName())
	namespace := portal.GetNamespace()

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-portal", portal.GetName()),
			Namespace: portal.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"portal.goharbor.io/name":      name,
					"portal.goharbor.io/namespace": namespace,
				},
			},
			Replicas: portal.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"portal.goharbor.io/name":      name,
						"portal.goharbor.io/namespace": namespace,
					},
					Annotations: map[string]string{
						"registry.goharbor.io/uid":        fmt.Sprintf("%v", portal.GetUID()),
						"registry.goharbor.io/generation": fmt.Sprintf("%v", portal.GetGeneration()),
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 portal.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Containers: []corev1.Container{
						{
							Name:  "portal",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
								},
							},

							ImagePullPolicy: corev1.PullAlways,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(port),
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(port),
									},
								},
							},
						},
					},
					Priority: portal.Spec.Priority,
				},
			},
			Paused: false,
		},
	}, nil
}
