package graph_test

import (
	. "github.com/goharbor/harbor-operator/pkg/graph"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// +kubebuilder:scaffold:imports
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("A resource type", func() {
	var resource Resource

	BeforeEach(func() {
		resource = &corev1.Secret{}
	})

	Describe("Compared", func() {
		Context("With the same resource", func() {
			var resource2 Resource

			BeforeEach(func() {
				resource2 = resource
			})

			It("Should be equal", func() {
				Expect(resource == resource2).To(BeTrue())
			})
		})

		Context("With another resource", func() {
			var resource2 Resource

			BeforeEach(func() {
				resource2 = &corev1.Secret{}
			})

			It("Should be equal", func() {
				Expect(resource == resource2).To(BeFalse())
			})
		})
	})
})
