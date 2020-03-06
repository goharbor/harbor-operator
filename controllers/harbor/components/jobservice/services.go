package jobservice

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
)

const (
	PublicPort = 80
)

func (j *JobService) GetServices(ctx context.Context) []*corev1.Service {
	operatorName := application.GetName(ctx)
	harborName := j.harbor.Name

	return []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      j.harbor.NormalizeComponentName(goharborv1alpha1.JobServiceName),
				Namespace: j.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.JobServiceName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Port:       PublicPort,
						TargetPort: intstr.FromInt(port),
					},
				},
				Selector: map[string]string{
					"app":      goharborv1alpha1.JobServiceName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
		},
	}
}
