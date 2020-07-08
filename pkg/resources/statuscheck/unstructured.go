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

	err := status.Augment(uResource)
	if err != nil {
		return false, errors.Wrap(err, "cannot augment unstructured resource")
	}

	s, err := status.Compute(uResource)
	if err != nil {
		return false, errors.Wrap(err, "cannot compute status")
	}

	for _, cond := range s.Conditions {
		if cond.Status != corev1.ConditionTrue {
			continue
		}

		if cond.Type == status.ConditionInProgress || cond.Type == status.ConditionFailed {
			return false, nil
		}
	}

	return true, nil
}
