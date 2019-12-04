package core

import (
	"context"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	extv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

func (c *HarborCore) GetIngresses(ctx context.Context) []*extv1.Ingress { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	u, err := url.Parse(c.harbor.Spec.PublicURL)
	if err != nil {
		panic(errors.Wrap(err, "invalid url"))
	}

	host := strings.SplitN(u.Host, ":", 1)

	var tls []extv1.IngressTLS
	if u.Scheme == "https" {
		tls = []extv1.IngressTLS{
			{
				SecretName: c.harbor.Spec.TLSSecretName,
			},
		}
	}

	return []*extv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.CoreName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: extv1.IngressSpec{
				TLS: tls,
				Backend: &extv1.IngressBackend{
					ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.PortalName),
					ServicePort: intstr.FromInt(80),
				},
				Rules: []extv1.IngressRule{
					{
						Host: host[0],
						IngressRuleValue: extv1.IngressRuleValue{
							HTTP: &extv1.HTTPIngressRuleValue{
								Paths: []extv1.HTTPIngressPath{
									{
										Path: "/api",
										Backend: extv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
											ServicePort: intstr.FromInt(80),
										},
									}, {
										Path: "/c",
										Backend: extv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
											ServicePort: intstr.FromInt(80),
										},
									}, {
										Path: "/chartrepo",
										Backend: extv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
											ServicePort: intstr.FromInt(80),
										},
									}, {
										Path: "/service",
										Backend: extv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
											ServicePort: intstr.FromInt(80),
										},
									}, {
										Path: "/service/notification",
										Backend: extv1.IngressBackend{
											ServiceName: "dev-null",
											ServicePort: intstr.FromInt(1),
										},
									}, {
										Path: "/v1",
										Backend: extv1.IngressBackend{
											ServiceName: "dev-null",
											ServicePort: intstr.FromInt(1),
										},
									}, {
										Path: "/v2",
										Backend: extv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.RegistryName),
											ServicePort: intstr.FromInt(80),
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
