package harbor

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Exporter graph.Resource

func (r *Reconciler) AddExporter(ctx context.Context, harbor *goharborv1.Harbor, core Core, certificate ExporterInternalCertificate) (Exporter, error) {
	if harbor.Spec.Exporter == nil {
		return nil, nil
	}

	exporter, err := r.GetExporter(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	exporterRes, err := r.AddBasicResource(ctx, exporter, core, certificate)

	return Exporter(exporterRes), errors.Wrap(err, "add")
}

func (r *Reconciler) AddExporterConfigurations(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (ExporterInternalCertificate, error) {
	if harbor.Spec.Exporter == nil {
		return nil, nil
	}

	certificate, err := r.AddExporterInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "certificate")
	}

	return certificate, nil
}

type ExporterInternalCertificate graph.Resource

func (r *Reconciler) AddExporterInternalCertificate(ctx context.Context, harbor *goharborv1.Harbor, tlsIssuer InternalTLSIssuer) (ExporterInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.ExporterTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return ExporterInternalCertificate(certRes), nil
}

func (r *Reconciler) GetExporter(ctx context.Context, harbor *goharborv1.Harbor) (*goharborv1.Exporter, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.ExporterTLS))

	postgresConn, err := harbor.Spec.Database.GetPostgresqlConnection(harbormetav1.ExporterComponent)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get database configuration")
	}

	encryptionKeyRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "encryptionkey")

	return &goharborv1.Exporter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: version.SetVersion(map[string]string{
				harbormetav1.NetworkPoliciesAnnotationName: harbormetav1.NetworkPoliciesAnnotationDisabled,
			}, harbor.Spec.Version),
		},
		Spec: goharborv1.ExporterSpec{
			ComponentSpec: harbor.GetComponentSpec(ctx, harbormetav1.ExporterComponent),
			Port:          harbor.Spec.Exporter.Port,
			Path:          harbor.Spec.Exporter.Path,
			TLS:           tls,
			Log: goharborv1.ExporterLogSpec{
				Level: harbor.Spec.LogLevel.Exporter(),
			},
			Cache: goharborv1.ExporterCacheSpec{
				Duration:      harbor.Spec.Exporter.Cache.Duration,
				CleanInterval: harbor.Spec.Exporter.Cache.CleanInterval,
			},
			Core: goharborv1.ExporterCoreSpec{
				URL: r.getCoreURL(ctx, harbor),
			},
			Database: goharborv1.ExporterDatabaseSpec{
				PostgresConnectionWithParameters: *postgresConn,
				EncryptionKeyRef:                 encryptionKeyRef,
			},
			JobService: &goharborv1.ExporterJobServiceSpec{
				Redis: &goharborv1.JobServicePoolRedisSpec{
					RedisConnection: harbor.Spec.RedisConnection(harbormetav1.JobServiceRedis),
				},
			},
			Network: harbor.Spec.Network,
		},
	}, nil
}
