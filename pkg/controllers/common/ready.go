package common

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serrors "github.com/goharbor/harbor-operator/pkg/controllers/common/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (c *Controller) EnsureReady(ctx context.Context, node graph.Resource) error {
	res, ok := node.(*Resource)
	if !ok {
		return errors.Errorf("unsupported resource type %+v", node)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "applyResource", opentracing.Tags{
		"Resource.Kind": res.resource.GetObjectKind().GroupVersionKind().GroupKind(),
	})
	defer span.Finish()

	objectKey, err := client.ObjectKeyFromObject(res.resource)
	if err != nil {
		return errors.Wrap(err, "cannot get object key")
	}

	span.
		SetTag("Resource.Name", objectKey.Name).
		SetTag("Resource.Namespace", objectKey.Namespace)

	result := res.resource.DeepCopyObject()

	err = c.Client.Get(ctx, objectKey, result)
	if err != nil {
		return errors.Wrap(err, "cannot get resource")
	}

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(result)
	if err != nil {
		return errors.Wrap(err, "cannot convert to unstructured")
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)

	ok, err = res.checkable(ctx, u)
	if err != nil {
		return errors.Wrap(err, "cannot get resource")
	}

	if !ok {
		return serrors.RetryLaterError{
			Cause: errors.Errorf("resource %+v not ready", res.resource),
		}
	}

	return nil
}
