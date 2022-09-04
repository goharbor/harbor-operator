package controller

import (
	"context"

	sgraph "github.com/goharbor/harbor-operator/pkg/controller/internal/graph"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Controller) Sweep(ctx context.Context, resource graph.Resource) error {
	obj, ok := resource.(client.Object)
	if !ok {
		return nil
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "sweep")
	defer span.Finish()

	namespace, name := obj.GetNamespace(), obj.GetName()
	gvk := c.AddGVKToSpan(ctx, span, obj)

	span.
		SetTag("resource.name", name).
		SetTag("resource.namespace", namespace)

	err := c.Delete(ctx, obj)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "sweep %s (%s/%s)", gvk, namespace, name)
	}

	return nil
}

func (c *Controller) Mark(ctx context.Context, owner client.Object) error { //nolint: funlen
	span, ctx := opentracing.StartSpanFromContext(ctx, "mark")
	defer span.Finish()

	g := sgraph.Get(ctx)
	if g == nil {
		return errors.Errorf("no graph in current context")
	}

	gvk := c.AddGVKToSpan(ctx, span, owner)
	reference := metav1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       owner.GetName(),
	}

	willDeletableResources := map[namespacedNameWithGVK]client.Object{}

	for deletableGVK := range c.deletableResources {
		unstructuredList := &unstructured.UnstructuredList{}
		unstructuredGVK := deletableGVK
		unstructuredList.SetGroupVersionKind(unstructuredGVK)

		if err := c.List(ctx, unstructuredList, &client.ListOptions{
			Namespace: owner.GetNamespace(),
		}); err != nil {
			return err
		}

		if err := unstructuredList.EachListItem(func(o runtime.Object) error {
			unstructured, ok := o.DeepCopyObject().(client.Object)
			if !ok {
				return nil
			}
			unstructured.GetObjectKind().SetGroupVersionKind(unstructuredGVK)
			if containsOwnerReferences(unstructured, reference) {
				willDeletableResources[namespacedNameWithGVK{
					NamespacedName: types.NamespacedName{
						Namespace: unstructured.GetNamespace(),
						Name:      unstructured.GetName(),
					},
					GroupVersionKind: unstructured.GetObjectKind().GroupVersionKind(),
				}] = unstructured
			}

			return nil
		}); err != nil {
			return err
		}
	}

	for _, res := range g.GetAllResources(ctx) {
		resource, ok := res.(*Resource)
		if !ok {
			continue
		}

		obj := resource.resource.DeepCopyObject().(client.Object)
		key := namespacedNameWithGVK{
			NamespacedName: types.NamespacedName{
				Namespace: obj.GetNamespace(),
				Name:      obj.GetName(),
			},
			GroupVersionKind: c.AddGVKToSpan(ctx, span, obj),
		}
		delete(willDeletableResources, key)
	}

	for _, o := range willDeletableResources {
		if err := g.AddResource(ctx, o.DeepCopyObject(), []graph.Resource{}, c.Sweep); err != nil {
			return err
		}
	}

	return nil
}

type namespacedNameWithGVK struct {
	types.NamespacedName
	schema.GroupVersionKind
}

func containsOwnerReferences(o client.Object, reference metav1.OwnerReference) bool {
	f := o.GetOwnerReferences()
	for _, e := range f {
		if e.APIVersion == reference.APIVersion &&
			e.Kind == reference.Kind &&
			e.Name == reference.Name {
			return true
		}
	}

	return false
}
