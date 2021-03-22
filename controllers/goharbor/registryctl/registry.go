package registryctl

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:rbac:groups=goharbor.io,resources=registries,verbs=get;list;watch

func (r *Reconciler) GetRegistry(ctx context.Context, registryCtl *goharborv1.RegistryController) (*goharborv1.Registry, error) {
	var registry goharborv1.Registry

	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: registryCtl.GetNamespace(),
		Name:      registryCtl.Spec.RegistryRef,
	}, &registry)

	return &registry, err
}
