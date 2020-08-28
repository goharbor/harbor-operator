package mutation

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/resources"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NewCertificate(mutate resources.Mutable) (result resources.Mutable) {
	result = func(ctx context.Context, certificateResource, certificateResult runtime.Object) controllerutil.MutateFn {
		result := certificateResult.(*certv1.Certificate)
		desired := certificateResource.(*certv1.Certificate)

		mutate := mutate(ctx, desired, result)

		return func() error {
			desired.Spec.DeepCopyInto(&result.Spec)

			return mutate()
		}
	}

	result.AppendMutation(MetadataMutateFn)

	return result
}

func NewIssuer(mutate resources.Mutable) (result resources.Mutable) {
	result = func(ctx context.Context, issuerResource, issuerResult runtime.Object) controllerutil.MutateFn {
		result := issuerResult.(*certv1.Issuer)
		desired := issuerResource.(*certv1.Issuer)

		mutate := mutate(ctx, desired, result)

		return func() error {
			desired.Spec.DeepCopyInto(&result.Spec)

			return mutate()
		}
	}

	result.AppendMutation(MetadataMutateFn)

	return result
}
