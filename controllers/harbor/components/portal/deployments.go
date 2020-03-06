package portal

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/goharbor/harbor-core-operator/api/v1alpha1"
	"github.com/goharbor/harbor-core-operator/pkg/factories/application"
)

const (
	port = 8080
)

var (
	revisionHistoryLimit int32 = 0 // nolint:golint
	varFalse                   = false
)

func (p *Portal) GetDeployments(ctx context.Context) []*appsv1.Deployment { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := p.harbor.GetName()

	return []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.harbor.NormalizeComponentName(containerregistryv1alpha1.PortalName),
				Namespace: p.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.PortalName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":      containerregistryv1alpha1.PortalName,
						"harbor":   harborName,
						"operator": operatorName,
					},
				},
				Replicas: p.harbor.Spec.Components.Portal.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"configuration/checksum": "",
							"secret/checksum":        "",
							"operator/version":       application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      containerregistryv1alpha1.PortalName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						NodeSelector:                 p.harbor.Spec.Components.Portal.NodeSelector,
						AutomountServiceAccountToken: &varFalse,
						Containers: []corev1.Container{
							{
								Name:  "portal",
								Image: p.harbor.Spec.Components.Portal.GetImage(),
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
						Priority: p.Option.GetPriority(),
					},
				},
				RevisionHistoryLimit: &revisionHistoryLimit,
				Paused:               p.harbor.Spec.Paused,
			},
		},
	}
}
