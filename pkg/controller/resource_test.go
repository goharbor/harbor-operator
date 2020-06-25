package controller_test

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"

	// +kubebuilder:scaffold:imports

	. "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = FContext("Adding", func() {
	var ctx context.Context
	var c *Controller

	resourceAdd := map[resources.Resource]func(*Controller, context.Context, resources.Resource, ...graph.Resource) (graph.Resource, error){
		&corev1.ConfigMap{}:   (*Controller).AddConfigMapToManage,
		&corev1.Secret{}:      (*Controller).AddSecretToManage,
		&netv1.Ingress{}:      (*Controller).AddIngressToManage,
		&certv1.Certificate{}: (*Controller).AddCertificateToManage,
		&appsv1.Deployment{}:  (*Controller).AddDeploymentToManage,
	}

	BeforeEach(func() {
		c, ctx = setupTest(context.TODO())

		ctx = c.PopulateContext(ctx, controllerruntime.Request{
			NamespacedName: types.NamespacedName{
				Name:      "resource-name",
				Namespace: "namespace",
			},
		})
	})

	for p, add := range resourceAdd {
		p, add := p, add

		Describe("A resource", func() {
			Context("Depending on", func() {
				var dependency graph.Resource

				Describe("An nil resource", func() {
					BeforeEach(func() {
						dependency = nil
					})

					It("Should work", func() {
						_, err := add(c, ctx, p, dependency)
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Describe("An unknown resource", func() {
					BeforeEach(func() {
						c := NewController(ctx, "test", nil, nil)
						d, err := c.AddConfigMapToManage(c.PopulateContext(context.TODO(), controllerruntime.Request{
							NamespacedName: types.NamespacedName{
								Name:      "resource-name",
								Namespace: "namespace",
							},
						}), &corev1.ConfigMap{})
						Expect(err).ToNot(HaveOccurred())

						dependency = d
					})

					It("Should return an error", func() {
						_, err := add(c, ctx, p, dependency)
						Expect(err).To(HaveOccurred())
					})
				})

				Describe("A configmap", func() {
					BeforeEach(func() {
						var err error
						dependency, err = c.AddConfigMapToManage(ctx, &corev1.ConfigMap{})
						Expect(err).ToNot(HaveOccurred())
					})

					It("Should work", func() {
						_, err := add(c, ctx, p, dependency)
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Describe("A certificate", func() {
					BeforeEach(func() {
						var err error
						dependency, err = c.AddCertificateToManage(ctx, &certv1.Certificate{})
						Expect(err).ToNot(HaveOccurred())
					})

					It("Should work", func() {
						_, err := add(c, ctx, p, dependency)
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Describe("A unstructured", func() {
					BeforeEach(func() {
						var err error
						dependency, err = c.AddUnsctructuredToManage(ctx, &unstructured.Unstructured{})
						Expect(err).ToNot(HaveOccurred())
					})

					It("Should work", func() {
						_, err := add(c, ctx, p, dependency)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

	}
})
