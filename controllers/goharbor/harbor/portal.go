package harbor

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
)

type Portal graph.Resource

func (r *Reconciler) AddPortal(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (PortalInternalCertificate, Portal, error) {
	cert, err := r.AddPortalInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "certificate")
	}

	portal, err := r.GetPortal(ctx, harbor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot get portal")
	}

	portalRes, err := r.AddBasicResource(ctx, portal, cert)

	return cert, portalRes, errors.Wrap(err, "cannot add basic resource")
}

type PortalInternalCertificate graph.Resource

func (r *Reconciler) AddPortalInternalCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (PortalInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, goharborv1alpha2.PortalTLS)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get TLS certificate")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add TLS certificate")
	}

	return PortalInternalCertificate(certRes), nil
}

func (r *Reconciler) GetPortal(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.Portal, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, goharborv1alpha2.PortalTLS))

	return &goharborv1alpha2.Portal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.PortalSpec{
			ComponentSpec: harbor.Spec.Portal,
			TLS:           tls,
		},
	}, nil
}
