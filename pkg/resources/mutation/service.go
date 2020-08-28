package mutation

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NewService(mutate resources.Mutable) (result resources.Mutable) {
	result = func(ctx context.Context, serviceResource, serviceResult runtime.Object) controllerutil.MutateFn {
		result := serviceResult.(*corev1.Service)
		desired := serviceResource.(*corev1.Service)

		mutate := mutate(ctx, desired, result)

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

			desired.Spec.DeepCopyInto(&result.Spec)

			return mutate()
		}
	}

	result.AppendMutation(MetadataMutateFn)

	return result
}
