package mutation

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

type MutateService func(context.Context, *corev1.Service, *corev1.Service) controllerutil.MutateFn

func NewService(service *corev1.Service, mutate MutateService) resources.Mutable {
	return func(ctx context.Context, serviceResource, serviceResult runtime.Object) controllerutil.MutateFn {
		result := serviceResult.(*corev1.Service)
		previous := serviceResource.(*corev1.Service)

		mutate := mutate(ctx, previous, result)

		return func() error {
			// Immutable field
			clusterIP := result.Spec.ClusterIP

			defer func() { result.Spec.ClusterIP = clusterIP }()

			for _, port := range result.Spec.Ports {
				port := port

				defer func() {
					ports := make([]corev1.ServicePort, len(result.Spec.Ports))

					for i, p := range result.Spec.Ports {
						if p.Name == port.Name {
							p.NodePort = port.NodePort
						}

						ports[i] = p
					}

					result.Spec.Ports = ports
				}()
			}

			previous.DeepCopyInto(result)

			return mutate()
		}
	}
}
