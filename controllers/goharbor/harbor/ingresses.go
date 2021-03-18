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
	netv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	DefaultIngressAnnotationsEnabled   = true
	IngressAnnotationsEnabledCOnfigKey = "ingress-annotations-enabled"
	NCPIngressValueTrue                = "true"
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

func (r *Reconciler) GetCoreIngress(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1beta1.Ingress, error) {
	if harbor.Spec.Expose.Core.Ingress == nil {
		return nil, nil
	}

	var tls []netv1beta1.IngressTLS

	if harbor.Spec.Expose.Core.TLS.Enabled() {
		tls = []netv1beta1.IngressTLS{{
			SecretName: harbor.Spec.Expose.Core.TLS.CertificateRef,
		}}
	}

	rules, err := r.GetCoreIngressRules(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "ingress rules")
	}

	return &netv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.NormalizeName(ctx, harbor.GetName()),
			Namespace:   harbor.GetNamespace(),
			Annotations: r.GetCoreIngressAnnotations(ctx, harbor),
		},
		Spec: netv1beta1.IngressSpec{
			TLS:   tls,
			Rules: rules,
		},
	}, nil
}

func (r *Reconciler) GetCoreIngressRules(ctx context.Context, harbor *goharborv1alpha2.Harbor) ([]netv1beta1.IngressRule, error) {
	corePort, err := harbor.Spec.InternalTLS.GetInternalPort(harbormetav1.CoreTLS)
	if err != nil {
		return nil, errors.Wrapf(err, "%s internal port", harbormetav1.CoreTLS)
	}

	coreBackend := netv1beta1.IngressBackend{
		ServiceName: r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String()),
		ServicePort: intstr.FromInt(int(corePort)),
	}

	portalPort, err := harbor.Spec.InternalTLS.GetInternalPort(harbormetav1.PortalTLS)
	if err != nil {
		return nil, errors.Wrapf(err, "%s internal port", harbormetav1.PortalTLS)
	}

	portalBackend := netv1beta1.IngressBackend{
		ServiceName: r.NormalizeName(ctx, harbor.GetName(), "portal"),
		ServicePort: intstr.FromInt(int(portalPort)),
	}

	ruleValue, err := r.GetCoreIngressRuleValue(ctx, harbor, coreBackend, portalBackend)
	if err != nil {
		return nil, errors.Wrap(err, "rule value")
	}

	return []netv1beta1.IngressRule{{
		Host:             harbor.Spec.Expose.Core.Ingress.Host,
		IngressRuleValue: *ruleValue,
	}}, nil
}

type NotaryIngress graph.Resource

func (r *Reconciler) AddNotaryIngress(ctx context.Context, harbor *goharborv1alpha2.Harbor, notary NotaryServer) (NotaryIngress, error) {
	ingress, err := r.GetNotaryServerIngress(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get notary ingress")
	}

	ingressRes, err := r.Controller.AddIngressToManage(ctx, ingress, notary)

	return NotaryIngress(ingressRes), errors.Wrapf(err, "cannot add notary ingress")
}

func (r *Reconciler) GetNotaryServerIngress(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1.Ingress, error) {
	if harbor.Spec.Notary == nil {
		return nil, nil
	}

	if harbor.Spec.Expose.Notary.Ingress == nil {
		return nil, nil
	}

	var tls []netv1beta1.IngressTLS

	if harbor.Spec.Expose.Notary.TLS.Enabled() {
		tls = []netv1beta1.IngressTLS{{
			SecretName: harbor.Spec.Expose.Notary.TLS.CertificateRef,
		}}
	}

	ingressRules, err := r.GetNotaryIngressRules(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get notary ingress rules")
	}

	return &netv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String()),
			Namespace:   harbor.GetNamespace(),
			Annotations: r.GetNotaryIngressAnnotations(ctx, harbor),
		},
		Spec: netv1beta1.IngressSpec{
			TLS:   tls,
			Rules: ingressRules,
		},
	}, nil
}

func (r *Reconciler) GetNotaryIngressRules(ctx context.Context, harbor *goharborv1alpha2.Harbor) ([]netv1beta1.IngressRule, error) {
	var ingressRuleValue netv1beta1.IngressRuleValue

	backend := netv1beta1.IngressBackend{
		ServiceName: r.NormalizeName(ctx, harbor.GetName(), controllers.NotaryServer.String()),
		ServicePort: intstr.FromInt(notaryserver.PublicPort),
	}

	pathTypeExact := netv1beta1.PathTypeExact

	switch harbor.Spec.Expose.Core.Ingress.Controller {
	case harbormetav1.IngressControllerDefault:
		ingressRuleValue = netv1beta1.IngressRuleValue{
			HTTP: &netv1beta1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{
					{
						Path:    "/",
						PathType: &pathTypeExact,
						Backend: backend,
					},
				},
			},
		}
	case harbormetav1.IngressControllerGCE:
		ingressRuleValue = netv1beta1.IngressRuleValue{
			HTTP: &netv1beta1.HTTPIngressRuleValue{
				Paths: []netv1beta1.HTTPIngressPath{
					{
						Path:    "/",
						PathType: &pathTypeExact,
						Backend: backend,
					},
				},
			},
		}
	case harbormetav1.IngressControllerNCP:
		ingressRuleValue = netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{
					{
						Path:    "/.*",
						Backend: backend,
					},
				},
			},
		}
	default:
		return nil, ErrInvalidIngressController{Controller: harbor.Spec.Expose.Core.Ingress.Controller}
	}

	return []netv1.IngressRule{
		{
			Host:             harbor.Spec.Expose.Notary.Ingress.Host,
			IngressRuleValue: ingressRuleValue,
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
	if harbor.Spec.Expose.Core.Ingress.Controller == harbormetav1.IngressControllerNCP {
		annotations["ncp/use-regex"] = NCPIngressValueTrue
		if harbor.Spec.InternalTLS.IsEnabled() {
			annotations["ncp/http-redirect"] = NCPIngressValueTrue
		}
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
	if harbor.Spec.Expose.Core.Ingress.Controller == harbormetav1.IngressControllerNCP {
		annotations["ncp/use-regex"] = NCPIngressValueTrue
		if harbor.Spec.InternalTLS.IsEnabled() {
			annotations["ncp/http-redirect"] = NCPIngressValueTrue
		}
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

func (r *Reconciler) GetCoreIngressRuleValue(ctx context.Context, harbor *goharborv1alpha2.Harbor, core, portal netv1beta1.IngressBackend) (*netv1beta1.IngressRuleValue, error) { // nolint:funlen
	pathTypePrefix := netv1beta1.PathTypePrefix

	switch harbor.Spec.Expose.Core.Ingress.Controller {
	case harbormetav1.IngressControllerDefault:
		return &netv1beta1.IngressRuleValue{
			HTTP: &netv1beta1.HTTPIngressRuleValue{
				Paths: []netv1beta1.HTTPIngressPath{{
					Path:     "/",
					PathType: &pathTypePrefix,
					Backend:  portal,
				}, {
					Path:     "/api/",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/service/",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/v2/",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/chartrepo/",
					PathType: &pathTypePrefix,
					Backend:  core,
				}, {
					Path:     "/c/",
					PathType: &pathTypePrefix,
					Backend:  core,
				}},
			},
		}, nil
	case harbormetav1.IngressControllerGCE:
		pathTypePrefix := netv1beta1.PathTypePrefix

		return &netv1beta1.IngressRuleValue{
			HTTP: &netv1beta1.HTTPIngressRuleValue{
				Paths: []netv1beta1.HTTPIngressPath{{
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
		pathTypeExact := netv1beta1.PathTypeExact
		pathTypePrefix := netv1beta1.PathTypePrefix

		return &netv1beta1.IngressRuleValue{
			HTTP: &netv1beta1.HTTPIngressRuleValue{
				Paths: []netv1beta1.HTTPIngressPath{{
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
