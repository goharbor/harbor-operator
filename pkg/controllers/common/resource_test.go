package common

import (
	"context"
	// +kubebuilder:scaffold:imports

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = FContext("Adding", func() {
	var c *Controller
	var ctx context.Context

	BeforeEach(func() {
		c, ctx = setupTest(context.TODO())
	})

	Describe("A pod", func() {
		var p *corev1.Pod

		BeforeEach(func() {
			p = &corev1.Pod{}
		})

		Context("Depending on", func() {
			var dependency graph.Resource

			Describe("An nil resource", func() {
				BeforeEach(func() {
					dependency = nil
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, p, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("An unknown resource", func() {
				BeforeEach(func() {
					dependency = &Resource{
						mutable:   c.GlobalMutateFn(ctx),
						checkable: statuscheck.BasicCheck,
						resource:  &corev1.Pod{},
					}
				})

				It("Should return an error", func() {
					_, err := c.AddBasicObjectToManage(ctx, p, dependency)
					Expect(err).To(HaveOccurred())
				})
			})

			Describe("A pod", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &corev1.Pod{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, p, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("A certificate", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &certv1.Certificate{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, p, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("A registry", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &goharborv1alpha2.Registry{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, p, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})

	Describe("A certificate", func() {
		var cert *certv1.Certificate

		BeforeEach(func() {
			cert = &certv1.Certificate{}
		})

		Context("Depending on", func() {
			var dependency graph.Resource

			Describe("An nil resource", func() {
				BeforeEach(func() {
					dependency = nil
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, cert, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("An unknown resource", func() {
				BeforeEach(func() {
					dependency = &Resource{
						mutable:   c.GlobalMutateFn(ctx),
						checkable: statuscheck.BasicCheck,
						resource:  &corev1.Pod{},
					}
				})

				It("Should return an error", func() {
					_, err := c.AddBasicObjectToManage(ctx, cert, dependency)
					Expect(err).To(HaveOccurred())
				})
			})

			Describe("A pod", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &corev1.Pod{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, cert, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("A certificate", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &certv1.Certificate{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, cert, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("A registry", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &goharborv1alpha2.Registry{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, cert, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})

	Describe("A registry", func() {
		var r *goharborv1alpha2.Registry

		BeforeEach(func() {
			r = &goharborv1alpha2.Registry{}
		})

		Context("Depending on", func() {
			var dependency graph.Resource

			Describe("An nil resource", func() {
				BeforeEach(func() {
					dependency = nil
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, r, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("An unknown resource", func() {
				BeforeEach(func() {
					dependency = &Resource{
						mutable:   c.GlobalMutateFn(ctx),
						checkable: statuscheck.BasicCheck,
						resource:  &corev1.Pod{},
					}
				})

				It("Should return an error", func() {
					_, err := c.AddBasicObjectToManage(ctx, r, dependency)
					Expect(err).To(HaveOccurred())
				})
			})

			Describe("A pod", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &corev1.Pod{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, r, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("A certificate", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &certv1.Certificate{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, r, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("A registry", func() {
				BeforeEach(func() {
					var err error
					dependency, err = c.AddBasicObjectToManage(ctx, &goharborv1alpha2.Registry{})
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should work", func() {
					_, err := c.AddBasicObjectToManage(ctx, r, dependency)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
