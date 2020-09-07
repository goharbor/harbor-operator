package graph_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/pkg/graph"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("With a dependency manager", func() {
	var rm Manager
	var ctx context.Context

	BeforeEach(func() {
		rm, ctx = setupTest(context.TODO())
	})

	Describe("Add resource with nil function", func() {
		var resource Resource

		BeforeEach(func() {
			resource = &corev1.Secret{}
		})

		It("Should fail", func() {
			Expect(rm.AddResource(ctx, resource, nil, nil)).
				ToNot(Succeed())
		})
	})

	Describe("Add 2 times the same resources", func() {
		var resource Resource

		BeforeEach(func() {
			resource = &corev1.Secret{}
		})

		It("Should fail", func() {
			noOp := func(ctx context.Context, resource Resource) error { return nil }

			Expect(rm.AddResource(ctx, resource, nil, noOp)).
				To(Succeed())

			Expect(rm.AddResource(ctx, resource, nil, noOp)).
				ToNot(Succeed())
		})
	})

	Describe("Add a resource with an unknown dependency", func() {
		var resource, dependency Resource

		BeforeEach(func() {
			resource = &corev1.ConfigMap{}
			dependency = &corev1.Secret{}
		})

		It("Should fail", func() {
			noOp := func(ctx context.Context, resource Resource) error { return nil }

			Expect(rm.AddResource(ctx, resource, []Resource{dependency}, noOp)).
				ToNot(Succeed())
		})
	})
})
