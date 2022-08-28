package mutation

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetAffinity(ctx context.Context, c client.Client, owner Getter) resources.Mutable {
	return func(ctx context.Context, o runtime.Object) error {
		var (
			setter      Setter
			defaultFunc func(*corev1.Affinity, *corev1.Affinity) *corev1.Affinity
		)

		switch obj := o.(type) {
		case Setter:
			setter = obj
			defaultFunc = func(targetValue, defaultValue *corev1.Affinity) *corev1.Affinity {
				if targetValue != nil {
					return targetValue
				}

				return defaultValue
			}

		case *appsv1.Deployment:
			setter = &deployment{Deployment: obj}
			defaultFunc = func(_, defaultValue *corev1.Affinity) *corev1.Affinity {
				// The default value comes from its owner
				// so the deployment only uses the default value
				return defaultValue
			}
		default:
			// This type is not supported
			return nil
		}

		remote := setter.DeepCopyObject().(Setter)

		if err := c.Get(ctx, client.ObjectKeyFromObject(remote), remote); err != nil {
			if apierrors.IsNotFound(err) {
				setter.SetAffinity(owner.GetAffinity())

				return nil
			}

			return err
		}

		setter.SetAffinity(defaultFunc(remote.GetAffinity(), owner.GetAffinity()))

		return nil
	}
}

type deployment struct {
	*appsv1.Deployment
}

func (d *deployment) SetAffinity(affinity *corev1.Affinity) {
	d.Spec.Template.Spec.Affinity = affinity
}

func (d *deployment) GetAffinity() *corev1.Affinity {
	return d.Spec.Template.Spec.Affinity
}

type Getter interface {
	client.Object
	GetAffinity() *corev1.Affinity
}

type Setter interface {
	Getter
	SetAffinity(*corev1.Affinity)
}
