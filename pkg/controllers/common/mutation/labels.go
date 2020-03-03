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

func GetLabelsMutation(key, value string, kv ...string) resources.Mutable {
	return func(ctx context.Context, _, result runtime.Object) controllerutil.MutateFn {
		resultMeta, ok := result.(metav1.Object)
		if !ok {
			logger.Get(ctx).Info("Cannot mutate labels: unexpected resource type")
			return func() error { return nil }
		}

		return func() error {
			labels := resultMeta.GetLabels()
			if labels == nil {
				labels = map[string]string{}
			}

			labels[key] = value

			for i := 0; i < len(kv); i += 2 {
				k, v := kv[i], kv[i+1]
				labels[k] = v
			}

			resultMeta.SetLabels(labels)

			return nil
		}
	}
}

func GetTemplateLabelsMutation(key, value string, kv ...string) resources.Mutable {
	return func(ctx context.Context, _, result runtime.Object) controllerutil.MutateFn {
		resultDeployment, ok := result.(*appsv1.Deployment)
		if !ok {
			logger.Get(ctx).Info("Cannot mutate labels: unexpected resource type")
			return func() error { return nil }
		}

		return func() error {
			labels := resultDeployment.Spec.Template.GetLabels()
			if labels == nil {
				labels = map[string]string{}
			}

			labels[key] = value

			for i := 0; i < len(kv); i += 2 {
				k, v := kv[i], kv[i+1]
				labels[k] = v
			}

			resultDeployment.Spec.Template.SetLabels(labels)

			return nil
		}
	}
}
