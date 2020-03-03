package graph

import (
	"context"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("Walk a dependency manager", func() {
	var rm *resourceManager
	var ctx context.Context

	BeforeEach(func() {
		rm, ctx = setupTest(context.TODO())
	})

	Context("With no resource", func() {
		It("Should not call the function", func() {
			err := rm.Run(ctx, func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				Fail("callback called")

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("With a single resource", func() {
		BeforeEach(func() {
			err := rm.AddResource(&corev1.Secret{}, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should call the function only once", func() {
			counter := int32(0)

			err := rm.Run(ctx, func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				atomic.AddInt32(&counter, 1)

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(counter).To(BeEquivalentTo(1))
		})
	})

	Context("With 4 isolated resources", func() {
		BeforeEach(func() {
			err := rm.AddResource(&corev1.Namespace{}, nil)
			Expect(err).ToNot(HaveOccurred())

			err = rm.AddResource(&corev1.Secret{}, nil)
			Expect(err).ToNot(HaveOccurred())

			err = rm.AddResource(&corev1.ConfigMap{}, nil)
			Expect(err).ToNot(HaveOccurred())

			err = rm.AddResource(&corev1.Node{}, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should call the function exatly 4 times", func() {
			counter := int32(0)

			err := rm.Run(ctx, func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				atomic.AddInt32(&counter, 1)

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(counter).To(BeEquivalentTo(4))
		})
	})

	Context("With a dependencies tree", func() {
		var expectations []types.GomegaMatcher

		BeforeEach(func() {
			expectations = []types.GomegaMatcher{}

			ns := &corev1.Namespace{}
			err := rm.AddResource(ns, nil)
			Expect(err).ToNot(HaveOccurred())

			expectations = append(expectations, Equal(ns))

			secret := &corev1.Secret{}
			err = rm.AddResource(secret, []Resource{ns})
			Expect(err).ToNot(HaveOccurred())

			cm := &corev1.ConfigMap{}
			err = rm.AddResource(cm, []Resource{ns})
			Expect(err).ToNot(HaveOccurred())

			expectations = append(expectations, BeElementOf(secret, cm), BeElementOf(secret, cm))

			no := &corev1.Node{}
			err = rm.AddResource(no, []Resource{secret, cm})
			Expect(err).ToNot(HaveOccurred())

			expectations = append(expectations, Equal(no))
		})

		It("Should call the function in the right order", func() {
			counter := int32(0)

			err := rm.Run(ctx, func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				Expect(resource).To(expectations[atomic.AddInt32(&counter, 1)-1])
				time.Sleep(200 * time.Millisecond)

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(counter).To(BeEquivalentTo(len(expectations)))
		})

		It("Should accept multiple runs", func() {
			const runCount = 2
			globalCounter := int32(0)

			var g errgroup.Group

			for i := 0; i < runCount; i++ {
				counter := int32(0)

				g.Go(func() error {
					defer GinkgoRecover()

					return rm.Run(ctx, func(ctx context.Context, resource Resource) error {
						defer GinkgoRecover()

						atomic.AddInt32(&globalCounter, 1)

						Expect(resource).To(expectations[atomic.AddInt32(&counter, 1)-1])
						time.Sleep(200 * time.Millisecond)

						return nil
					})
				})
			}

			err := g.Wait()
			Expect(err).ToNot(HaveOccurred())

			Expect(globalCounter).To(BeEquivalentTo(runCount * len(expectations)))
		})
	})
})
