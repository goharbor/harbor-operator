package notary

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
	"github.com/goharbor/harbor-operator/pkg/ingress"
)

func (n *Notary) GetIngresses(ctx context.Context) []*netv1.Ingress {
	operatorName := application.GetName(ctx)
	harborName := n.harbor.Name

	u, err := url.Parse(n.harbor.Spec.Components.Notary.PublicURL)
	if err != nil {
		panic(errors.Wrap(err, "invalid url"))
	}

	host := strings.SplitN(u.Host, ":", 1)

	var tls []netv1.IngressTLS
	if u.Scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: n.harbor.Spec.TLSSecretName,
				Hosts: []string{
					host[0],
				},
			},
		}
	}

	// Add annotations for cert-manager awareness
	annotations := ingress.GenerateIngressCertAnnotations(n.harbor.Spec)

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
						Host: host[0],
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
