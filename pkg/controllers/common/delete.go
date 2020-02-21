package common

import (
	"context"
	"fmt"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

var (
	gvkToDelete = []schema.GroupVersionKind{
		{
			Group:   corev1.SchemeGroupVersion.Group,
			Version: corev1.SchemeGroupVersion.Version,
			Kind:    "Service",
		}, {
			Group:   corev1.SchemeGroupVersion.Group,
			Version: corev1.SchemeGroupVersion.Version,
			Kind:    "ConfigMap",
		}, {
			Group:   netv1.SchemeGroupVersion.Group,
			Version: netv1.SchemeGroupVersion.Version,
			Kind:    "Ingress",
		}, {
			Group:   corev1.SchemeGroupVersion.Group,
			Version: corev1.SchemeGroupVersion.Version,
			Kind:    "Secret",
		}, {
			Group:   certv1.SchemeGroupVersion.Group,
			Version: certv1.SchemeGroupVersion.Version,
			Kind:    "Certificate",
		}, {
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    "Deployment",
		},
	}
)

// +kubebuilder:rbac:groups="",resources="configmaps",verbs=delete
// +kubebuilder:rbac:groups="",resources="secrets",verbs=delete
// +kubebuilder:rbac:groups="",resources="services",verbs=delete
// +kubebuilder:rbac:groups="apps",resources="deployments",verbs=delete
// +kubebuilder:rbac:groups="cert-manager.io",resources="certificates",verbs=delete
// +kubebuilder:rbac:groups="networking.k8s.io",resources="ingresses",verbs=delete

func (r *Controller) DeleteResourceCollection(ctx context.Context, harbor *goharborv1alpha1.Harbor, componentName string, gvk schema.GroupVersionKind) error {
	u := &unstructured.UnstructuredList{}
	u.SetGroupVersionKind(gvk)

	matchingLabel := client.MatchingLabels{
		goharborv1alpha1.ComponentNameLabel: componentName,
	}
	inNamespace := client.InNamespace(harbor.GetNamespace())

	// TODO Use r.Client.DeleteAllOf function

	limit := 5

	err := r.Client.List(ctx, u, inNamespace, matchingLabel, client.Limit(limit))

	if apierrors.IsNotFound(err) {
		logger.Get(ctx).Info("Cannot list resource to delete, endpoint not found", "GVK.Group", gvk.Group, "GVK.Version", gvk.Version, "GVK.Kind", gvk.Kind)
		return nil
	}

	countToDelete := len(u.Items)
	if countToDelete == 0 {
		return nil
	}

	count := 0
	err = u.EachListItem(func(object runtime.Object) error {
		err := r.Client.Delete(ctx, object)
		err = client.IgnoreNotFound(err)
		if err == nil {
			count++
		}
		return err
	})

	logger.Get(ctx).Info(fmt.Sprintf("%d/%d resources deleted", count, countToDelete), "GVK.Group", gvk.Group, "GVK.Version", gvk.Version, "GVK.Kind", gvk.Kind)

	if err != nil {
		return errors.Wrap(err, "cannot delete object")
	}

	if limit == countToDelete {
		return errors.New("some resource to delete may remain")
	}

	return nil
}

func (r *Controller) DeleteComponent(ctx context.Context, harbor *goharborv1alpha1.Harbor, componentName string) error {
	var g errgroup.Group

	l := logger.Get(ctx).WithValues("Component", componentName)
	logger.Set(&ctx, l)

	l.Info("Deleting component")

	for _, gvk := range gvkToDelete {
		gvk := gvk

		g.Go(func() error {
			err := r.DeleteResourceCollection(ctx, harbor, componentName, gvk)
			return errors.Wrapf(err, "deletecollection failed for %s", gvk.String())
		})
	}

	return g.Wait()
}
