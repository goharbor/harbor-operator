package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler interface {
	reconcile.Reconciler

	NormalizeName(context.Context, string, ...string) string
	SetupWithManager(context.Context, manager.Manager) error
}
