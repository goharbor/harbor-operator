package registry

import (
	"context"
	"text/template"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *Reconciler) GetDataFuncFromArraySecret(ctx context.Context, getter func(int) (interface{}, types.NamespacedName, error), itemsCount int) (func(interface{}) (map[string]string, error), error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getDataFuncFromSecret")
	defer span.Finish()

	datas := make(map[interface{}]map[string]string, itemsCount)
	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < itemsCount; i++ {
		index := i

		result := map[string]string{}
		datas[index] = result

		g.Go(func() error {
			index, key, err := getter(index)
			if err != nil {
				return errors.Wrapf(err, "index %v", index)
			}

			if key.Name != "" {
				span, ctx := opentracing.StartSpanFromContext(ctx, "getHookDataFunc", opentracing.Tags{
					"reference": key,
				})
				defer span.Finish()

				var secret corev1.Secret

				err := r.Client.Get(ctx, key, &secret)
				if err != nil {
					return errors.Wrapf(err, "%v", key)
				}

				for key, value := range secret.Data {
					result[key] = string(value)
				}

				return nil
			}

			return nil
		})
	}

	err := g.Wait()

	return func(index interface{}) (map[string]string, error) {
		data, ok := datas[index]
		if !ok {
			return nil, errors.Errorf("no data found for %v", index)
		}

		return data, nil
	}, err
}

func (r *Reconciler) GetHookDataFunc(ctx context.Context, registry *goharborv1.Registry) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "hookDataFunc")
	defer span.Finish()

	namespace := registry.GetNamespace()

	return r.GetDataFuncFromArraySecret(ctx, func(index int) (interface{}, types.NamespacedName, error) {
		return index, types.NamespacedName{
			Namespace: namespace,
			Name:      registry.Spec.Log.Hooks[index].OptionsRef,
		}, nil
	}, len(registry.Spec.Log.Hooks))
}

func (r *Reconciler) GetReportingDataFunc(ctx context.Context, registry *goharborv1.Registry) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "registryMiddleware")
	defer span.Finish()

	namespace := registry.GetNamespace()
	reportingCount := len(registry.Spec.Reporting)

	indexes := make([]string, reportingCount)
	i := 0

	for name := range registry.Spec.Reporting {
		indexes[i] = name
		i++
	}

	return r.GetDataFuncFromArraySecret(ctx, func(index int) (interface{}, types.NamespacedName, error) {
		return indexes[index], types.NamespacedName{
			Namespace: namespace,
			Name:      registry.Spec.Reporting[indexes[index]],
		}, nil
	}, reportingCount)
}

func (r *Reconciler) GetRegistryMiddlewareDataFunc(ctx context.Context, registry *goharborv1.Registry) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "registryMiddlewareDataFunc")
	defer span.Finish()

	namespace := registry.GetNamespace()

	return r.GetDataFuncFromArraySecret(ctx, func(index int) (interface{}, types.NamespacedName, error) {
		return index, types.NamespacedName{
			Namespace: namespace,
			Name:      registry.Spec.Middlewares.Registry[index].OptionsRef,
		}, nil
	}, len(registry.Spec.Middlewares.Registry))
}

func (r *Reconciler) GetRepositoryMiddlewareDataFunc(ctx context.Context, registry *goharborv1.Registry) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repositoryMiddlewareDataFunc")
	defer span.Finish()

	namespace := registry.GetNamespace()

	return r.GetDataFuncFromArraySecret(ctx, func(index int) (interface{}, types.NamespacedName, error) {
		return index, types.NamespacedName{
			Namespace: namespace,
			Name:      registry.Spec.Middlewares.Repository[index].OptionsRef,
		}, nil
	}, len(registry.Spec.Middlewares.Repository))
}

func (r *Reconciler) GetStorageMiddlewareDataFunc(ctx context.Context, registry *goharborv1.Registry) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "storageMiddlewareDataFunc")
	defer span.Finish()

	namespace := registry.GetNamespace()

	return r.GetDataFuncFromArraySecret(ctx, func(index int) (interface{}, types.NamespacedName, error) {
		return index, types.NamespacedName{
			Namespace: namespace,
			Name:      registry.Spec.Middlewares.Storage[index].OptionsRef,
		}, nil
	}, len(registry.Spec.Middlewares.Storage))
}

func (r *Reconciler) GetConfigFuncs(ctx context.Context, registry *goharborv1.Registry) (template.FuncMap, error) {
	var hookDataFunc, storageDataFunc, reportingData, registryMiddlewareData, repositoryMiddlewareData, storageMiddlewareData interface{}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		hookDataFunc, err = r.GetHookDataFunc(gctx, registry)

		return errors.Wrap(err, "hook")
	})

	g.Go(func() error {
		var err error
		reportingData, err = r.GetReportingDataFunc(gctx, registry)

		return errors.Wrap(err, "reporting")
	})

	g.Go(func() error {
		var err error
		registryMiddlewareData, err = r.GetRegistryMiddlewareDataFunc(gctx, registry)

		return errors.Wrap(err, "registry middleware")
	})

	g.Go(func() error {
		var err error
		repositoryMiddlewareData, err = r.GetRepositoryMiddlewareDataFunc(gctx, registry)

		return errors.Wrap(err, "repository middleware")
	})

	g.Go(func() error {
		var err error
		storageMiddlewareData, err = r.GetStorageMiddlewareDataFunc(gctx, registry)

		return errors.Wrap(err, "storage middleware")
	})

	err := g.Wait()

	return template.FuncMap{
		"hooksData":                hookDataFunc,
		"storageData":              storageDataFunc,
		"reportingData":            reportingData,
		"registryMiddlewareData":   registryMiddlewareData,
		"repositoryMiddlewareData": repositoryMiddlewareData,
		"storageMiddlewareData":    storageMiddlewareData,
	}, err
}
