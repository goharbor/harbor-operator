package test

import (
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureReady(ctx context.Context, object client.Object, timeouts ...interface{}) {
	matchReadyStatus, f := getStatusCheckFunc(ctx, object)

	gomega.Eventually(f, timeouts...).
		Should(matchReadyStatus, "resource should be applied")

	gomega.Expect(GetClient(ctx).Get(ctx, GetNamespacedName(object), object)).
		ToNot(gomega.HaveOccurred())

	gomega.Consistently(f, 2*time.Second, 500*time.Millisecond).
		Should(matchReadyStatus, "once ready, status should be constant")
}

func getStatusCheckFunc(ctx context.Context, object client.Object) (gomega.OmegaMatcher, func() (interface{}, error)) {
	k8sClient := GetClient(ctx)

	return gomega.BeTrue(), func() (interface{}, error) {
		err := k8sClient.Get(ctx, GetNamespacedName(object), object)
		if err != nil {
			return false, err
		}

		return statuscheck.BasicCheck(ctx, object)
	}
}

func ScaleUp(ctx context.Context, object client.Object) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	replicas, ok, err := unstructured.NestedInt64(data, "spec", "replicas")
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	if !ok {
		replicas = 1
	}

	replicas++

	patch := fmt.Sprintf(`[{"op":"replace","path":"/spec/replicas","value":%d}]`, replicas)

	gomega.Expect(GetClient(ctx).Patch(ctx, object, client.RawPatch(types.JSONPatchType, []byte(patch)))).
		To(gomega.Succeed())
}
