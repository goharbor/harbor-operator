package harborcluster

import (
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/gos"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	defaultWaitCycle = 10 * time.Second
)

// Reconcile logic of the HarborCluster
func (r *Reconciler) Reconcile(req ctrl.Request) (res ctrl.Result, err error) {
	ctx := context.TODO()
	log := r.Log.WithValues("resource", req.NamespacedName)

	// Get the harborcluster first
	harborcluster := &v1alpha2.HarborCluster{}
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
		log.Info("harbor cluster is being deleted", "name", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	// For tracking status
	st := NewStatus(harborcluster).
		WithContext(ctx).
		WithClient(r.Client).
		WithLog(log)

	defer func() {
		// Execute the status update operation
		if er := st.Update(); er != nil {
			sec, wait := apierrors.SuggestsClientDelay(err)
			if wait {
				res.RequeueAfter = time.Duration(sec) * time.Second
				r.Log.Info("suggest client delay", "seconds", sec)
			}

			er = fmt.Errorf("defer: update status error: %s", er)
			if err != nil {
				err = fmt.Errorf("%s, upstreaming error: %w", er.Error(), err)
			} else {
				err = er
			}
		}
	}()

	// Deploy or check dependent services concurrently and fail earlier
	g, gtx := gos.NewGroup(ctx)
	g.Go(func() error {
		cacheStatus, err := r.CacheCtrl.Apply(gtx, harborcluster)
		if cacheStatus != nil {
			st.UpdateCache(cacheStatus.Condition)
		}

		if err == nil {
			r.HarborCtrl.WithDependency(v1alpha2.ComponentCache, cacheStatus)
		}

		return err
	})

	g.Go(func() error {
		dbStatus, err := r.DatabaseCtrl.Apply(gtx, harborcluster)
		if dbStatus != nil {
			st.UpdateDatabase(dbStatus.Condition)
		}

		if err == nil {
			r.HarborCtrl.WithDependency(v1alpha2.ComponentDatabase, dbStatus)
		}

		return err
	})

	g.Go(func() error {
		storageStatus, err := r.StorageCtrl.Apply(gtx, harborcluster)
		if storageStatus != nil {
			st.UpdateStorage(storageStatus.Condition)
		}

		if err == nil {
			r.HarborCtrl.WithDependency(v1alpha2.ComponentStorage, storageStatus)
		}

		return err
	})

	if err := g.Wait(); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile dependent services error: %w", err)
	}

	if !st.DependsReady() {
		r.Log.Info("not all the dependent services are ready")
		return ctrl.Result{
			RequeueAfter: defaultWaitCycle,
		}, nil
	}

	// Create Harbor instance now
	harborStatus, err := r.HarborCtrl.Apply(ctx, harborcluster)
	if harborStatus != nil {
		st.UpdateHarbor(harborStatus.Condition)
	}
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile harbor service error: %w", err)
	}

	// Reconcile done
	r.Log.Info("reconcile is completed")
	return ctrl.Result{}, nil
}
