package harbor

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers/goharbor/core"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notaryserver"
	"github.com/goharbor/harbor-operator/controllers/goharbor/portal"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
)

func getHostAndIngresses(harbor *goharborv1alpha2.Harbor) (string, []netv1.IngressTLS, error) {
	u, err := url.Parse(harbor.Spec.PublicURL)
	if err != nil {
		return "", nil, errors.Wrap(err, "invalid url")
	}

	var tls []netv1.IngressTLS
	if u.Scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: harbor.Spec.TLSSecretName,
			},
		}
	}

	return strings.SplitN(u.Host, ":", 1)[0], tls, nil
}

func getCoreRules(harbor *goharborv1alpha2.Harbor) []netv1.HTTPIngressPath {
	coreBackend := netv1.IngressBackend{
		ServiceName: harbor.NormalizeComponentName(goharborv1alpha2.CoreName),
		ServicePort: intstr.FromInt(core.PublicPort),
	}

	return []netv1.HTTPIngressPath{
		{
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
		},
	}
}

func (r *Reconciler) GetCoreIngresse(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1.Ingress, error) {
	host, tls, err := getHostAndIngresses(harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get host and ingresses")
	}

	rules := getCoreRules(harbor)

	if harbor.Spec.Components.Portal != nil {
		rules = append(rules, netv1.HTTPIngressPath{
			Path: "/",
			Backend: netv1.IngressBackend{
				ServiceName: harbor.NormalizeComponentName(goharborv1alpha2.PortalName),
				ServicePort: intstr.FromInt(portal.PublicPort),
			},
		})
	}

	if harbor.Spec.Components.Registry != nil {
		rules = append(rules, netv1.HTTPIngressPath{
			Path: "/v2",
			Backend: netv1.IngressBackend{
				ServiceName: harbor.NormalizeComponentName(goharborv1alpha2.RegistryName),
				ServicePort: intstr.FromInt(registry.PublicPort),
			},
		})
	}

	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-harbor-core", harbor.GetName()),
			Namespace: harbor.GetNamespace(),
			Annotations: map[string]string{
				// resolve 413(Too Large Entity) error when push large image. It only works for NGINX ingress.
				"nginx.ingress.kubernetes.io/proxy-body-size": "0",
			},
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: rules,
						},
					},
				},
			},
		},
	}, nil
}

func (r *Reconciler) GetNotaryServerIngresse(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1.Ingress, error) {
	if harbor.Spec.Components.NotaryServer == nil {
		return nil, nil
	}

	u, err := url.Parse(harbor.Spec.Components.NotaryServer.PublicURL)
	if err != nil {
		panic(errors.Wrap(err, "invalid url"))
	}

	notaryHost := strings.SplitN(u.Host, ":", 1)

	var tls []netv1.IngressTLS
	if u.Scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: harbor.Spec.TLSSecretName,
			},
		}
	}

	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-harbor-notary", harbor.GetName()),
			Namespace: harbor.GetNamespace(),
			Annotations: map[string]string{
				// resolve 413(Too Large Entity) error when push large image. It only works for NGINX ingress.
				"nginx.ingress.kubernetes.io/proxy-body-size": "0",
			},
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{
				{
					Host: notaryHost[0],
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path: "/",
									Backend: netv1.IngressBackend{
										ServiceName: harbor.NormalizeComponentName(goharborv1alpha2.NotaryServerName),
										ServicePort: intstr.FromInt(notaryserver.PublicPort),
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
