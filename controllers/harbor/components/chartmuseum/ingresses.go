package chartmuseum

import (
	"context"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/goharbor/harbor-core-operator/api/v1alpha1"
	"github.com/goharbor/harbor-core-operator/pkg/factories/application"
)

func (c *ChartMuseum) GetIngresses(ctx context.Context) []*netv1.Ingress { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	u, err := url.Parse(c.harbor.Spec.PublicURL)
	if err != nil {
		panic(errors.Wrap(err, "invalid url"))
	}

	host := strings.SplitN(u.Host, ":", 1) // nolint:mnd

	var tls []netv1.IngressTLS
	if u.Scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: c.harbor.Spec.TLSSecretName,
			},
		}
	}

	return []*netv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.ChartMuseumName,
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
										Path: "/chartrepo",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
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
