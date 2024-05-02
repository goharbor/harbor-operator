package project

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	goharborv1beta1 "github.com/plotly/harbor-operator/apis/goharbor.io/v1beta1"
	harborClient "github.com/plotly/harbor-operator/pkg/rest"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	ErrHarborCfgNotFound         = errors.New("harbor server configuration not found")
	ErrUnexpectedHarborCfgStatus = errors.New("status of Harbor server referred in configuration %s is unexpected")
)

// Reconcile does project reconcile.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) { //nolint:funlen
	log := r.Log.WithValues("resource", req.NamespacedName)
	log.Info("Start reconciling")

	// get HarborProject k8s resource from API
	hp := &goharborv1beta1.HarborProject{}
	if err = r.Client.Get(ctx, req.NamespacedName, hp); err != nil {
		if apierrors.IsNotFound(err) {
			// The resource may have be deleted after reconcile request coming in
			// Reconcile is done
			return ctrl.Result{}, nil
		}

		hp.Status.Reason = "HarborProjectError"

		return ctrl.Result{}, errors.Wrapf(err, "error get harbor project %v", req)
	}

	hp.Status.Status = goharborv1beta1.HarborProjectStatusUnknown

	defer func() {
		if err != nil {
			hp.Status.Status = goharborv1beta1.HarborProjectStatusFail
			hp.Status.Message = err.Error()
		} else {
			hp.Status.Status = goharborv1beta1.HarborProjectStatusReady
			hp.Status.Reason = ""
			hp.Status.Message = ""
			now := metav1.Now()
			hp.Status.LastApplyTime = &now
		}

		log.Info("Reconcile end", "result", res, "error", err, "updateStatusError", r.Client.Status().Update(ctx, hp))
	}()

	// set harbor client
	err = r.setHarborClient(ctx, hp.Spec.HarborServerConfig)
	if err != nil {
		err = errors.Wrapf(err, "error get harbor client")
		hp.Status.Reason = "HarborClientError"

		return
	}

	if !hp.ObjectMeta.DeletionTimestamp.IsZero() { //nolint:nestif
		// The object is being deleted
		if controllerutil.ContainsFinalizer(hp, finalizerID) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.Harbor.DeleteProject(hp.Spec.ProjectName); err != nil {
				hp.Status.Reason = "DeleteProjectError"
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(hp, finalizerID)

			if err := r.Update(ctx, hp); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(hp, finalizerID) {
		controllerutil.AddFinalizer(hp, finalizerID)

		if err := r.Update(ctx, hp); err != nil {
			return ctrl.Result{}, err
		}
	}

	projectExists, err := r.Harbor.ProjectExists(hp.Spec.ProjectName)
	if err != nil {
		err = errors.Wrapf(err, "error finding existing harbor project")
		hp.Status.Reason = "FindProjectError"

		return ctrl.Result{}, err
	}

	if projectExists {
		// update project
		if err = r.Harbor.UpdateProject(hp.Spec.ProjectName, hp); err != nil {
			err = errors.Wrapf(err, "error update harbor project")
			hp.Status.Reason = "UpdateProjectError"

			return ctrl.Result{}, err
		}
	} else {
		// create project
		id, err := r.Harbor.CreateProject(hp)
		if err != nil {
			err = errors.Wrapf(err, "error apply harbor project")
			hp.Status.Reason = "ApplyProjectError"

			return ctrl.Result{}, err
		}
		hp.Status.ProjectID = id
	}

	// reconcile project quota
	if err = r.reconcileQuota(hp, log); err != nil {
		err = errors.Wrapf(err, "error updating harbor project quota")
		hp.Status.Reason = "UpdateProjectQuotaError"

		return ctrl.Result{}, err
	}

	// reconcile project user/group memberships
	if err = r.reconcileMembership(hp, log); err != nil {
		err = errors.Wrapf(err, "error updating harbor project memberships")
		hp.Status.Reason = "UpdateProjectMembersError"

		return ctrl.Result{}, err
	}

	r.Log.Info("Reconcile is completed")

	return ctrl.Result{RequeueAfter: time.Minute * time.Duration(r.RequeueAfterMinutes)}, nil
}

// setHarborClient sets harbor client.
func (r *Reconciler) setHarborClient(ctx context.Context, harborServerConfigName string) error {
	harborCfg, err := r.getHarborServerConfig(ctx, harborServerConfigName)
	if err != nil {
		return fmt.Errorf("error finding harborCfg: %w", err)
	}

	if harborCfg == nil {
		// Not exist
		return fmt.Errorf("%w: %s", ErrHarborCfgNotFound, harborServerConfigName)
	}

	if harborCfg.Status.Status == goharborv1beta1.HarborServerConfigurationStatusUnknown || harborCfg.Status.Status == goharborv1beta1.HarborServerConfigurationStatusFail {
		return fmt.Errorf("%w harborCfg %s with %s", ErrUnexpectedHarborCfgStatus, harborCfg.Name, harborCfg.Status.Status)
	}

	// Create harbor client
	harborv2, err := harborClient.CreateHarborV2Client(ctx, r.Client, harborCfg)
	if err != nil {
		return err
	}

	r.Harbor = harborv2.WithContext(ctx)

	return nil
}

func (r *Reconciler) getHarborServerConfig(ctx context.Context, name string) (*goharborv1beta1.HarborServerConfiguration, error) {
	hsc := &goharborv1beta1.HarborServerConfiguration{}
	// HarborServerConfiguration is cluster scoped resource
	namespacedName := types.NamespacedName{
		Name: name,
	}
	if err := r.Client.Get(ctx, namespacedName, hsc); err != nil {
		// Explicitly check not found error
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return hsc, nil
}
