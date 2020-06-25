package mutation

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

func GetOwnerMutation(scheme *runtime.Scheme, owner metav1.Object) resources.Mutable {
	return func(ctx context.Context, _, result runtime.Object) controllerutil.MutateFn {
		resourceMeta, ok := result.(metav1.Object)
		if !ok {
			logger.Get(ctx).Info("Cannot mutate owner: unexpected resource type")
			return func() error { return nil }
		}

		return func() error {
			err := controllerutil.SetControllerReference(owner, resourceMeta, scheme)
			return errors.Wrapf(err, "cannot set controller reference for %+v", result)
		}
	}
}
