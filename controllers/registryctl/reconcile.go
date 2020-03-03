package registryctl

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

// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=registries,verbs=get;list;watch
// +kubebuilder:rbac:groups=containerregistry.ovhcloud.com,resources=registries/status,verbs=get;update;patch

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()
	application.SetName(&ctx, r.GetName())
	application.SetVersion(&ctx, r.GetVersion())

	span, ctx := opentracing.StartSpanFromContext(ctx, "reconcile", opentracing.Tags{
		"RegistryController.Namespace": req.Namespace,
		"RegistryController.Name":      req.Name,
	})
	defer span.Finish()

	span.LogFields(
		log.String("RegistryController.Namespace", req.Namespace),
		log.String("RegistryController.Name", req.Name),
	)

	reqLogger := r.Log.WithValues("Request", req.NamespacedName, "RegistryController.Namespace", req.Namespace, "RegistryController.Name", req.Name)

	logger.Set(&ctx, reqLogger)

	// Fetch the RegistryController instance
	registryctl := &goharborv1alpha2.RegistryController{}

	ok, err := r.Controller.GetAndFilter(ctx, req.NamespacedName, registryctl)
	if err != nil {
		// Error reading the object
		return reconcile.Result{}, err
	}
	if !ok {
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		reqLogger.Info("RegistryController does not exists")
		return reconcile.Result{}, nil
	}

	err = r.AddResources(ctx, registryctl)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "cannot add resources")
	}

	return r.Controller.Reconcile(ctx, registryctl)
}
