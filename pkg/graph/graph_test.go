package graph_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	// +kubebuilder:scaffold:imports

	. "github.com/goharbor/harbor-operator/pkg/graph"
)

var _ = Describe("With a dependency manager", func() {
	var rm Manager
	var ctx context.Context

	BeforeEach(func() {
		rm, ctx = setupTest(context.TODO())
	})

	Describe("Add 2 times the same resources", func() {
		It("Should fail", func() {
			secret := &corev1.Secret{}

			err := rm.AddResource(ctx, secret, nil)
			Expect(err).ToNot(HaveOccurred())

			err = rm.AddResource(ctx, secret, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Add a resource with an unknown dependency", func() {
		It("Should fail", func() {
			cm := &corev1.ConfigMap{}
			secret := &corev1.Secret{}

			err := rm.AddResource(ctx, secret, []Resource{cm})
			Expect(err).To(HaveOccurred())
		})
	})
})
