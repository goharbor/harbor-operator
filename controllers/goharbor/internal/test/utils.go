package test

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func AddVersionAnnotations(annotations map[string]string) map[string]string {
	return version.SetVersion(annotations, GetVersion())
}

func GetVersion() string {
	return "2.6.0"
}

func GetGVK(ctx context.Context, object runtime.Object) schema.GroupVersionKind {
	gvks, _, err := GetScheme(ctx).ObjectKinds(object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	return gvks[0]
}

func GetNamespacedName(object runtime.Object) types.NamespacedName {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	name, ok, err := unstructured.NestedString(data, "metadata", "name")
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(ok).To(gomega.BeTrue())

	namespace, _, err := unstructured.NestedString(data, "metadata", "namespace")
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

var SuccessOrExists = gomega.SatisfyAny(
	gomega.Succeed(),
	gomega.WithTransform(apierrors.IsAlreadyExists, gomega.BeTrue()),
)
