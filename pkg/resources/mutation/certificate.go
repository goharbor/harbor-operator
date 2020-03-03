package mutation

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

type MutateCertificate func(context.Context, *certv1.Certificate, *certv1.Certificate) controllerutil.MutateFn

func NewCertificate(certificate *certv1.Certificate, mutate MutateCertificate) resources.Mutable {
	return func(ctx context.Context, certificateResource, certificateResult runtime.Object) controllerutil.MutateFn {
		result := certificateResult.(*certv1.Certificate)
		previous := certificateResource.(*certv1.Certificate)

		mutate := mutate(ctx, previous, result)

		return func() error {
			previous.DeepCopyInto(result)

			return mutate()
		}
	}
}
