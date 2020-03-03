package controllers

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Controller interface {
	reconcile.Reconciler

	SetupWithManager(ctrl.Manager) error
}
