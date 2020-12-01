package mutation

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetOwnerMutation(scheme *runtime.Scheme, owner metav1.Object) resources.Mutable {
	return func(ctx context.Context, result runtime.Object) error {
		resourceMeta, ok := result.(metav1.Object)
		if !ok {
			return ErrorResourceType
		}

		err := controllerutil.SetControllerReference(owner, resourceMeta, scheme)

		return errors.Wrapf(err, "cannot set controller reference for %+v", result)
	}
}
