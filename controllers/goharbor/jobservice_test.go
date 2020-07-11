/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package goharbor_test

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

func newJobServiceController() controllerTest {
	return controllerTest{
		Setup:         setupValidJobService,
		Update:        updateJobService,
		GetStatusFunc: getJobServiceStatusFunc,
	}
}

func setupJobServiceResourceDependencies(ctx context.Context, ns string) (string, string) {
	redisSecret := newName("redis")
	registrySecret := newName("registry")

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisSecret,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.RedisPasswordKey: "redis-password",
		},
		Type: goharborv1alpha2.SecretTypeRedis,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registrySecret,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: "registry-password",
		},
		Type: goharborv1alpha2.SecretTypeSingle,
	})).To(Succeed())

	return redisSecret, registrySecret
}
func setupValidJobService(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	redisSecret, registrySecret := setupJobServiceResourceDependencies(ctx, ns)

	coreResource, _ := setupValidCore(ctx, ns)

	core := coreResource.(*goharborv1alpha2.Core)

	name := newName("jobservice")
	jobService := &goharborv1alpha2.JobService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.JobServiceSpec{
			Core: goharborv1alpha2.JobServiceCoreSpec{
				URL:       fmt.Sprintf("http://%s", core.GetName()),
				SecretRef: core.Spec.SecretRef,
			},
			WorkerPool: goharborv1alpha2.JobServicePoolSpec{
				Redis: goharborv1alpha2.JobServicePoolRedisSpec{
					OpacifiedDSN: goharborv1alpha2.OpacifiedDSN{
						DSN:         "redis://the.redis.url",
						PasswordRef: redisSecret,
					},
				},
			},
			SecretRef: core.Spec.Components.JobService.SecretRef,
			Registry: goharborv1alpha2.CoreComponentsRegistryCredentialsSpec{
				PasswordRef: registrySecret,
			},
			JobLoggers: goharborv1alpha2.JobServiceLoggerConfigSpec{
				STDOUT: &goharborv1alpha2.JobServiceLoggerConfigSTDOUTSpec{},
			},
			Loggers: goharborv1alpha2.JobServiceLoggerConfigSpec{
				STDOUT: &goharborv1alpha2.JobServiceLoggerConfigSTDOUTSpec{},
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
	jobService, ok := object.(*goharborv1alpha2.JobService)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if jobService.Spec.Replicas != nil {
		replicas = *jobService.Spec.Replicas + 1
	}

	jobService.Spec.Replicas = &replicas
}

func getJobServiceStatusFunc(ctx context.Context, key client.ObjectKey) func() goharborv1alpha2.ComponentStatus {
	return func() goharborv1alpha2.ComponentStatus {
		var jobService goharborv1alpha2.JobService

		err := k8sClient.Get(ctx, key, &jobService)

		Expect(err).ToNot(HaveOccurred())

		return jobService.Status
	}
}
