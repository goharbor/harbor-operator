package harborcluster

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/lcm"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func (r *Reconciler) NewEmpty(_ context.Context) resources.Resource {
	return &v1alpha2.HarborCluster{}
}

func (r *Reconciler) AddResources(ctx context.Context, resource resources.Resource) error {
	harborcluster, ok := resource.(*v1alpha2.HarborCluster)
	if !ok {
		return serrors.UnrecoverrableError(errors.Errorf("%+v", resource), serrors.OperatorReason, "unable to add resource")
	}

	componentToStatus := r.DefaultComponentStatus()
	cacheStatus, err := r.CacheCtrl.Apply(harborcluster)
	componentToStatus[v1alpha2.ComponentCache] = cacheStatus
	if err != nil {
		r.Log.Error(err, "error when reconcile cache component.")
		updateErr := r.UpdateHarborClusterStatus(ctx, harborcluster, componentToStatus)
		if updateErr != nil {
			r.Log.Error(updateErr, "update harbor cluster status")
		}
		return err
	}

	dbStatus, err := r.DatabaseCtrl.Apply(harborcluster)
	componentToStatus[v1alpha2.ComponentDatabase] = dbStatus
	if err != nil {
		r.Log.Error(err, "error when reconcile database component.")
		updateErr := r.UpdateHarborClusterStatus(ctx, harborcluster, componentToStatus)
		if updateErr != nil {
			r.Log.Error(updateErr, "update harbor cluster status")
		}
		return err
	}

	storageStatus, err := r.StorageCtrl.Apply(harborcluster)
	componentToStatus[v1alpha2.ComponentStorage] = storageStatus
	if err != nil {
		r.Log.Error(err, "error when reconcile storage component.")
		updateErr := r.UpdateHarborClusterStatus(ctx, harborcluster, componentToStatus)
		if updateErr != nil {
			r.Log.Error(updateErr, "update harbor cluster status")
		}
		return err
	}

	// if components is not all ready, requeue the HarborCluster
	if !r.ComponentsAreAllReady(componentToStatus) {
		r.Log.Info("components not all ready.",
			string(v1alpha2.ComponentCache), cacheStatus,
			string(v1alpha2.ComponentDatabase), dbStatus,
			string(v1alpha2.ComponentStorage), storageStatus)
		err = r.UpdateHarborClusterStatus(ctx, harborcluster, componentToStatus)
		return err
	}

	//getRegistry := func() *string {
	//	if harborcluster.Spec.ImageSource != nil && harborcluster.Spec.ImageSource.Registry != "" {
	//		return &harborCluster.Spec.ImageSource.Registry
	//	}
	//	return nil
	//}
	//var imageGetter image.Getter
	//if imageGetter, err = image.NewImageGetter(getRegistry(), harborCluster.Spec.Version); err != nil {
	//	log.Error(err, "error when create Getter.")
	//	return ReconcileWaitResult, err
	//}
	//r.option.ImageGetter = imageGetter

	harborStatus, err := r.HarborCtrl.Apply(harborcluster)
	if err != nil {
		r.Log.Error(err, "error when reconcile harbor service.")
		return err
	}
	componentToStatus[v1alpha2.ComponentHarbor] = harborStatus

	err = r.UpdateHarborClusterStatus(ctx, harborcluster, componentToStatus)
	if err != nil {
		r.Log.Error(err, "error when update harbor cluster status.")
		return err
	}
	// wait to resync to update status.
	return nil
}

func (r *Reconciler) DefaultComponentStatus() map[v1alpha2.Component]*lcm.CRStatus {
	return map[v1alpha2.Component]*lcm.CRStatus{
		v1alpha2.ComponentCache:    lcm.New(v1alpha2.CacheReady).WithStatus(corev1.ConditionUnknown),
		v1alpha2.ComponentDatabase: lcm.New(v1alpha2.DatabaseReady).WithStatus(corev1.ConditionUnknown),
		v1alpha2.ComponentStorage:  lcm.New(v1alpha2.CacheReady).WithStatus(corev1.ConditionUnknown),
		v1alpha2.ComponentHarbor:   lcm.New(v1alpha2.ServiceReady).WithStatus(corev1.ConditionUnknown),
	}
}

// ServicesAreAllReady check whether these components(includes cache, db, storage) are all ready.
func (r *Reconciler) ComponentsAreAllReady(serviceToMap map[v1alpha2.Component]*lcm.CRStatus) bool {
	for _, status := range serviceToMap {
		if status == nil {
			return false
		}

		if status.Condition.Type == v1alpha2.ServiceReady {
			continue
		}
		if status.Condition.Status != corev1.ConditionTrue {
			return false
		}
	}
	return true
}

// UpdateHarborClusterStatus will Update HarborCluster CR status, according the services reconcile result.
func (r *Reconciler) UpdateHarborClusterStatus(
	ctx context.Context,
	harborCluster *v1alpha2.HarborCluster,
	componentToCRStatus map[v1alpha2.Component]*lcm.CRStatus) error {
	for component, status := range componentToCRStatus {
		if status == nil {
			continue
		}
		var conditionType v1alpha2.HarborClusterConditionType
		var ok bool
		if conditionType, ok = ComponentToConditionType[component]; !ok {
			r.Log.Info(fmt.Sprintf("can not found the condition type for %s", component))
		}
		harborClusterCondition, defaulted := r.getHarborClusterCondition(harborCluster, conditionType)
		r.updateHarborClusterCondition(harborClusterCondition, status)
		if defaulted {
			harborCluster.Status.Conditions = append(harborCluster.Status.Conditions, *harborClusterCondition)
		}
	}
	r.Log.Info("update harbor cluster.", "harborcluster", harborCluster)
	return r.Update(ctx, harborCluster)
}

// updateHarborClusterCondition update condition according to status.
func (r *Reconciler) updateHarborClusterCondition(condition *v1alpha2.HarborClusterCondition, crStatus *lcm.CRStatus) {
	if condition.Type != crStatus.Condition.Type {
		return
	}

	if condition.Status != crStatus.Condition.Status ||
		condition.Message != crStatus.Condition.Message ||
		condition.Reason != crStatus.Condition.Reason {
		condition.Status = crStatus.Condition.Status
		condition.Message = crStatus.Condition.Message
		condition.Reason = crStatus.Condition.Reason
		condition.LastTransitionTime = metav1.Now()
	}
}

// getHarborClusterCondition will get HarborClusterCondition by conditionType
func (r *Reconciler) getHarborClusterCondition(
	harborCluster *v1alpha2.HarborCluster,
	conditionType v1alpha2.HarborClusterConditionType) (condition *v1alpha2.HarborClusterCondition, defaulted bool) {
	for i := range harborCluster.Status.Conditions {
		condition = &harborCluster.Status.Conditions[i]
		if condition.Type == conditionType {
			return condition, false
		}
	}
	return &v1alpha2.HarborClusterCondition{
		Type:               conditionType,
		LastTransitionTime: metav1.Now(),
		Status:             corev1.ConditionUnknown,
	}, true
}

var (
	ComponentToConditionType = map[v1alpha2.Component]v1alpha2.HarborClusterConditionType{
		v1alpha2.ComponentHarbor:   v1alpha2.ServiceReady,
		v1alpha2.ComponentCache:    v1alpha2.CacheReady,
		v1alpha2.ComponentStorage:  v1alpha2.StorageReady,
		v1alpha2.ComponentDatabase: v1alpha2.DatabaseReady,
	}
	ReconcileWaitResult = reconcile.Result{RequeueAfter: 30 * time.Second}
)
