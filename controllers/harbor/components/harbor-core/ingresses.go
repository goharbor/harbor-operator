package core

import (
	"context"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/controllers/harbor/components/chartmuseum"
	"github.com/ovh/harbor-operator/controllers/harbor/components/portal"
	"github.com/ovh/harbor-operator/controllers/harbor/components/registry"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	emptyPort = 1
)

func (c *HarborCore) GetIngresses(ctx context.Context) []*netv1.Ingress { // nolint:funlen
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
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.CoreName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: netv1.IngressSpec{
				TLS: tls,
				Backend: &netv1.IngressBackend{
					ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.PortalName),
					ServicePort: intstr.FromInt(portal.PublicPort),
				},
				Rules: []netv1.IngressRule{
					{
						Host: host[0],
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{
										Path: "/api",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
											ServicePort: intstr.FromInt(PublicPort),
										},
									}, {
										Path: "/c",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
											ServicePort: intstr.FromInt(PublicPort),
										},
									}, {
										Path: "/chartrepo",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
											ServicePort: intstr.FromInt(chartmuseum.PublicPort),
										},
									}, {
										Path: "/service",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.CoreName),
											ServicePort: intstr.FromInt(PublicPort),
										},
									}, {
										Path: "/service/notification",
										Backend: netv1.IngressBackend{
											ServiceName: "dev-null",
											ServicePort: intstr.FromInt(emptyPort),
										},
									}, {
										Path: "/v1",
										Backend: netv1.IngressBackend{
											ServiceName: "dev-null",
											ServicePort: intstr.FromInt(emptyPort),
										},
									}, {
										Path: "/v2",
										Backend: netv1.IngressBackend{
											ServiceName: c.harbor.NormalizeComponentName(containerregistryv1alpha1.RegistryName),
											ServicePort: intstr.FromInt(registry.PublicPort),
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
