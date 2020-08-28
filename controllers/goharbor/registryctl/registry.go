package registryctl

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:rbac:groups=goharbor.io,resources=registries,verbs=get;list;watch

func (r *Reconciler) GetRegistry(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*goharborv1alpha2.Registry, error) {
	var registry goharborv1alpha2.Registry

	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: registryCtl.GetNamespace(),
		Name:      registryCtl.Spec.RegistryRef,
	}, &registry)

	return &registry, err
}
