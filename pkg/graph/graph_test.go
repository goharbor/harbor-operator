package graph_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/pkg/graph"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// +kubebuilder:scaffold:imports
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("With a dependency manager", func() {
	var rm Manager
	var ctx context.Context

	BeforeEach(func() {
		rm, ctx = setupTest(context.TODO())
	})

	Describe("Add resource with nil function", func() {
		It("Should fail", func() {
			secret := &corev1.Secret{}

			err := rm.AddResource(ctx, secret, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Add 2 times the same resources", func() {
		It("Should fail", func() {
			secret := &corev1.Secret{}

			err := rm.AddResource(ctx, secret, nil, func(ctx context.Context, resource Resource) error { return nil })
			Expect(err).ToNot(HaveOccurred())

			err = rm.AddResource(ctx, secret, nil, func(ctx context.Context, resource Resource) error { return nil })
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Add a resource with an unknown dependency", func() {
		It("Should fail", func() {
			cm := &corev1.ConfigMap{}
			secret := &corev1.Secret{}

			err := rm.AddResource(ctx, secret, []Resource{cm}, func(ctx context.Context, resource Resource) error { return nil })
			Expect(err).To(HaveOccurred())
		})
	})
})
