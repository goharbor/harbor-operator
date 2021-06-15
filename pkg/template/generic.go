package template

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ErrKeyNotFound struct {
	Key       string
	Reference string
}

func (err *ErrKeyNotFound) Error() string {
	return fmt.Sprintf("%s not found for reference %s", err.Key, err.Reference)
}

func GetK8SNamespacedDataFunc(ctx context.Context, c client.Client, namespace string, object client.Object, getData func(context.Context, client.Object) (map[string]interface{}, error), ignoreNotFound bool) interface{} {
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

				return nil, &ErrKeyNotFound{key, reference}
			}

			return result, nil
		default:
			results := make(map[string]interface{}, len(data))

			for _, key := range keys {
				result, ok := data[key]
				if !ok {
					if !ignoreNotFound {
						return nil, &ErrKeyNotFound{key, reference}
					}

					result = ""
				}

				results[key] = result
			}

			return results, nil
		}
	}
}
