package harbor

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (r *Reconciler) AddChartMuseumConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (ChartMuseumInternalCertificate, error) {
	if harbor.Spec.ChartMuseum == nil {
		return nil, nil
	}

	certificate, err := r.AddChartMuseumInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "certificate")
	}

	return certificate, nil
}

type ChartMuseumInternalCertificate graph.Resource

func (r *Reconciler) AddChartMuseumInternalCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (ChartMuseumInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.ChartMuseumTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return ChartMuseumInternalCertificate(certRes), nil
}

const (
	ChartMuseumAuthenticationUsername = "chart_controller"
)

type ChartMuseum graph.Resource

func (r *Reconciler) AddChartMuseum(ctx context.Context, harbor *goharborv1alpha2.Harbor, certificate ChartMuseumInternalCertificate, coreSecret CoreSecret) (ChartMuseum, error) {
	if harbor.Spec.ChartMuseum == nil {
		return nil, nil
	}

	chartmuseum, err := r.GetChartMuseum(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	chartmuseumRes, err := r.AddBasicResource(ctx, chartmuseum, certificate, coreSecret)

	return ChartMuseum(chartmuseumRes), errors.Wrap(err, "add")
}

func (r *Reconciler) GetChartMuseum(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.ChartMuseum, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	basicAuthRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "secret")
	debug := harbor.Spec.LogLevel == harbormetav1.HarborDebug

	redis := harbor.Spec.RedisConnection(harbormetav1.ChartMuseumRedis)

	publicURL, err := url.Parse(harbor.Spec.ExternalURL)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parseexternalURL")
	}

	publicURL.Path += "/chartrepo"
	maxStorageObjects := int64(0)
	parallelLimit := int32(0)

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.ChartMuseumTLS))

	return &goharborv1alpha2.ChartMuseum{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.ChartMuseumSpec{
			ComponentSpec: harbor.Spec.ChartMuseum.ComponentSpec,
			Authentication: goharborv1alpha2.ChartMuseumAuthSpec{
				AnonymousGet: false,
				BasicAuthRef: basicAuthRef,
			},
			Server: goharborv1alpha2.ChartMuseumServerSpec{
				TLS: tls,
			},
			Cache: goharborv1alpha2.ChartMuseumCacheSpec{
				Redis: &redis,
			},
			Chart: goharborv1alpha2.ChartMuseumChartSpec{
				AllowOvewrite: &varTrue,
				Storage: goharborv1alpha2.ChartMuseumChartStorageSpec{
					ChartMuseumChartStorageDriverSpec: r.ChartMuseumStorage(ctx, harbor),
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
