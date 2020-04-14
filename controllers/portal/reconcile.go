package portal

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=portals,verbs=get;list;watch
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=portals/status,verbs=get;update;patch

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()

	span, ctx := opentracing.StartSpanFromContext(ctx, "reconcile", opentracing.Tags{
		"Portal.Namespace": req.Namespace,
		"Portal.Name":      req.Name,
	})
	defer span.Finish()

	span.LogFields(
		log.String("Portal.Namespace", req.Namespace),
		log.String("Portal.Name", req.Name),
	)

	logger.Set(&ctx, r.Log)
	ctx = r.PopulateContext(ctx, req)
	l := logger.Get(ctx)

	// Fetch the Portal instance
	portal := &goharborv1alpha2.Portal{}

	ok, err := r.Controller.GetAndFilter(ctx, req.NamespacedName, portal)
	if err != nil {
		// Error reading the object
		return reconcile.Result{}, err
	}

	if !ok {
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		l.Info("Portal does not exists")
		return reconcile.Result{}, nil
	}

	err = r.AddResources(ctx, portal)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "cannot add resources")
	}

	return r.Controller.Reconcile(ctx, portal)
}
