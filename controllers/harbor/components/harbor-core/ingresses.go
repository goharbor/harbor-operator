package core

import (
	"context"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/ingress"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (c *HarborCore) GetIngresses(ctx context.Context) []*netv1.Ingress { // nolint:funlen
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
				Hosts: []string{
					h,
				},
			},
		}
	}

	// Add annotations for cert-manager awareness
	annotations := ingress.GenerateIngressCertAnnotations(c.harbor.Spec)

	return []*netv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(goharborv1alpha1.CoreName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.CoreName,
					"harbor":   harborName,
					"operator": operatorName,
				},
				Annotations: annotations,
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
										Path: "/api",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(goharborv1alpha1.CoreName),
											ServicePort: intstr.FromInt(PublicPort),
										},
									}, {
										Path: "/c",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(goharborv1alpha1.CoreName),
											ServicePort: intstr.FromInt(PublicPort),
										},
									}, {
										Path: "/service",
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
