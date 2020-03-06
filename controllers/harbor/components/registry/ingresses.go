package registry

import (
	"context"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
)

func (r *Registry) GetIngresses(ctx context.Context) []*netv1.Ingress { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := r.harbor.Name

	u, err := url.Parse(r.harbor.Spec.PublicURL)
	if err != nil {
		panic(errors.Wrap(err, "invalid url"))
	}

	host := strings.SplitN(u.Host, ":", 1) // nolint:mnd

	var tls []netv1.IngressTLS
	if u.Scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: r.harbor.Spec.TLSSecretName,
			},
		}
	}

	return []*netv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      r.harbor.NormalizeComponentName(goharborv1alpha1.RegistryName),
				Namespace: r.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.RegistryName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: netv1.IngressSpec{
				TLS: tls,
				Rules: []netv1.IngressRule{
					{
						Host: host[0],
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{
										Path: "/v2",
										Backend: netv1.IngressBackend{
											ServiceName: r.harbor.NormalizeComponentName(goharborv1alpha1.RegistryName),
											ServicePort: intstr.FromInt(PublicPort),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
