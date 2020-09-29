package statuscheck

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
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

	conditions, found, err := unstructured.NestedSlice(uResource.UnstructuredContent(), "status", "conditions")
	if err != nil {
		return false, err
	}

	if !found || len(conditions) == 0 {
		return false, nil
	}

	ready := true
	errored := false
	inProgress := false
	ignoredConditions := []string{}

	for _, condition := range conditions {
		cond := condition.(map[string]interface{})

		switch cond["type"].(string) {
		case status.ConditionInProgress.String(), string(appsv1.DeploymentProgressing):
			inProgress = inProgress || cond["status"].(string) != string(corev1.ConditionFalse)
		case status.ConditionFailed.String(), string(appsv1.DeploymentReplicaFailure):
			errored = errored || cond["status"].(string) == string(corev1.ConditionTrue)
		case "Ready", string(appsv1.DeploymentAvailable):
			ready = ready && cond["status"].(string) == string(corev1.ConditionTrue)
		default:
			ignoredConditions = append(ignoredConditions, cond["type"].(string))
		}
	}

	if len(ignoredConditions) > 0 {
		logger.Get(ctx).V(1).
			Info("unexpected conditions", "conditions", ignoredConditions)
	}

	return ready && !inProgress && !errored, nil
}
