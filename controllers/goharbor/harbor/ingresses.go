package harbor

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers/goharbor/core"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	"github.com/goharbor/harbor-operator/controllers/goharbor/portal"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
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

func getHostAndIngresses(harbor *goharborv1alpha2.Harbor) (string, []netv1.IngressTLS, error) {
	u, err := url.Parse(harbor.Spec.ExternalURL)
	if err != nil {
		return "", nil, errors.Wrap(err, "invalid url")
	}

	var tls []netv1.IngressTLS

	if harbor.Spec.Expose.TLS != nil {
		tls = []netv1.IngressTLS{
			{
				SecretName: harbor.Spec.Expose.TLS.CertificateRef,
			},
		}
	}

	return u.Hostname(), tls, nil
}

func (r *Reconciler) GetCoreIngresse(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1.Ingress, error) {
	if harbor.Spec.Expose.Ingress == nil {
		return nil, nil
	}

	host, tls, err := getHostAndIngresses(harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get host and ingresses")
	}

	coreBackend := netv1.IngressBackend{
		ServiceName: r.NormalizeName(ctx, harbor.GetName(), "core"),
		ServicePort: intstr.FromInt(core.PublicPort),
	}
	portalBackend := netv1.IngressBackend{
		ServiceName: r.NormalizeName(ctx, harbor.GetName(), "portal"),
		ServicePort: intstr.FromInt(portal.PublicPort),
	}
	registryBackend := netv1.IngressBackend{
		ServiceName: r.NormalizeName(ctx, harbor.GetName(), "registry"),
		ServicePort: intstr.FromInt(registry.PublicPort),
	}

	rules := []netv1.HTTPIngressPath{{
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
		Path:    "/",
		Backend: portalBackend,
	}, {
		Path:    "/v2",
		Backend: registryBackend,
	}}

	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-harbor-core", harbor.GetName()),
			Namespace:   harbor.GetNamespace(),
			Annotations: r.GetIngressAnnotations(),
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{{
				Host: host,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: rules,
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

	u, err := url.Parse(harbor.Spec.Expose.Ingress.Hosts.Notary)
	if err != nil {
		return nil, errors.Wrap(err, "invalid url")
	}

	var tls []netv1.IngressTLS

	if harbor.Spec.Expose.TLS != nil {
		tls = []netv1.IngressTLS{{
			SecretName: harbor.Spec.Expose.TLS.CertificateRef,
		}}
	}

	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-harbor-notary", harbor.GetName()),
			Namespace:   harbor.GetNamespace(),
			Annotations: r.GetIngressAnnotations(),
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{{
				Host: u.Hostname(),
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

func (r *Reconciler) GetIngressAnnotations() map[string]string {
	return map[string]string{
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	}
}
