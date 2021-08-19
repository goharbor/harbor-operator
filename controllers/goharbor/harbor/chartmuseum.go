package harbor

import (
	"context"
	"net/url"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) AddChartMuseumConfigurations(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (ChartMuseumInternalCertificate, error) {
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

func (r *Reconciler) AddChartMuseumInternalCertificate(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (ChartMuseumInternalCertificate, error) {
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

func (r *Reconciler) AddChartMuseum(ctx context.Context, harbor *goharborv1.Harbor, certificate ChartMuseumInternalCertificate, coreSecret CoreSecret) (ChartMuseum, error) {
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

func (r *Reconciler) GetChartMuseum(ctx context.Context, harbor *goharborv1.Harbor) (*goharborv1.ChartMuseum, error) { //nolint:funlen
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	basicAuthRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "secret")
	debug := harbor.Spec.LogLevel == harbormetav1.HarborDebug

	redis := harbor.Spec.RedisConnection(harbormetav1.ChartMuseumRedis)

	publicURL, err := url.Parse(harbor.Spec.ExternalURL)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse externalURL")
	}

	chartServerURL := ""
	if harbor.Spec.ChartMuseum.AbsoluteURL {
		chartServerURL = publicURL.String()
	}

	publicURL.Path += "/chartrepo"
	maxStorageObjects := int64(0)
	parallelLimit := int32(0)

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.ChartMuseumTLS))

	return &goharborv1.ChartMuseum{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: version.SetVersion(map[string]string{
				harbormetav1.NetworkPoliciesAnnotationName: harbormetav1.NetworkPoliciesAnnotationDisabled,
			}, harbor.Spec.Version),
		},
		Spec: goharborv1.ChartMuseumSpec{
			ComponentSpec: harbor.GetComponentSpec(ctx, harbormetav1.ChartMuseumComponent),
			Authentication: goharborv1.ChartMuseumAuthSpec{
				AnonymousGet: false,
				BasicAuthRef: basicAuthRef,
			},
			Server: goharborv1.ChartMuseumServerSpec{
				TLS: tls,
			},
			Cache: goharborv1.ChartMuseumCacheSpec{
				Redis: &redis,
			},
			Chart: goharborv1.ChartMuseumChartSpec{
				AllowOverwrite: &varTrue,
				Storage: goharborv1.ChartMuseumChartStorageSpec{
					ChartMuseumChartStorageDriverSpec: r.ChartMuseumStorage(ctx, harbor),
					MaxStorageObjects:                 &maxStorageObjects,
				},
				Index: goharborv1.ChartMuseumChartIndexSpec{
					ParallelLimit: &parallelLimit,
				},
				URL: chartServerURL,
			},
			Log: goharborv1.ChartMuseumLogSpec{
				Debug: debug,
				JSON:  true,
			},
			CertificateInjection: harbor.Spec.ChartMuseum.CertificateInjection,
			Network:              harbor.Spec.Network,
		},
	}, nil
}
