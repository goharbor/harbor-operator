package harbor

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	return cert, portalRes, errors.Wrap(err, "cannot add portal")
}

type PortalInternalCertificate graph.Resource

func (r *Reconciler) AddPortalInternalCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (PortalInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.PortalTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return PortalInternalCertificate(certRes), nil
}

func (r *Reconciler) GetPortal(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.Portal, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.PortalTLS))

	return &goharborv1alpha2.Portal{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: version.SetVersion(nil, harbor.Spec.Version),
		},
		Spec: goharborv1alpha2.PortalSpec{
			ComponentSpec: r.getComponentSpec(ctx, harbor, harbormetav1.PortalComponent),
			TLS:           tls,
		},
	}, nil
}
