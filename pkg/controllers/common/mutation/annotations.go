package mutation

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/resources"
)

func GetAnnotationsMutation(key, value string, kv ...string) resources.Mutable {
	return func(ctx context.Context, _, result runtime.Object) controllerutil.MutateFn {
		resourceMeta, ok := result.(metav1.Object)
		if !ok {
			logger.Get(ctx).Info("Cannot mutate annotations: unexpected resource type")
			return func() error { return nil }
		}

		return func() error {
			annotations := resourceMeta.GetAnnotations()
			if annotations == nil {
				annotations = map[string]string{}
			}

			annotations[key] = value

			for i := 0; i < len(kv); i += 2 {
				k, v := kv[i], kv[i+1]
				annotations[k] = v
			}

			resourceMeta.SetAnnotations(annotations)

			return nil
		}
	}
}

func GetTemplateAnnotationsMutation(key, value string, kv ...string) resources.Mutable {
	return func(ctx context.Context, _, result runtime.Object) controllerutil.MutateFn {
		resultDeployment, ok := result.(*appsv1.Deployment)
		if !ok {
			logger.Get(ctx).Info("Cannot mutate annotations: unexpected resource type")
			return func() error { return nil }
		}

		return func() error {
			annotations := resultDeployment.Spec.Template.GetAnnotations()
			if annotations == nil {
				annotations = map[string]string{}
			}

			annotations[key] = value

			for i := 0; i < len(kv); i += 2 {
				k, v := kv[i], kv[i+1]
				annotations[k] = v
			}

			resultDeployment.Spec.Template.SetAnnotations(annotations)

			return nil
		}
	}
}
