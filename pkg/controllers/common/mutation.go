package common

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cmutation "github.com/goharbor/harbor-operator/pkg/controllers/common/mutation"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

const (
	WarningAnnotation = "containerregistry.ovhcloud.com/warning"
	WarningValueTmpl  = "⚠️ This Resource is managed by *%s* ⚠️"
)

const (
	OperatorNameLabel    = "containerregistry.ovhcloud.com/controller"
	OperatorVersionLabel = "containerregistry.ovhcloud.com/version"
	OwnerLabel           = "containerregistry.ovhcloud.com/owner"
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
			mutation.AppendMutation(cmutation.GetLabelsMutation(OwnerLabel, owner.GetName()))
		}

		mutation := mutation(ctx, resource, result)

		return func() (err error) {
			return mutation()
		}
	}
}
func (c *Controller) DeploymentMutateFn(ctx context.Context) resources.Mutable {
	var mutation resources.Mutable = c.GlobalMutateFn(ctx)

	mutation.AppendMutation(cmutation.GetTemplateAnnotationsMutation(WarningAnnotation, fmt.Sprintf(WarningValueTmpl, c.GetName())))
	mutation.AppendMutation(cmutation.GetTemplateLabelsMutation(OperatorNameLabel, c.GetName(), OperatorVersionLabel, c.GetVersion()))

	return mutation
}
