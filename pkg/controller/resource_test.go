package controller_test

import (
	"context"

	"github.com/goharbor/harbor-operator/controllers"
	. "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Context("Adding", func() {
	var ctx context.Context
	var c *Controller

	resourceAdd := map[resources.Resource]func(*Controller, context.Context, resources.Resource, ...graph.Resource) (graph.Resource, error){
		&corev1.ConfigMap{}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddConfigMapToManage(ctx, res.(*corev1.ConfigMap), dep...)
		},
		&corev1.Secret{}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddSecretToManage(ctx, res.(*corev1.Secret), dep...)
		},
		&netv1.Ingress{}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddIngressToManage(ctx, res.(*netv1.Ingress), dep...)
		},
		&certv1.Certificate{}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddCertificateToManage(ctx, res.(*certv1.Certificate), dep...)
		},
		&appsv1.Deployment{}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddDeploymentToManage(ctx, res.(*appsv1.Deployment), dep...)
		},
		&unstructured.Unstructured{}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddUnsctructuredToManage(ctx, res.(*unstructured.Unstructured), dep...)
		},
	}

	BeforeEach(func() {
		setupCtx := context.TODO()

		application.SetName(&setupCtx, "test-app")
		application.SetVersion(&setupCtx, "test")
		application.SetGitCommit(&setupCtx, "test")

		c = NewController(setupCtx, controllers.Controller(0), nil, nil)
		scheme, err := scheme.New(setupCtx)
		Expect(err).ToNot(HaveOccurred())

		c.Scheme = scheme

		ctx = c.PopulateContext(context.TODO(), controllerruntime.Request{
			NamespacedName: types.NamespacedName{
				Name:      "resource-name",
				Namespace: "namespace",
			},
		})

		owner.Set(&ctx, &appsv1.Deployment{})
		application.SetName(&ctx, "test-app")
		application.SetVersion(&ctx, "test")
		application.SetGitCommit(&ctx, "test")
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
						c := NewController(ctx, controllers.Controller(0), nil, nil)
						scheme, err := scheme.New(ctx)
						Expect(err).ToNot(HaveOccurred())

						c.Scheme = scheme

						context := c.PopulateContext(context.TODO(), controllerruntime.Request{
							NamespacedName: types.NamespacedName{
								Name:      "resource-name",
								Namespace: "namespace",
							},
						})

						owner.Set(&context, &appsv1.Deployment{})

						d, err := c.AddConfigMapToManage(context, &corev1.ConfigMap{})
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
