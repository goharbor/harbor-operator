package harborcluster

import (
	"context"
	"fmt"
	"time"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/gos"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconcile logic of the HarborCluster.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) { //nolint:funlen
	ctx = r.PopulateContext(ctx, req)

	// Get the harborcluster first
	harborcluster := &goharborv1.HarborCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, harborcluster); err != nil {
		if apierrors.IsNotFound(err) {
			// The resource may have be deleted after reconcile request coming in
			// Reconcile is done
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get harbor cluster CR error: %w", err)
	}

	// Check if it is being deleted
	if !harborcluster.ObjectMeta.DeletionTimestamp.IsZero() {
		r.Log.Info("harbor cluster is being deleted", "name", req.NamespacedName)

		return ctrl.Result{}, nil
	}

	if err := r.PrepareStatus(ctx, harborcluster); err != nil {
		return r.HandleError(ctx, harborcluster, errors.Wrap(err, "cannot prepare owner status"))
	}

	// For tracking status
	st := newStatus(harborcluster).
		WithContext(ctx).
		WithClient(r.Client).
		WithLog(r.Log)

	defer func() {
		// Execute the status update operation
		if er := st.Update(); er != nil {
			sec, wait := apierrors.SuggestsClientDelay(err)
			if wait {
				res.RequeueAfter = time.Duration(sec) * time.Second
				r.Log.Info("suggest client delay", "seconds", sec)
			}

			er = fmt.Errorf("defer: update status error: %w", er)

			if err != nil {
				err = fmt.Errorf("%s, upstreaming error: %w", er.Error(), err)
			} else {
				err = er
			}
		}
	}()

	// Deploy or check dependent services concurrently and fail earlier.
	// Only need to do check if they're configured.
	g, gtx := gos.NewGroup(ctx)
	g.Go(func() error {
		mgr := NewServiceManager(goharborv1.ComponentCache)

		return mgr.WithContext(gtx).
			TrackedBy(st).
			From(harborcluster).
			Use(r.CacheCtrl).
			For(r.HarborCtrl).
			Apply()
	})

	g.Go(func() error {
		mgr := NewServiceManager(goharborv1.ComponentDatabase)

		return mgr.WithContext(gtx).
			TrackedBy(st).
			From(harborcluster).
			Use(r.DatabaseCtrl).
			For(r.HarborCtrl).
			Apply()
	})

	g.Go(func() error {
		mgr := NewServiceManager(goharborv1.ComponentStorage)

		return mgr.WithContext(gtx).
			TrackedBy(st).
			From(harborcluster).
			Use(r.StorageCtrl).
			For(r.HarborCtrl).
			Apply()
	})

	if err := g.Wait(); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile dependent services error: %w", err)
	}

	if !st.DependsReady() {
		r.Log.Info("not all the dependent services are ready")

		// The controller owns the dependent services so just return directly.
		return ctrl.Result{}, nil
	}

	// Create Harbor instance now
	harborStatus, err := r.HarborCtrl.Apply(ctx, harborcluster, lcm.WithDependencies(st.GetDependencies()))
	if harborStatus != nil {
		st.UpdateCondition(goharborv1.ServiceReady, harborStatus.Condition)
	}

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile harbor service error: %w", err)
	}

	// Reconcile done
	r.Log.Info("Reconcile is completed")

	if harborStatus.Condition.Status == corev1.ConditionTrue {
		return ctrl.Result{}, r.SetSuccessStatus(ctx, harborcluster)
	}

	return ctrl.Result{}, nil
}
