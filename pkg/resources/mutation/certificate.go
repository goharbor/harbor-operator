package mutation

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

func NewCertificate(mutate resources.Mutable) resources.Mutable {
	return func(ctx context.Context, certificateResource, certificateResult runtime.Object) controllerutil.MutateFn {
		result := certificateResult.(*certv1.Certificate)
		desired := certificateResource.(*certv1.Certificate)

		mutate := mutate(ctx, desired, result)

		return func() error {
			desired.Spec.DeepCopyInto(&result.Spec)

			return mutate()
		}
	}
}
