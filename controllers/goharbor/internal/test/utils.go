package test

import (
	"context"

	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func GetGVK(ctx context.Context, object runtime.Object) schema.GroupVersionKind {
	gvks, _, err := GetScheme(ctx).ObjectKinds(object)
	Expect(err).ToNot(HaveOccurred())

	return gvks[0]
}

func GetNamespacedName(object runtime.Object) types.NamespacedName {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	Expect(err).ToNot((HaveOccurred()))

	name, ok, err := unstructured.NestedString(data, "metadata", "name")
	Expect(err).ToNot(HaveOccurred())
	Expect(ok).To(BeTrue())

	namespace, _, err := unstructured.NestedString(data, "metadata", "namespace")
	Expect(err).ToNot(HaveOccurred())

	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}
