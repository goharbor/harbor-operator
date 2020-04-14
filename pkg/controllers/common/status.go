package common

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

// UpdateStatus applies current in-memory statuses to the remote resource
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#status-subresource
func (c *Controller) UpdateStatus(ctx context.Context, result *ctrl.Result, object runtime.Object) error {
	err := c.Client.Status().Update(ctx, object)
	if err != nil {
		result.Requeue = true

		seconds, needWait := apierrors.SuggestsClientDelay(err)
		if needWait {
			result.RequeueAfter = time.Second * time.Duration(seconds)
		}

		if apierrors.IsConflict(err) {
			// the object has been modified; please apply your changes to the latest version and try again
			logger.Get(ctx).Error(err, "cannot update status field")
			return nil
		}

		return errors.Wrap(err, "cannot update status field")
	}

	return nil
}

func (c *Controller) ConditionToMap(ctx context.Context, condition goharborv1alpha2.Condition) map[string]interface{} {
	result := map[string]interface{}{}

	data, err := json.Marshal(condition)
	if err != nil {
		panic(errors.Wrap(err, "cannot convert to map: marshal failure"))
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		panic(errors.Wrap(err, "cannot convert to map: unmarshal failure"))
	}

	return result
}
