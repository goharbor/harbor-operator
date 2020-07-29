package harbor

import (
	"context"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

const (
	DefaultIngressAnnotationsEnabled   = true
	IngressAnnotationsEnabledCOnfigKey = "ingress-annotations-enabled"
)

type CoreIngress graph.Resource

func (r *Reconciler) AddCoreIngress(ctx context.Context, harbor *goharborv1alpha2.Harbor, core Core, portal Portal, registry Registry) (CoreIngress, error) {
	ingress, err := r.GetCoreIngresse(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get core ingress")
	}

	ingressRes, err := r.Controller.AddIngressToManage(ctx, ingress, core, portal, registry)

	return CoreIngress(ingressRes), errors.Wrap(err, "cannot add core ingress")
}

func (r *Reconciler) GetCoreIngresse(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1.Ingress, error) { // nolint:funlen
	if harbor.Spec.Expose.Ingress == nil {
		return nil, nil
	}

	var tls []netv1.IngressTLS

	if harbor.Spec.Expose.TLS.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: harbor.Spec.Expose.TLS.Core.CertificateRef,
		}}
	}

	portalPort, err := harbor.Spec.InternalTLS.GetInternalPort(goharborv1alpha2.PortalTLS)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get core internal port")
	}

	corePort, err := harbor.Spec.InternalTLS.GetInternalPort(goharborv1alpha2.CoreTLS)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get core internal port")
	}

	coreBackend := netv1.IngressBackend{
		ServiceName: r.NormalizeName(ctx, harbor.GetName(), "core"),
		ServicePort: intstr.FromInt(int(corePort)),
	}

	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.NormalizeName(ctx, harbor.GetName()),
			Namespace:   harbor.GetNamespace(),
			Annotations: r.GetIngressAnnotations(ctx, harbor),
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{{
				Host: harbor.Spec.Expose.Ingress.Hosts.Core,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{{
							Path:    "/api",
							Backend: coreBackend,
						}, {
							Path:    "/c",
							Backend: coreBackend,
						}, {
							Path:    "/chartrepo",
							Backend: coreBackend,
						}, {
							Path:    "/service",
							Backend: coreBackend,
						}, {
							Path: "/",
							Backend: netv1.IngressBackend{
								ServiceName: r.NormalizeName(ctx, harbor.GetName(), "portal"),
								ServicePort: intstr.FromInt(int(portalPort)),
							},
						}, {
							Path:    "/v2",
							Backend: coreBackend,
						}},
					},
				},
			}},
		},
	}, nil
}

type NotaryIngress graph.Resource

func (r *Reconciler) AddNotaryIngress(ctx context.Context, harbor *goharborv1alpha2.Harbor, notary NotaryServer) (NotaryIngress, error) {
	ingress, err := r.GetNotaryServerIngresse(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get notary ingress")
	}

	ingressRes, err := r.Controller.AddIngressToManage(ctx, ingress, notary)

	return NotaryIngress(ingressRes), errors.Wrapf(err, "cannot add notary ingress")
}

func (r *Reconciler) GetNotaryServerIngresse(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1.Ingress, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	if harbor.Spec.Expose.Ingress == nil {
		return nil, nil
	}

	var tls []netv1.IngressTLS

	if harbor.Spec.Expose.TLS.Enabled() && harbor.Spec.Expose.TLS.Notary.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: harbor.Spec.Expose.TLS.Notary.CertificateRef,
		}}
	}

	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.NormalizeName(ctx, harbor.GetName(), "notary"),
			Namespace:   harbor.GetNamespace(),
			Annotations: r.GetIngressAnnotations(ctx, harbor),
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{{
				Host: harbor.Spec.Expose.Ingress.Hosts.Notary,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{{
							Path: "/",
							Backend: netv1.IngressBackend{
								ServiceName: r.NormalizeName(ctx, harbor.GetName(), "notary-server"),
								ServicePort: intstr.FromInt(notaryserver.PublicPort),
							},
						}},
					},
				},
			}},
		},
	}, nil
}

func (r *Reconciler) GetIngressAnnotations(ctx context.Context, harbor *goharborv1alpha2.Harbor) map[string]string {
	// https://github.com/kubernetes/ingress-nginx/blob/master/internal/ingress/annotations/backendprotocol/main.go#L34
	protocol := "HTTP"

	if harbor.Spec.InternalTLS.IsEnabled() {
		protocol = "HTTPS"
	}

	return map[string]string{
		"nginx.ingress.kubernetes.io/backend-protocol": protocol,
		// resolve 413(Too Large Entity) error when push large image. It only works for NGINX ingress.
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	}
}
