package common

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/harbor/components"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

func (r *Controller) MutateAnnotations(ctx context.Context, resource metav1.Object) {
	annotations := resource.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	// Warning annotation
	annotations[goharborv1alpha1.WarningLabel] = fmt.Sprintf("⚠️ This Resource is managed by *%s* ⚠️", r.GetName())
	resource.SetAnnotations(annotations)
}

func (r *Controller) MutateLabels(ctx context.Context, resource metav1.Object) {
	labels := resource.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	labels[goharborv1alpha1.OperatorNameLabel] = r.GetName()
	labels[goharborv1alpha1.OperatorVersionLabel] = r.GetVersion()

	labels[goharborv1alpha1.ComponentNameLabel] = components.ComponentName(ctx)

	resource.SetLabels(labels)
}

func (r *Controller) CreateResource(ctx context.Context, harbor *goharborv1alpha1.Harbor, resource components.Resource) error {
	kind, version := resource.
		GetObjectKind().
		GroupVersionKind().
		ToAPIVersionAndKind()

	span, ctx := opentracing.StartSpanFromContext(ctx, "createResource", opentracing.Tags{
		"Resource.Kind":    kind,
		"Resource.Version": version,
	})
	defer span.Finish()

	r.MutateAnnotations(ctx, resource)
	r.MutateLabels(ctx, resource)

	// Set Harbor instance as the owner and controller of the resource
	err := controllerutil.SetControllerReference(harbor, resource, r.Scheme)
	if err != nil {
		return errors.Wrap(err, "cannot set controller reference")
	}

	err = r.Client.Create(ctx, resource)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			return nil
		}

		return errors.Wrapf(err, "cannot create/update %s/%s", resource.GroupVersionKind().GroupKind(), resource.GetName())
	}

	logger.Get(ctx).Info("resource created")

	return nil
}

func (r *Controller) CreateResources(ctx context.Context, harbor *goharborv1alpha1.Harbor, resources []components.Resource) error {
	var g errgroup.Group

	for _, resource := range resources {
		resource := resource

		g.Go(func() error {
			return r.CreateResource(ctx, harbor, resource)
		})
	}

	return g.Wait()
}

// +kubebuilder:rbac:groups="",resources="configmaps",verbs=create
// +kubebuilder:rbac:groups="",resources="secrets",verbs=create
// +kubebuilder:rbac:groups="",resources=services,verbs="create"
// +kubebuilder:rbac:groups="apps",resources="deployments",verbs=create
// +kubebuilder:rbac:groups="cert-manager.io",resources="certificates",verbs=create
// +kubebuilder:rbac:groups="networking.k8s.io",resources="ingresses",verbs=create

func (r *Controller) CreateComponent(ctx context.Context, harbor *goharborv1alpha1.Harbor, component *components.ComponentRunner) error {
	return component.ParallelRun(ctx, harbor, r.CreateResources, r.CreateResources, r.CreateResources, r.CreateResources, r.CreateResources, r.CreateResources, true)
}

func (r *Controller) Create(ctx context.Context, harbor *goharborv1alpha1.Harbor) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "apply")
	defer span.Finish()

	harborResource, err := components.GetComponents(ctx, harbor)
	if err != nil {
		return errors.Wrap(err, "cannot get resources to manage")
	}

	err = harborResource.ParallelRun(ctx, harbor, r.CreateComponent)

	return errors.Wrap(err, "cannot deploy component")
}
