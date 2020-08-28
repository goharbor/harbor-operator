package mutation

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type MetaGetter func(metav1.Object) map[string]string

type MetaSetter func(metav1.Object, map[string]string)

func GetMetaMutation(getter MetaGetter, setter MetaSetter, key, value string, kv ...string) resources.Mutable {
	return func(ctx context.Context, _, result runtime.Object) controllerutil.MutateFn {
		resultMeta, ok := result.(metav1.Object)
		if !ok {
			logger.Get(ctx).Info("Cannot mutate meta: unexpected resource type")

			return func() error { return nil }
		}

		return func() error {
			data := getter(resultMeta)
			if data == nil {
				data = map[string]string{}
			}

			data[key] = value

			for i := 0; i < len(kv); i += 2 {
				k, v := kv[i], kv[i+1]
				data[k] = v
			}

			setter(resultMeta, data)

			return nil
		}
	}
}

func GetTemplateMetaMutation(getter MetaGetter, setter MetaSetter, key, value string, kv ...string) resources.Mutable {
	return func(ctx context.Context, _, result runtime.Object) controllerutil.MutateFn {
		resultDeployment, ok := result.(*appsv1.Deployment)
		if !ok {
			logger.Get(ctx).Info("Cannot mutate meta: unexpected resource type")

			return func() error { return nil }
		}

		return func() error {
			data := getter(&resultDeployment.Spec.Template)
			if data == nil {
				data = map[string]string{}
			}

			data[key] = value

			for i := 0; i < len(kv); i += 2 {
				k, v := kv[i], kv[i+1]
				data[k] = v
			}

			setter(&resultDeployment.Spec.Template, data)

			return nil
		}
	}
}
