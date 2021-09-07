package harbor

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Portal graph.Resource

func (r *Reconciler) AddPortal(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (PortalInternalCertificate, Portal, error) {
	if harbor.Spec.Portal == nil {
		return nil, nil, nil
	}

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

func (r *Reconciler) AddPortalInternalCertificate(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (PortalInternalCertificate, error) {
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

func (r *Reconciler) GetPortal(ctx context.Context, harbor *goharborv1.Harbor) (*goharborv1.Portal, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.PortalTLS))

	annotation := map[string]string{
		harbormetav1.NetworkPoliciesAnnotationName: harbormetav1.NetworkPoliciesAnnotationDisabled,
	}

	if harbor.Spec.Expose.Core.Ingress != nil {
		annotation[harbormetav1.IngressControllerAnnotationName] = string(harbor.Spec.Expose.Core.Ingress.Controller)
	}

	return &goharborv1.Portal{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: version.SetVersion(annotation, harbor.Spec.Version),
		},
		Spec: goharborv1.PortalSpec{
			ComponentSpec: harbor.GetComponentSpec(ctx, harbormetav1.PortalComponent),
			TLS:           tls,
			Network:       harbor.Spec.Network,
		},
	}, nil
}
