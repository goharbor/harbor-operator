package goharbor_test

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/redis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newJobServiceController() controllerTest {
	return controllerTest{
		Setup:         setupValidJobService,
		Update:        updateJobService,
		GetStatusFunc: getJobServiceStatusFunc,
	}
}

func setupJobServiceResourceDependencies(ctx context.Context, ns string) string {
	registrySecret := newName("registry")

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registrySecret,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: "registry-password",
		},
		Type: harbormetav1.SecretTypeSingle,
	})).To(Succeed())

	return registrySecret
}

func setupValidJobService(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	registrySecret := setupJobServiceResourceDependencies(ctx, ns)

	coreResource, _ := setupValidCore(ctx, ns)
	redis := redis.New(ctx, ns)

	core := coreResource.(*goharborv1.Core)

	name := newName("jobservice")
	jobService := &goharborv1.JobService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: test.AddVersionAnnotations(nil),
		},
		Spec: goharborv1.JobServiceSpec{
			Core: goharborv1.JobServiceCoreSpec{
				URL:       fmt.Sprintf("http://%s-core", core.GetName()),
				SecretRef: core.Spec.SecretRef,
			},
			WorkerPool: goharborv1.JobServicePoolSpec{
				Redis: goharborv1.JobServicePoolRedisSpec{
					RedisConnection: redis,
				},
			},
			SecretRef: core.Spec.Components.JobService.SecretRef,
			Registry: goharborv1.RegistryControllerConnectionSpec{
				ControllerURL: "http://the.registryctl.url",
				RegistryURL:   "http://the.registry.url",
				Credentials: goharborv1.CoreComponentsRegistryCredentialsSpec{
					PasswordRef: registrySecret,
				},
			},
			JobLoggers: goharborv1.JobServiceLoggerConfigSpec{
				STDOUT: &goharborv1.JobServiceLoggerConfigSTDOUTSpec{},
			},
			Loggers: goharborv1.JobServiceLoggerConfigSpec{
				STDOUT: &goharborv1.JobServiceLoggerConfigSTDOUTSpec{},
			},
			TokenService: goharborv1.JobServiceTokenSpec{
				URL: "http://the.tokenservice.url",
			},
		},
	}

	Expect(k8sClient.Create(ctx, jobService)).To(Succeed())

	return jobService, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateJobService(ctx context.Context, object Resource) {
	jobService, ok := object.(*goharborv1.JobService)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if jobService.Spec.Replicas != nil {
		replicas = *jobService.Spec.Replicas + 1
	}

	jobService.Spec.Replicas = &replicas
}

func getJobServiceStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var jobService goharborv1.JobService

		err := k8sClient.Get(ctx, key, &jobService)

		Expect(err).ToNot(HaveOccurred())

		return jobService.Status
	}
}
