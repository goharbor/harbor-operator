package harbor

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/graph"
)

func (r *Reconciler) AddJobServiceConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (JobServiceInternalCertificate, JobServiceSecret, error) {
	certificate, err := r.AddJobServiceInternalCertificate(ctx, harbor, tlsIssuer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "certificate")
	}

	secret, err := r.AddJobServiceSecret(ctx, harbor)

	return certificate, secret, errors.Wrap(err, "secret")
}

type JobServiceSecret graph.Resource

func (r *Reconciler) AddJobServiceSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (JobServiceSecret, error) {
	secret, err := r.GetJobServiceSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	secretRes, err := r.AddSecretToManage(ctx, secret)

	return JobServiceSecret(secretRes), errors.Wrap(err, "add")
}

const (
	JobServiceSecretLength      = 16
	JobServiceSecretNumDigits   = 6
	JobServiceSecretNumSpecials = 6
)

func (r *Reconciler) GetJobServiceSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), controllers.JobService.String(), "secret")
	namespace := harbor.GetNamespace()

	secret, err := password.Generate(JobServiceSecretLength, JobServiceSecretNumDigits, JobServiceSecretNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varTrue,
		Type:      harbormetav1.SecretTypeSingle,
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: secret,
		},
	}, nil
}

type JobService graph.Resource

func (r *Reconciler) AddJobService(ctx context.Context, harbor *goharborv1alpha2.Harbor, certificate JobServiceInternalCertificate, core Core, coreSecret CoreSecret, jobServiceSecret JobServiceSecret) (JobService, error) {
	jobservice, err := r.GetJobService(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	jobserviceRes, err := r.AddBasicResource(ctx, jobservice, core, certificate, coreSecret, jobServiceSecret)

	return jobserviceRes, errors.Wrap(err, "add")
}

type JobServiceInternalCertificate graph.Resource

func (r *Reconciler) AddJobServiceInternalCertificate(ctx context.Context, harbor *goharborv1alpha2.Harbor, tlsIssuer InternalTLSIssuer) (JobServiceInternalCertificate, error) {
	cert, err := r.GetInternalTLSCertificate(ctx, harbor, harbormetav1.JobServiceTLS)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	certRes, err := r.Controller.AddCertificateToManage(ctx, cert, tlsIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return JobServiceInternalCertificate(certRes), nil
}

const (
	DefaultJobServiceLogSweeper = 14 * time.Hour
)

func (r *Reconciler) GetJobService(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.JobService, error) { // nolint:funlen
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	coreURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String()),
	}).String()
	coreSecretRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String(), "secret")
	registryAuthRef := r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String(), "basicauth")
	secretRef := r.NormalizeName(ctx, harbor.GetName(), controllers.JobService.String(), "secret")
	logLevel := harbor.Spec.LogLevel.JobService()
	registryURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.Registry.String()),
	}).String()
	serviceTokenURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   r.NormalizeName(ctx, harbor.GetName(), controllers.Core.String()),
		Path:   "/service/token",
	}).String()

	registryctlPort, err := harbor.Spec.InternalTLS.GetInternalPort(harbormetav1.RegistryControllerTLS)
	if err != nil {
		return nil, serrors.UnrecoverrableError(errors.Wrap(err, "cannot get registryController port"), serrors.OperatorReason, "unable to configure registry controller url")
	}

	registryControllerURL := (&url.URL{
		Scheme: harbor.Spec.InternalTLS.GetScheme(),
		Host:   fmt.Sprintf("%s:%d", r.NormalizeName(ctx, harbor.GetName(), controllers.RegistryController.String()), registryctlPort),
	}).String()

	redisDSN := harbor.Spec.RedisDSN(harbormetav1.JobServiceRedis)

	tls := harbor.Spec.InternalTLS.GetComponentTLSSpec(r.GetInternalTLSCertificateSecretName(ctx, harbor, harbormetav1.JobServiceTLS))

	return &goharborv1alpha2.JobService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: goharborv1alpha2.JobServiceSpec{
			ComponentSpec: harbor.Spec.Registry.ComponentSpec,
			Core: goharborv1alpha2.JobServiceCoreSpec{
				SecretRef: coreSecretRef,
				URL:       coreURL,
			},
			JobLoggers: goharborv1alpha2.JobServiceLoggerConfigSpec{
				Files: []goharborv1alpha2.JobServiceLoggerConfigFileSpec{{
					Level: logLevel,
					Sweeper: &metav1.Duration{
						Duration: DefaultJobServiceLogSweeper,
					},
				}},
			},
			Loggers: goharborv1alpha2.JobServiceLoggerConfigSpec{
				STDOUT: &goharborv1alpha2.JobServiceLoggerConfigSTDOUTSpec{
					Level: logLevel,
				},
			},
			WorkerPool: goharborv1alpha2.JobServicePoolSpec{
				WorkerCount: harbor.Spec.JobService.WorkerCount,
				Redis: goharborv1alpha2.JobServicePoolRedisSpec{
					OpacifiedDSN: redisDSN,
				},
			},
			Registry: goharborv1alpha2.RegistryControllerConnectionSpec{
				Credentials: goharborv1alpha2.CoreComponentsRegistryCredentialsSpec{
					PasswordRef: registryAuthRef,
					Username:    RegistryAuthenticationUsername,
				},
				RegistryURL:   registryURL,
				ControllerURL: registryControllerURL,
			},
			TokenService: goharborv1alpha2.JobServiceTokenSpec{
				URL: serviceTokenURL,
			},
			SecretRef: secretRef,
			TLS:       tls,
		},
	}, nil
}
