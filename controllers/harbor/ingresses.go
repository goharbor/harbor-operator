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

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers/core"
	"github.com/goharbor/harbor-operator/controllers/notaryserver"
	"github.com/goharbor/harbor-operator/controllers/portal"
	"github.com/goharbor/harbor-operator/controllers/registry"
)

func (r *Reconciler) GetCoreIngresse(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*netv1.Ingress, error) {
	u, err := url.Parse(harbor.Spec.PublicURL)
	if err != nil {
		panic(errors.Wrap(err, "invalid url"))
	}

	host := strings.SplitN(u.Host, ":", 1)

	var tls []netv1.IngressTLS
	if u.Scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: harbor.Spec.TLSSecretName,
			},
		}
	}

	rules := []netv1.HTTPIngressPath{
		{
			Path: "/api",
			Backend: netv1.IngressBackend{
				ServiceName: harbor.NormalizeComponentName(goharborv1alpha2.CoreName),
				ServicePort: intstr.FromInt(core.PublicPort),
			},
		}, {
			Path: "/c",
			Backend: netv1.IngressBackend{
				ServiceName: harbor.NormalizeComponentName(goharborv1alpha2.CoreName),
				ServicePort: intstr.FromInt(core.PublicPort),
			},
		}, {
			Path: "/chartrepo",
			Backend: netv1.IngressBackend{
				ServiceName: harbor.NormalizeComponentName(goharborv1alpha2.CoreName),
				ServicePort: intstr.FromInt(core.PublicPort),
			},
		}, {
			Path: "/service",
			Backend: netv1.IngressBackend{
				ServiceName: harbor.NormalizeComponentName(goharborv1alpha2.CoreName),
				ServicePort: intstr.FromInt(core.PublicPort),
			},
		},
	}

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
		},
		Spec: netv1.IngressSpec{
			TLS: tls,
			Rules: []netv1.IngressRule{
				{
					Host: host[0],
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
