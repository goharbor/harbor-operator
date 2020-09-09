package registry

import (
	"context"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/ingress"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Registry) GetIngresses(ctx context.Context) []*netv1.Ingress { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := r.harbor.Name

	scheme, h, err := ingress.GetHostAndSchema(r.harbor.Spec.PublicURL)
	if err != nil {
		panic(err)
	}

	var tls []netv1.IngressTLS
	if scheme == "https" {
		tls = []netv1.IngressTLS{
			{
				SecretName: r.harbor.Spec.TLSSecretName,
				Hosts: []string{
					h,
				},
			},
		}
	}

	annotations := make(map[string]string)
	// resolve 413(Too Large Entity) error when push large image. It only works for NGINX ingress.
	annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = "0"

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
