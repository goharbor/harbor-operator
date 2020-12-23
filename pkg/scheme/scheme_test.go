package scheme_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/pkg/scheme"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Runtime scheme", func() {
	var ctx context.Context
	var scheme *runtime.Scheme

	BeforeEach(func() {
		s, err := New(ctx)
		Expect(err).ToNot(HaveOccurred())

		scheme = s
	})

	It("Must contains AdmissionReview from admission.k8s.io/v1", func() {
		Ω(scheme.AllKnownTypes()).Should(HaveKey(schema.GroupVersionKind{
			Group:   "admission.k8s.io",
			Version: "v1",
			Kind:    "AdmissionReview",
		}))
	})

	It("Must contains AdmissionReview from admission.k8s.io/v1beta1", func() {
		Ω(scheme.AllKnownTypes()).Should(HaveKey(schema.GroupVersionKind{
			Group:   "admission.k8s.io",
			Version: "v1beta1",
			Kind:    "AdmissionReview",
		}))
	})
})
