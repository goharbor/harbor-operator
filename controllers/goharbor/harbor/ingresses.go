package harbor

import (
	"context"
	"fmt"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultIngressAnnotationsEnabled   = true
	IngressAnnotationsEnabledCOnfigKey = "ingress-annotations-enabled"
)

type CoreIngress graph.Resource

func (r *Reconciler) AddCoreIngress(ctx context.Context, harbor *goharborv1alpha2.Harbor, core Core, portal Portal) (CoreIngress, error) {
	ingress, err := r.GetCoreIngress(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get core ingress")
	}

	ingressRes, err := r.Controller.AddIngressToManage(ctx, ingress, core, portal)

	return CoreIngress(ingressRes), errors.Wrap(err, "cannot add core ingress")
}

func (r *Reconciler) GetCoreIngress(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1.Ingress, error) {
	if harbor.Spec.Expose.Core.Ingress == nil {
		return nil, nil
	}

	var tls []netv1.IngressTLS

	if harbor.Spec.Expose.Core.TLS.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: harbor.Spec.Expose.Core.TLS.CertificateRef,
		}}
	}

	rules, err := r.GetCoreIngressRules(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "ingress rules")
	}

	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.NormalizeName(ctx, harbor.GetName()),
			Namespace:   harbor.GetNamespace(),
			Annotations: r.GetCoreIngressAnnotations(ctx, harbor),
		},
		Spec: netv1.IngressSpec{
			TLS:   tls,
			Rules: rules,
		},
	}, nil
}

func (r *Reconciler) GetCoreIngressRules(ctx context.Context, harbor *goharborv1alpha2.Harbor) ([]netv1.IngressRule, error) {
	corePort, err := harbor.Spec.InternalTLS.GetInternalPort(harbormetav1.CoreTLS)
	if err != nil {
		return nil, errors.Wrapf(err, "%s internal port", harbormetav1.CoreTLS)
	}

	coreBackend := netv1.IngressBackend{
		Service: &netv1.IngressServiceBackend{
			Name: r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String()),
			Port: netv1.ServiceBackendPort{
				Number: int32(corePort),
			},
		},
	}

	portalPort, err := harbor.Spec.InternalTLS.GetInternalPort(harbormetav1.PortalTLS)
	if err != nil {
		return nil, errors.Wrapf(err, "%s internal port", harbormetav1.PortalTLS)
	}

	portalBackend := netv1.IngressBackend{
		Service: &netv1.IngressServiceBackend{
			Name: r.NormalizeName(ctx, harbor.GetName(), "portal"),
			Port: netv1.ServiceBackendPort{
				Number: int32(portalPort),
			},
		},
	}

	ruleValue, err := r.GetCoreIngressRuleValue(ctx, harbor, coreBackend, portalBackend)
	if err != nil {
		return nil, errors.Wrap(err, "rule value")
	}

	return []netv1.IngressRule{{
		Host:             harbor.Spec.Expose.Core.Ingress.Host,
		IngressRuleValue: *ruleValue,
	}}, nil
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

	if harbor.Spec.Expose.Notary.Ingress == nil {
		return nil, nil
	}

	var tls []netv1.IngressTLS

	if harbor.Spec.Expose.Notary.TLS.Enabled() {
		tls = []netv1.IngressTLS{{
			SecretName: harbor.Spec.Expose.Notary.TLS.CertificateRef,
		}}
	}

	pathTypeExact := netv1.PathTypeExact

	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String()),
			Namespace:   harbor.GetNamespace(),
			Annotations: r.GetNotaryIngressAnnotations(ctx, harbor),
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{{
				Host: harbor.Spec.Expose.Notary.Ingress.Host,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathTypeExact,
							Backend: netv1.IngressBackend{
								Service: &netv1.IngressServiceBackend{
									Name: r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String()),
									Port: netv1.ServiceBackendPort{
										Number: notaryserver.PublicPort,
									},
								},
							},
						}},
					},
				},
			}},
		},
	}, nil
}

func (r *Reconciler) GetCoreIngressAnnotations(ctx context.Context, harbor *goharborv1alpha2.Harbor) map[string]string {
	// https://github.com/kubernetes/ingress-nginx/blob/master/internal/ingress/annotations/backendprotocol/main.go#L34
	protocol := "HTTP"

	if harbor.Spec.InternalTLS.IsEnabled() {
		protocol = "HTTPS"
	}

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/backend-protocol": protocol,
		// resolve 413(Too Large Entity) error when push large image. It only works for NGINX ingress.
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	}

	for key, value := range harbor.Spec.Expose.Core.Ingress.Annotations {
		annotations[key] = value
	}

	return annotations
}

func (r *Reconciler) GetNotaryIngressAnnotations(ctx context.Context, harbor *goharborv1alpha2.Harbor) map[string]string {
	// https://github.com/kubernetes/ingress-nginx/blob/master/internal/ingress/annotations/backendprotocol/main.go#L34
	protocol := "HTTP"

	if harbor.Spec.InternalTLS.IsEnabled() {
		protocol = "HTTPS"
	}

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/backend-protocol": protocol,
		// resolve 413(Too Large Entity) error when push large image. It only works for NGINX ingress.
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	}

	for key, value := range harbor.Spec.Expose.Notary.Ingress.Annotations {
		annotations[key] = value
	}

	return annotations
}

type ErrInvalidIngressController struct {
	Controller harbormetav1.IngressController
}

func (err ErrInvalidIngressController) Error() string {
	return fmt.Sprintf("controller %s unsupported", err.Controller)
}

func (r *Reconciler) GetCoreIngressRuleValue(ctx context.Context, harbor *goharborv1alpha2.Harbor, core, portal netv1.IngressBackend) (*netv1.IngressRuleValue, error) { // nolint:funlen
	pathTypeExact := netv1.PathTypeExact

	switch harbor.Spec.Expose.Core.Ingress.Controller {
	case harbormetav1.IngressControllerDefault:
		return &netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{{
					Path:     "/",
					PathType: &pathTypeExact,
					Backend:  portal,
				}, {
					Path:     "/api/",
					PathType: &pathTypeExact,
					Backend:  core,
				}, {
					Path:     "/service/",
					PathType: &pathTypeExact,
					Backend:  core,
				}, {
					Path:     "/v2/",
					PathType: &pathTypeExact,
					Backend:  core,
				}, {
					Path:     "/chartrepo/",
					PathType: &pathTypeExact,
					Backend:  core,
				}, {
					Path:     "/c/",
					PathType: &pathTypeExact,
					Backend:  core,
				}},
			},
		}, nil
	case harbormetav1.IngressControllerGCE:
		pathTypePrefix := netv1.PathTypePrefix

		return &netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{{
					Path:     "/",
					PathType: &pathTypePrefix,
					Backend:  portal,
				}, {
					Path:     "/api",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/service",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/v2",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/chartrepo",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/c",
					PathType: &pathTypePrefix,
					Backend:  core,
				}},
			},
		}, nil
	case harbormetav1.IngressControllerNCP:
		pathTypeExact := netv1.PathTypeExact
		pathTypePrefix := netv1.PathTypePrefix

		return &netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{{
					Path:     "/",
					PathType: &pathTypeExact,
					Backend:  portal,
				}, {
					Path:     "/api",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/service",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/v2",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/chartrepo",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/c",
					PathType: &pathTypePrefix,
					Backend:  core,
				}},
			},
		}, nil
	default:
		return nil, ErrInvalidIngressController{harbor.Spec.Expose.Core.Ingress.Controller}
	}
}
