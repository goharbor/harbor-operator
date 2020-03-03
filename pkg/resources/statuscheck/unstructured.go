package statuscheck

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kstatus/status"
)

func UnstructuredCheck(ctx context.Context, object runtime.Object) (bool, error) {
	uResource := object.(*unstructured.Unstructured)

	s, err := status.Compute(uResource)
	if err != nil {
		return false, errors.Wrap(err, "cannot compute status")
	}

	for _, cond := range s.Conditions {
		if cond.Type == status.ConditionInProgress && cond.Status == corev1.ConditionTrue {
			return false, nil
		}
	}

	for _, cond := range s.Conditions {
		if cond.Type == status.ConditionFailed && cond.Status == corev1.ConditionTrue {
			return false, nil
		}
	}

	return true, nil
}
