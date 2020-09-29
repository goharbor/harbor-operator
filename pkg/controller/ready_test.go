package controller_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/pkg/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/kustomize/kstatus/status"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Context("Checking a ready resource", func() {
	var ctx context.Context
	var c *Controller

	readyResourceAdd := map[resources.Resource]func(*Controller, context.Context, resources.Resource, ...graph.Resource) (graph.Resource, error){
		&corev1.ConfigMap{}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddConfigMapToManage(ctx, res.(*corev1.ConfigMap), dep...)
		},
		&corev1.Secret{}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddSecretToManage(ctx, res.(*corev1.Secret), dep...)
		},
		&netv1.Ingress{
			Status: netv1.IngressStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{
						Hostname: "the-host.name",
						IP:       "127.0.0.1",
					}},
				},
			},
		}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddIngressToManage(ctx, res.(*netv1.Ingress), dep...)
		},
		&certv1.Certificate{
			Status: certv1.CertificateStatus{
				Conditions: []certv1.CertificateCondition{{
					Type:   certv1.CertificateConditionReady,
					Status: cmmeta.ConditionTrue,
				}},
			},
		}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddCertificateToManage(ctx, res.(*certv1.Certificate), dep...)
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 2,
			},
			Status: appsv1.DeploymentStatus{
				ObservedGeneration:  2,
				AvailableReplicas:   1,
				UpdatedReplicas:     1,
				ReadyReplicas:       1,
				Replicas:            1,
				UnavailableReplicas: 0,
				Conditions: []appsv1.DeploymentCondition{{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				}, {
					Type:   appsv1.DeploymentProgressing,
					Status: corev1.ConditionFalse,
				}, {
					Type:   appsv1.DeploymentReplicaFailure,
					Status: corev1.ConditionFalse,
				}},
			},
		}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddDeploymentToManage(ctx, res.(*appsv1.Deployment), dep...)
		},
		&goharborv1alpha2.ChartMuseum{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
			},
			Status: harbormetav1.ComponentStatus{
				ObservedGeneration: 1,
				Conditions: []harbormetav1.Condition{{
					Type:   status.ConditionInProgress,
					Status: corev1.ConditionFalse,
				}, {
					Type:   status.ConditionFailed,
					Status: corev1.ConditionFalse,
				}},
			},
		}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddBasicResource(ctx, res.(*goharborv1alpha2.ChartMuseum), dep...)
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

	for r, f := range readyResourceAdd {
		r, f := r, f

		var resource resources.Resource
		var add func(*Controller, context.Context, resources.Resource, ...graph.Resource) (graph.Resource, error)
		var dependency graph.Resource

		BeforeEach(func() {
			resource, add = r.DeepCopyObject().(resources.Resource), f
		})

		s, err := scheme.New(context.TODO())
		Expect(err).ToNot(HaveOccurred())
		Expect(s).ToNot(BeNil())

		gvks, _, err := s.ObjectKinds(r)
		Expect(err).ToNot(HaveOccurred())

		Describe(gvks[0].Kind, func() {
			Context("Depending on an nil resource", func() {
				BeforeEach(func() {
					dependency = nil
				})

				It("Should be ready", func() {
					node, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())

					resource := node.(*Resource)

					Ω(resource.Checkable(ctx, resource.Resource)).
						Should(BeTrue())
				})
			})

			Context("Depending on a configmap", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddConfigMapToManage(ctx, &corev1.ConfigMap{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should be ready", func() {
					node, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())

					resource := node.(*Resource)

					Ω(resource.Checkable(ctx, resource.Resource)).
						Should(BeTrue())
				})
			})

			Context("Depending on a not ready certificate", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddCertificateToManage(ctx, &certv1.Certificate{
						Status: certv1.CertificateStatus{
							Conditions: []certv1.CertificateCondition{
								{
									Type:   certv1.CertificateConditionReady,
									Status: cmmeta.ConditionFalse,
								},
							},
						},
					})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should be ready", func() {
					node, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())

					resource := node.(*Resource)

					Ω(resource.Checkable(ctx, resource.Resource)).
						Should(BeTrue())
				})
			})
		})
	}
})

var _ = Context("Checking a not ready resource", func() {
	var ctx context.Context
	var c *Controller

	readyResourceAdd := map[resources.Resource]func(*Controller, context.Context, resources.Resource, ...graph.Resource) (graph.Resource, error){
		&certv1.Certificate{
			Status: certv1.CertificateStatus{
				Conditions: []certv1.CertificateCondition{{
					Type:   certv1.CertificateConditionReady,
					Status: cmmeta.ConditionFalse,
				}},
			},
		}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddCertificateToManage(ctx, res.(*certv1.Certificate), dep...)
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 2,
			},
			Status: appsv1.DeploymentStatus{
				ObservedGeneration:  2,
				AvailableReplicas:   0, // Not ready here here
				UpdatedReplicas:     1,
				ReadyReplicas:       1,
				Replicas:            1,
				UnavailableReplicas: 0,
				Conditions: []appsv1.DeploymentCondition{{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				}, {
					Type:   appsv1.DeploymentProgressing,
					Status: corev1.ConditionFalse,
				}, {
					Type:   appsv1.DeploymentReplicaFailure,
					Status: corev1.ConditionFalse,
				}},
			},
		}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddDeploymentToManage(ctx, res.(*appsv1.Deployment), dep...)
		},
		&goharborv1alpha2.ChartMuseum{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 2,
			},
			Status: harbormetav1.ComponentStatus{
				ObservedGeneration: 1, // Failure here
				Conditions: []harbormetav1.Condition{{
					Type:   status.ConditionInProgress,
					Status: corev1.ConditionFalse,
				}, {
					Type:   status.ConditionFailed,
					Status: corev1.ConditionFalse,
				}},
			},
		}: func(c *Controller, ctx context.Context, res resources.Resource, dep ...graph.Resource) (graph.Resource, error) {
			return c.AddBasicResource(ctx, res.(*goharborv1alpha2.ChartMuseum), dep...)
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

	for r, f := range readyResourceAdd {
		r, f := r, f

		var resource resources.Resource
		var add func(*Controller, context.Context, resources.Resource, ...graph.Resource) (graph.Resource, error)
		var dependency graph.Resource

		BeforeEach(func() {
			resource, add = r.DeepCopyObject().(resources.Resource), f
		})

		s, err := scheme.New(context.TODO())
		Expect(err).ToNot(HaveOccurred())
		Expect(s).ToNot(BeNil())

		gvks, _, err := s.ObjectKinds(r)
		Expect(err).ToNot(HaveOccurred())

		Describe(gvks[0].Kind, func() {
			Context("Depending on an nil resource", func() {
				BeforeEach(func() {
					dependency = nil
				})

				It("Should be not ready", func() {
					node, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())

					resource := node.(*Resource)

					Ω(resource.Checkable(ctx, resource.Resource)).
						Should(BeFalse())
				})
			})

			Context("Depending on a configmap", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddConfigMapToManage(ctx, &corev1.ConfigMap{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should be not ready", func() {
					node, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())

					resource := node.(*Resource)

					Ω(resource.Checkable(ctx, resource.Resource)).
						Should(BeFalse())
				})
			})

			Context("Depending on a not ready certificate", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddCertificateToManage(ctx, &certv1.Certificate{
						Status: certv1.CertificateStatus{
							Conditions: []certv1.CertificateCondition{
								{
									Type:   certv1.CertificateConditionReady,
									Status: cmmeta.ConditionFalse,
								},
							},
						},
					})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should be not ready", func() {
					node, err := add(c, ctx, resource, dependency)
					Expect(err).ToNot(HaveOccurred())

					resource := node.(*Resource)

					Ω(resource.Checkable(ctx, resource.Resource)).
						Should(BeFalse())
				})
			})
		})
	}
})
