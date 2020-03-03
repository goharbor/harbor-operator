package core

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=cores,verbs=get;list;watch
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=cores/status,verbs=get;update;patch

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()
	application.SetName(&ctx, r.GetName())
	application.SetVersion(&ctx, r.GetVersion())

	span, ctx := opentracing.StartSpanFromContext(ctx, "reconcile", opentracing.Tags{
		"Core.Namespace": req.Namespace,
		"Core.Name":      req.Name,
	})
	defer span.Finish()

	span.LogFields(
		log.String("Core.Namespace", req.Namespace),
		log.String("Core.Name", req.Name),
	)

	reqLogger := r.Log.WithValues("Request", req.NamespacedName, "Core.Namespace", req.Namespace, "Core.Name", req.Name)

	logger.Set(&ctx, reqLogger)

	// Fetch the Core instance
	core := &goharborv1alpha2.Core{}

	ok, err := r.Controller.GetAndFilter(ctx, req.NamespacedName, core)
	if err != nil {
		// Error reading the object
		return reconcile.Result{}, err
	}
	if !ok {
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		reqLogger.Info("Core does not exists")
		return reconcile.Result{}, nil
	}

	err = r.AddResources(ctx, core)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "cannot add resources")
	}

	return r.Controller.Reconcile(ctx, core)
}
