package controller_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/pkg/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Context("Adding a resource", func() {
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
	}

	BeforeEach(func() {
		setupCtx := logger.Context(ctrlzap.New(ctrlzap.UseDevMode(true)))

		application.SetName(&setupCtx, "test-app")
		application.SetVersion(&setupCtx, "test")

		c = NewController(setupCtx, "test", nil, nil)

		s, err := scheme.New(setupCtx)
		Expect(err).ToNot(HaveOccurred())
		Expect(s).ToNot(BeNil())

		c.Scheme = s

		ctx = c.NewContext(controllerruntime.Request{
			NamespacedName: types.NamespacedName{
				Name:      "resource-name",
				Namespace: "namespace",
			},
		})

		logger.Set(&ctx, logger.Get(setupCtx))

		application.SetName(&ctx, "test-app")
		application.SetVersion(&ctx, "test")
	})

	for r, f := range resourceAdd {
		r, f := r, f

		var resource resources.Resource
		var add func(*Controller, context.Context, resources.Resource, ...graph.Resource) (graph.Resource, error)
		var dependency graph.Resource

		BeforeEach(func() {
			resource, add = r.DeepCopyObject().(resources.Resource), f
		})

		var kind string

		func() {
			defer GinkgoRecover()

			s, err := scheme.New(context.TODO())
			Expect(err).ToNot(HaveOccurred())
			Expect(s).ToNot(BeNil())

			gvks, _, err := s.ObjectKinds(r)
			Expect(err).ToNot(HaveOccurred())

			kind = gvks[0].Kind
		}()

		Describe(kind, func() {
			Context("Depending on an nil resource", func() {
				BeforeEach(func() {
					dependency = nil
				})

				It("Should work", func() {
					_, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("Depending on an unknown resource", func() {
				BeforeEach(func() {
					c := NewController(ctx, "test", nil, nil)

					s, err := scheme.New(ctx)
					Expect(err).ToNot(HaveOccurred())
					Expect(s).ToNot(BeNil())

					c.Scheme = s

					d, err := c.AddConfigMapToManage(c.NewContext(controllerruntime.Request{
						NamespacedName: types.NamespacedName{
							Name:      "resource-name",
							Namespace: "namespace",
						},
					}), &corev1.ConfigMap{})
					Expect(err).ToNot(HaveOccurred())

					dependency = d
				})

				It("Should return an error", func() {
					_, err := add(c, ctx, resource, dependency)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("Depending on a configmap", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddConfigMapToManage(ctx, &corev1.ConfigMap{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("Depending on a certificate", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddCertificateToManage(ctx, &certv1.Certificate{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("Depending on a unstructured", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddUnsctructuredToManage(ctx, &unstructured.Unstructured{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	}
})
