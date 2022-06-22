package registry

import (
	"context"
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=goharbor.io,resources=registrycontrollers,verbs=get;list;watch

func (r *Reconciler) GetRegistryCtl(ctx context.Context, registry *goharborv1.Registry) (*goharborv1.RegistryController, error) {
	var registryCtl goharborv1.RegistryController

	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: registry.GetNamespace(),
		Name:      registry.Name,
	}, &registryCtl)

	return &registryCtl, err
}

// CleanUpRegistryCtlResources cleanup registryctl related resources.
func (r *Reconciler) CleanUpRegistryCtlResources(ctx context.Context, registryCtl *goharborv1.RegistryController) (err error) {
	namespace := registryCtl.GetNamespace()
	name := strings.NormalizeName(registryCtl.GetName(), RegistryCtlName)
	key := client.ObjectKey{Namespace: namespace, Name: name}

	// clean registrycontroller deployment
	if err = r.DeleteResourceIfExist(ctx, key, &appsv1.Deployment{}); err != nil {
		return fmt.Errorf("clean registryctl deployment error: %w", err)
	}

	// clean registrycontroller networkpolicy
	if err = r.DeleteResourceIfExist(ctx, client.ObjectKey{Namespace: namespace, Name: name + "-ingress"}, &netv1.NetworkPolicy{}); err != nil {
		return fmt.Errorf("clean registryctl networkpolicy error: %w", err)
	}

	svc := &corev1.Service{}
	if err = r.Client.Get(ctx, key, svc); err == nil && owneredByRegistryCtl(svc, registryCtl) {
		err = r.DeleteResourceIfExist(ctx, key, svc)
		if err != nil {
			return fmt.Errorf("clean registryctl service error: %w", err)
		}
	}

	cm := &corev1.ConfigMap{}
	if err = r.Client.Get(ctx, key, cm); err == nil && owneredByRegistryCtl(svc, registryCtl) {
		err = r.DeleteResourceIfExist(ctx, key, cm)
		if err != nil {
			return fmt.Errorf("clean registryctl configmap error: %w", err)
		}
	}

	return nil
}

// DeleteResourceIfExist deletes existed resources.
func (r *Reconciler) DeleteResourceIfExist(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	err := r.Client.Get(ctx, key, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	return r.Client.Delete(ctx, obj)
}

// owneredByRegistryCtl checks the resource whether ownered by
// registrycontroller.
func owneredByRegistryCtl(obj client.Object, registryCtl *goharborv1.RegistryController) bool {
	owners := obj.GetOwnerReferences()
	for _, o := range owners {
		if o.Kind == registryCtl.Kind && o.Name == registryCtl.Name && o.UID == registryCtl.UID {
			return true
		}
	}

	return false
}
