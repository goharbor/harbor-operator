package test

import (
	"context"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kstatus/status"
)

func EnsureReady(ctx context.Context, object runtime.Object, timeouts ...interface{}) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	Expect(err).ToNot((HaveOccurred()))

	gvk := GetGVK(ctx, object)
	namespacedName := GetNamespacedName(object)
	k8sClient := GetClient(ctx)

	generation, ok, err := unstructured.NestedInt64(data, "metadata", "generation")
	Expect(err).ToNot(HaveOccurred())
	Expect(ok).To(BeTrue())

	Eventually(func() interface{} {
		u := &unstructured.Unstructured{}

		u.SetUnstructuredContent(data)
		u.SetGroupVersionKind(gvk)

		Expect(k8sClient.Get(ctx, namespacedName, u)).
			To(Succeed())

		r, ok, err := unstructured.NestedFieldNoCopy(u.UnstructuredContent(), "status")
		Expect(err).
			ToNot(HaveOccurred())
		if !ok {
			return map[string]interface{}{}
		}

		return r
	}, timeouts...).
		Should(MatchKeys(IgnoreExtras, Keys{
			"observedGeneration": BeEquivalentTo(generation),
			"conditions": ContainElements(MatchKeys(IgnoreExtras, Keys{
				"type":   BeEquivalentTo(status.ConditionInProgress),
				"status": BeEquivalentTo(corev1.ConditionFalse),
			}), MatchKeys(IgnoreExtras, Keys{
				"type":   BeEquivalentTo(status.ConditionFailed),
				"status": BeEquivalentTo(corev1.ConditionFalse),
			})),
			"operator": MatchKeys(IgnoreExtras, Keys{
				"controllerVersion": BeEquivalentTo(application.GetVersion(ctx)),
			}),
		}), "resource should be applied")

	Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(data, object)).
		To(Succeed())
}

func ScaleUp(ctx context.Context, object runtime.Object) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	Expect(err).ToNot((HaveOccurred()))

	gvk := GetGVK(ctx, object)
	k8sClient := GetClient(ctx)

	u := &unstructured.Unstructured{}

	u.SetUnstructuredContent(data)
	u.SetGroupVersionKind(gvk)

	replicas, ok, err := unstructured.NestedInt64(data, "spec", "replicas")
	Expect(err).ToNot(HaveOccurred())
	if !ok {
		replicas = 1
	}

	replicas++

	unstructured.SetNestedField(data, replicas, "spec", "replicas")

	Expect(k8sClient.Update(ctx, object)).
		To(Succeed())
}
