package graph_test

import (
	"context"
	"sync/atomic"

	. "github.com/goharbor/harbor-operator/pkg/graph"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// +kubebuilder:scaffold:imports
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Walk a dependency manager", func() {
	var rm Manager
	var ctx context.Context

	BeforeEach(func() {
		rm, ctx = setupTest(context.TODO())
	})

	Context("With no resource", func() {
		It("Should not call any function", func() {
			err := rm.Run(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("With a single resource", func() {
		var counter int32

		BeforeEach(func() {
			counter = 0

			add1 := func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				atomic.AddInt32(&counter, 1)

				return nil
			}

			err := rm.AddResource(ctx, &corev1.Secret{}, nil, add1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should call the function only once", func() {
			err := rm.Run(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(counter).To(BeEquivalentTo(1))
		})
	})

	Context("With 4 isolated resources", func() {
		var counter int32

		BeforeEach(func() {
			counter = 0

			add1 := func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				atomic.AddInt32(&counter, 1)

				return nil
			}

			err := rm.AddResource(ctx, &corev1.Namespace{}, nil, add1)
			Expect(err).ToNot(HaveOccurred())

			err = rm.AddResource(ctx, &corev1.Secret{}, nil, add1)
			Expect(err).ToNot(HaveOccurred())

			err = rm.AddResource(ctx, &corev1.ConfigMap{}, nil, add1)
			Expect(err).ToNot(HaveOccurred())

			err = rm.AddResource(ctx, &corev1.Node{}, nil, add1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should call the function exatly 4 times", func() {
			err := rm.Run(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(counter).To(BeEquivalentTo(4))
		})
	})

	Context("With a dependencies tree", func() {
		var expectations []types.GomegaMatcher

		var counter int32

		BeforeEach(func() {
			counter = 0

			countAndCheck := func(ctx context.Context, resource Resource) error {
				defer GinkgoRecover()

				Expect(resource).To(expectations[int(atomic.AddInt32(&counter, 1)-1)%len(expectations)])

				return nil
			}

			expectations = []types.GomegaMatcher{}

			ns := &corev1.Namespace{}
			err := rm.AddResource(ctx, ns, nil, countAndCheck)
			Expect(err).ToNot(HaveOccurred())

			expectations = append(expectations, Equal(ns))

			secret := &corev1.Secret{}
			err = rm.AddResource(ctx, secret, []Resource{ns}, countAndCheck)
			Expect(err).ToNot(HaveOccurred())

			cm := &corev1.ConfigMap{}
			err = rm.AddResource(ctx, cm, []Resource{ns}, countAndCheck)
			Expect(err).ToNot(HaveOccurred())

			expectations = append(expectations, BeElementOf(secret, cm), BeElementOf(secret, cm))

			no := &corev1.Node{}
			err = rm.AddResource(ctx, no, []Resource{secret, cm}, countAndCheck)
			Expect(err).ToNot(HaveOccurred())

			expectations = append(expectations, Equal(no))
		})

		It("Should call the function in the right order", func() {
			err := rm.Run(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(counter).To(BeEquivalentTo(len(expectations)))
		})

		It("Should accept multiple runs", func() {
			const runCount = 2

			for i := 0; i < runCount; i++ {
				err := rm.Run(ctx)
				Expect(err).ToNot(HaveOccurred())

				Expect(counter).To(BeEquivalentTo((i + 1) * len(expectations)))
			}
		})
	})
})
