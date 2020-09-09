package notary

import (
	"context"
	"fmt"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/ingress"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (n *Notary) GetIngresses(ctx context.Context) []*netv1.Ingress { // nolint:funlen
	if n.harbor.Spec.Components.Notary == nil {
		// Not configured
		return make([]*netv1.Ingress, 0)
	}

	operatorName := application.GetName(ctx)
	harborName := n.harbor.Name

	scheme, h, err := ingress.GetHostAndSchema(n.harbor.Spec.Components.Notary.PublicURL)
	if err != nil {
		panic(err)
	}

	// Add annotations for cert-manager awareness
	annotations := ingress.GenerateIngressCertAnnotations(n.harbor.Spec)

	var tls []netv1.IngressTLS
	if scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: fmt.Sprintf("%s-%s", n.harbor.Spec.TLSSecretName, goharborv1alpha1.NotaryName),
				Hosts: []string{
					h,
				},
			},
		}
	}

	return []*netv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      n.harbor.NormalizeComponentName(goharborv1alpha1.NotaryName),
				Namespace: n.harbor.Namespace,
				Labels: map[string]string{
					"app":                         goharborv1alpha1.NotaryName,
					"harbor":                      harborName,
					"operator":                    operatorName,
					"kubernetes.io/ingress.class": goharborv1alpha1.NotaryName,
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
										Path: "/",
										Backend: netv1.IngressBackend{
											ServiceName: n.harbor.NormalizeComponentName(NotaryServerName),
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
