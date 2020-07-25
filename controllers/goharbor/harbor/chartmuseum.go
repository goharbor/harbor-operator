package harbor

import (
	"context"
	"net/url"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
)

func (r *Reconciler) AddChartMuseumConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor) (ChartMuseumAuthSecret, error) {
	if harbor.Spec.ChartMuseum == nil {
		return nil, nil
	}

	authSecret, err := r.AddChartMuseumAuthenticationSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "authentication secret")
	}

	return authSecret, nil
}

type ChartMuseumAuthSecret graph.Resource

func (r *Reconciler) AddChartMuseumAuthenticationSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (RegistryAuthSecret, error) {
	authSecret, err := r.GetChartMuseumAuthenticationSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get secret")
	}

	authSecretRes, err := r.AddSecretToManage(ctx, authSecret)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add secret")
	}

	return ChartMuseumAuthSecret(authSecretRes), nil
}

const (
	ChartMuseumAuthenticationUsername = "chart_controller"

	ChartMuseumAuthenticationPasswordLength      = 32
	ChartMuseumAuthenticationPasswordNumDigits   = 10
	ChartMuseumAuthenticationPasswordNumSpecials = 10
)

func (r *Reconciler) GetChartMuseumAuthenticationSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), "chartmuseum", "basicauth")
	namespace := harbor.GetNamespace()

	password, err := password.Generate(ChartMuseumAuthenticationPasswordLength, ChartMuseumAuthenticationPasswordNumDigits, ChartMuseumAuthenticationPasswordNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate password")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varFalse,
		Type:      goharborv1alpha2.SecretTypeHTPasswd,
		StringData: map[string]string{
			corev1.BasicAuthUsernameKey: ChartMuseumAuthenticationUsername,
			corev1.BasicAuthPasswordKey: password,
		},
	}, nil
}

type ChartMuseum graph.Resource

func (r *Reconciler) AddChartMuseum(ctx context.Context, harbor *goharborv1alpha2.Harbor) (ChartMuseum, error) {
	if harbor.Spec.ChartMuseum == nil {
		return nil, nil
	}

	chartmuseum, err := r.GetChartMuseum(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get chartmuseum")
	}

	chartmuseumRes, err := r.AddBasicResource(ctx, chartmuseum)

	return ChartMuseum(chartmuseumRes), errors.Wrap(err, "cannot add basic resource")
}

func (r *Reconciler) GetChartMuseum(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.ChartMuseum, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	basicAuthRef := r.NormalizeName(ctx, harbor.GetName(), "chartmuseum", "basicauth")
	debug := harbor.Spec.LogLevel == goharborv1alpha2.HarborDebug

	redisDSN := harbor.Spec.RedisDSN(goharborv1alpha2.ChartMuseumRedis)

	publicURL, err := url.Parse(harbor.Spec.ExternalURL)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parseexternalURL")
	}

	publicURL.Path += "/chartrepo"
	maxStorageObjects := int64(0)
	parallelLimit := int32(0)

	return &goharborv1alpha2.ChartMuseum{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.ChartMuseumSpec{
			ComponentSpec: harbor.Spec.ChartMuseum.ComponentSpec,
			Auth: goharborv1alpha2.ChartMuseumAuthSpec{
				AnonymousGet: false,
				BasicAuthRef: basicAuthRef,
			},
			Cache: goharborv1alpha2.ChartMuseumCacheSpec{
				Redis: &redisDSN,
			},
			Chart: goharborv1alpha2.ChartMuseumChartSpec{
				AllowOvewrite: &varTrue,
				Storage: goharborv1alpha2.ChartMuseumChartStorageSpec{
					ChartMuseumChartStorageDriverSpec: harbor.Spec.Persistence.ImageChartStorage.ChartMuseum(),
					MaxStorageObjects:                 &maxStorageObjects,
				},
				Index: goharborv1alpha2.ChartMuseumChartIndexSpec{
					ParallelLimit: &parallelLimit,
				},
				URL: publicURL.String(),
			},
			Log: goharborv1alpha2.ChartMuseumLogSpec{
				Debug: debug,
				JSON:  true,
			},
		},
	}, nil
}
