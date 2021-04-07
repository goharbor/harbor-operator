package scheme_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/pkg/scheme"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Runtime scheme ", func() {
	var ctx context.Context
	var scheme *runtime.Scheme

	BeforeEach(func() {
		s, err := New(ctx)
		Expect(err).ToNot(HaveOccurred())

		scheme = s
	})

	DescribeTable("Must contain",
		func(gvk schema.GroupVersionKind) {
			Ω(scheme.AllKnownTypes()).Should(HaveKey(gvk))
		}, Entry("v1.cert-manager.io/Certificate", schema.GroupVersionKind{
			Group:   "cert-manager.io",
			Version: "v1",
			Kind:    "Certificate",
		}), Entry("v1.apps/Deployment", schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		}), Entry("v1beta1.extensions/Ingress", schema.GroupVersionKind{
			Group:   "extensions",
			Version: "v1beta1",
			Kind:    "Ingress",
		}), Entry("v1alpha3.goharbor.io/Harbor", schema.GroupVersionKind{
			Group:   "goharbor.io",
			Version: "v1alpha3",
			Kind:    "Exporter",
		}), Entry("v1alpha3.goharbor.io/Harbor", schema.GroupVersionKind{
			Group:   "goharbor.io",
			Version: "v1alpha3",
			Kind:    "Harbor",
		}),
	)
})
