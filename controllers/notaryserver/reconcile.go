package notaryserver

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

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=notaryservers,verbs=get;list;watch
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=notaryservers/status,verbs=get;update;patch

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()
	application.SetName(&ctx, r.GetName())
	application.SetVersion(&ctx, r.GetVersion())

	span, ctx := opentracing.StartSpanFromContext(ctx, "reconcile", opentracing.Tags{
		"NotaryServer.Namespace": req.Namespace,
		"NotaryServer.Name":      req.Name,
	})
	defer span.Finish()

	span.LogFields(
		log.String("NotaryServer.Namespace", req.Namespace),
		log.String("NotaryServer.Name", req.Name),
	)

	reqLogger := r.Log.WithValues("Request", req.NamespacedName, "NotaryServer.Namespace", req.Namespace, "NotaryServer.Name", req.Name)

	logger.Set(&ctx, reqLogger)

	// Fetch the NotaryServer instance
	notary := &goharborv1alpha2.NotaryServer{}

	ok, err := r.Controller.GetAndFilter(ctx, req.NamespacedName, notary)
	if err != nil {
		// Error reading the object
		return reconcile.Result{}, err
	}
	if !ok {
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		reqLogger.Info("NotaryServer does not exists")
		return reconcile.Result{}, nil
	}

	err = r.AddResources(ctx, notary)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "cannot add resources")
	}

	return r.Controller.Reconcile(ctx, notary)
}
