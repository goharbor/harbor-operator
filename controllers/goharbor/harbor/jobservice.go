package harbor

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
)

type JobServiceSecret graph.Resource

func (r *Reconciler) AddJobServiceConfigurations(ctx context.Context, harbor *goharborv1alpha2.Harbor) (JobServiceSecret, error) {
	secret, err := r.GetJobServiceSecret(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get secret")
	}

	secretRes, err := r.AddSecretToManage(ctx, secret)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add secret")
	}

	return JobServiceSecret(secretRes), nil
}

const (
	JobServiceSecretLength      = 16
	JobServiceSecretNumDigits   = 6
	JobServiceSecretNumSpecials = 6
)

func (r *Reconciler) GetJobServiceSecret(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*corev1.Secret, error) {
	name := r.NormalizeName(ctx, harbor.GetName(), "jobservice", "secret")
	namespace := harbor.GetNamespace()

	secret, err := password.Generate(JobServiceSecretLength, JobServiceSecretNumDigits, JobServiceSecretNumSpecials, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate secret")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Immutable: &varTrue,
		Type:      goharborv1alpha2.SecretTypeSingle,
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: secret,
		},
	}, nil
}

type JobService graph.Resource

func (r *Reconciler) AddJobService(ctx context.Context, harbor *goharborv1alpha2.Harbor, core Core, coreSecret CoreSecret, jobServiceSecret JobServiceSecret) (JobService, error) {
	jobservice, err := r.GetJobService(ctx, harbor)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get jobservice")
	}

	jobserviceRes, err := r.AddBasicResource(ctx, jobservice, core, coreSecret, jobServiceSecret)

	return jobserviceRes, errors.Wrap(err, "cannot add basic resource")
}

const (
	DefaultJobServiceLogSweeper = 14 * time.Hour
)

func (r *Reconciler) GetJobService(ctx context.Context, harbor *goharborv1alpha2.Harbor) (*goharborv1alpha2.JobService, error) {
	name := r.NormalizeName(ctx, harbor.GetName())
	namespace := harbor.GetNamespace()

	coreURL := fmt.Sprintf("http://%s", r.NormalizeName(ctx, harbor.GetName(), "core"))
	coreSecretRef := r.NormalizeName(ctx, harbor.GetName(), "core", "secret")
	registryAuthRef := r.NormalizeName(ctx, harbor.GetName(), "registry", "basicauth")
	secretRef := r.NormalizeName(ctx, harbor.GetName(), "jobservice", "secret")

	logLevel := harbor.Spec.LogLevel.JobService()

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
					OpacifiedDSN: harbor.Spec.RedisDSN(goharborv1alpha2.JobServiceRedis),
				},
			},
			Registry: goharborv1alpha2.CoreComponentsRegistryCredentialsSpec{
				PasswordRef: registryAuthRef,
				Username:    RegistryAuthenticationUsername,
			},
			SecretRef: secretRef,
		},
	}, nil
}
