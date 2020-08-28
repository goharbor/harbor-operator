package controller

import (
	"context"
	"fmt"
	"strings"

	cmutation "github.com/goharbor/harbor-operator/pkg/controller/mutation"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	WarningAnnotation = "goharbor.io/warning"
	WarningValueTmpl  = "⚠️ This Resource is managed by *%s* ⚠️"
)

const (
	OperatorNameLabel    = "goharbor.io/operator-controller"
	OperatorVersionLabel = "goharbor.io/operator-version"
)

func (c *Controller) GlobalMutateFn(ctx context.Context) resources.Mutable {
	var mutation resources.Mutable = cmutation.NoOp

	mutation.AppendMutation(cmutation.GetAnnotationsMutation(WarningAnnotation, fmt.Sprintf(WarningValueTmpl, c.GetName())))
	mutation.AppendMutation(cmutation.GetLabelsMutation(OperatorNameLabel, c.GetName(), OperatorVersionLabel, c.GetVersion()))

	return func(ctx context.Context, resource, result runtime.Object) controllerutil.MutateFn {
		// Get owner from this context, otherwise it is probably absent
		owner := owner.Get(ctx)
		if owner == nil {
			logger.Get(ctx).Info("Cannot add owner mutation: owner not found")
		} else {
			mutation.AppendMutation(cmutation.GetOwnerMutation(c.Scheme, owner))
		}

		mutation := mutation(ctx, resource, result)

		return func() (err error) {
			return mutation()
		}
	}
}

func (c *Controller) GetFQDN() string {
	return fmt.Sprintf("%s.goharbor.io", strings.ToLower(c.GetName()))
}

func (c *Controller) Label(suffix ...string) string {
	return c.LabelWithPrefix("", suffix...)
}

func (c *Controller) LabelWithPrefix(prefix string, suffix ...string) string {
	var suffixString string
	if len(suffix) > 0 {
		suffixString = "/" + strings.Join(suffix, "-")
	}

	if prefix != "" {
		prefix = "." + prefix
	}

	return fmt.Sprintf("%s%s%s", prefix, c.GetFQDN(), suffixString)
}

func (c *Controller) DeploymentMutateFn(ctx context.Context, dependencies ...graph.Resource) resources.Mutable {
	var mutation resources.Mutable = c.GlobalMutateFn(ctx)

	mutation.AppendMutation(cmutation.GetTemplateAnnotationsMutation(WarningAnnotation, fmt.Sprintf(WarningValueTmpl, c.GetName())))
	mutation.AppendMutation(cmutation.GetTemplateLabelsMutation(OperatorNameLabel, c.GetName(), OperatorVersionLabel, c.GetVersion()))

	fqdn := c.GetFQDN()

	mutation.AppendMutation(func(ctx context.Context, expected, result runtime.Object) controllerutil.MutateFn {
		var mutation resources.Mutable = c.GlobalMutateFn(ctx)

		for _, dep := range dependencies {
			res, ok := dep.(*Resource)
			if !ok {
				logger.Get(ctx).Info("Cannot add dependency checksum", "resource", dep)
				continue
			}

			depRemote, ok := res.resource.DeepCopyObject().(resources.Resource)
			if !ok {
				logger.Get(ctx).Info("Cannot add dependency checksum", "resource", dep)
				continue
			}

			name := res.resource.GetName()
			namespace := res.resource.GetNamespace()

			err := c.Client.Get(ctx, types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}, depRemote)
			if err != nil {
				return func() error { return errors.Wrap(err, "cannot get dependency") }
			}

			kind := strings.ToLower(depRemote.GetObjectKind().GroupVersionKind().Kind)

			mutation.AppendMutation(cmutation.GetTemplateAnnotationsMutation(
				fmt.Sprintf("%s.%s.%s.%s/uuid", name, namespace, kind, fqdn), string(depRemote.GetUID()),
				fmt.Sprintf("%s.%s.%s.%s/version", name, namespace, kind, fqdn), depRemote.GetResourceVersion(),
			))
		}

		return mutation(ctx, expected, result)
	})

	return mutation
}
