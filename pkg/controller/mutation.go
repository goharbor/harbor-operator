package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/goharbor/harbor-operator/pkg/controller/mutation"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var ErrorOwnerNotFound = errors.New("owner not found")

const (
	WarningAnnotation = "goharbor.io/warning"
	WarningValueTmpl  = "⚠️ This Resource is managed by *%s* ⚠️"
)

const (
	OperatorNameLabel    = "goharbor.io/operator-controller"
	OperatorVersionLabel = "goharbor.io/operator-version"
)

func (c *Controller) GlobalMutateFn(ctx context.Context) (resources.Mutable, error) {
	var mutate resources.Mutable = mutation.NoOp

	mutate.AppendMutation(mutation.GetAnnotationsMutation(WarningAnnotation, fmt.Sprintf(WarningValueTmpl, c.GetName())))
	mutate.AppendMutation(mutation.GetLabelsMutation(OperatorNameLabel, c.GetName(), OperatorVersionLabel, c.GetVersion()))

	owner := owner.Get(ctx)
	if owner == nil {
		return nil, ErrorOwnerNotFound
	}

	mutate.AppendMutation(mutation.GetOwnerMutation(c.Scheme, owner))

	return mutate, nil
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

func (c *Controller) DeploymentMutateFn(ctx context.Context, dependencies ...graph.Resource) (resources.Mutable, error) {
	mutate, err := c.GlobalMutateFn(ctx)
	if err != nil {
		return nil, err
	}

	mutate.AppendMutation(mutation.GetTemplateAnnotationsMutation(WarningAnnotation, fmt.Sprintf(WarningValueTmpl, c.GetName())))
	mutate.AppendMutation(mutation.GetTemplateLabelsMutation(OperatorNameLabel, c.GetName(), OperatorVersionLabel, c.GetVersion()))

	fqdn := c.GetFQDN()

	mutate.AppendMutation(func(ctx context.Context, obj runtime.Object) error {
		for i, dep := range dependencies {
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
				return errors.Wrapf(err, "%d: cannot get dependency", i)
			}

			kind := strings.ToLower(depRemote.GetObjectKind().GroupVersionKind().Kind)

			err = mutation.GetTemplateAnnotationsMutation(
				fmt.Sprintf("%s.%s.%s.%s/uuid", name, namespace, kind, fqdn), string(depRemote.GetUID()),
				fmt.Sprintf("%s.%s.%s.%s/version", name, namespace, kind, fqdn), depRemote.GetResourceVersion(),
			)(ctx, obj)
			if err != nil {
				return errors.Wrapf(err, "%d: annotation mutation", i)
			}
		}

		return nil
	})

	return mutate, nil
}
