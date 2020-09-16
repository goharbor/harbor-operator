package chartmuseum

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/ingress"

	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
)

func (c *ChartMuseum) GetIngresses(ctx context.Context) []*netv1.Ingress { // nolint:funlen
	if c.harbor.Spec.Components.ChartMuseum == nil {
		// Not configured
		return make([]*netv1.Ingress, 0)
	}

	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	scheme, h, err := ingress.GetHostAndSchema(c.harbor.Spec.PublicURL)
	if err != nil {
		panic(err)
	}

	var tls []netv1.IngressTLS
	if scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: c.harbor.Spec.TLSSecretName,
			},
		}
	}

	return []*netv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(goharborv1alpha1.ChartMuseumName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.ChartMuseumName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: netv1.IngressSpec{
				TLS: tls,
				Rules: []netv1.IngressRule{
					{
						Host: h,
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{
										Path: "/chartrepo",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(goharborv1alpha1.CoreName),
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