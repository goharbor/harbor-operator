package template

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetK8SNamespacedDataFunc(ctx context.Context, c client.Client, namespace string, object runtime.Object, getData func(context.Context, runtime.Object) (map[string]interface{}, error), ignoreNotFound bool) interface{} {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getK8SNamespacedDataFunc")
	defer span.Finish()

	return func(reference string, keys ...string) (interface{}, error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "getK8SNamespacedData", opentracing.Tags{
			"namespace": namespace,
			"reference": reference,
		})
		defer span.Finish()

		err := c.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      reference,
		}, object)
		if err != nil {
			return nil, errors.Wrapf(err, "%v", reference)
		}

		data, err := getData(ctx, object)
		if err != nil {
			return nil, errors.Wrapf(err, "%v", reference)
		}

		switch len(keys) {
		case 0:
			return data, nil
		case 1:
			key := keys[0]

			result, ok := data[key]
			if !ok {
				if ignoreNotFound {
					return "", nil
				}

				return nil, errors.Errorf("%s not found for reference %s", key, reference)
			}

			return result, nil
		default:
			results := make(map[string]interface{}, len(data))

			for _, key := range keys {
				result, ok := data[key]
				if !ok {
					if !ignoreNotFound {
						return nil, errors.Errorf("%s not found for reference %s", key, reference)
					}

					result = ""
				}

				results[key] = result
			}

			return results, nil
		}
	}
}
