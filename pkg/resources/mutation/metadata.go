package mutation

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func MetadataMutateFn(ctx context.Context, configResource, configResult runtime.Object) controllerutil.MutateFn {
	result := configResult.(metav1.Object)
	desired := configResource.(metav1.Object)

	return func() error {
		result.SetAnnotations(desired.GetAnnotations())
		result.SetLabels(desired.GetLabels())

		return nil
	}
}
