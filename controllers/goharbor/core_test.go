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

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

func newCoreController() controllerTest {
	return controllerTest{
		Setup:         setupValidCore,
		Update:        updateCore,
		GetStatusFunc: getCoreStatusFunc,
	}
}

func setupCoreResourceDependencies(ctx context.Context, ns string) (string, string, string, string, string, string, string, string, string) {
	database := newName("database")
	encryption := newName("encryption")
	csrf := newName("csrf")
	registryCtl := newName("registryctl")
	admin := newName("admin-password")
	core := newName("core-secret")
	jobservice := newName("jobservice-secret")
	tokenCert := newName("token-certificate")
	redis := newName("redis")

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      database,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.PostgresqlPasswordKey: "pg-password",
		},
		Type: goharborv1alpha2.SecretTypePostgresql,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      encryption,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: "1234567890123456",
		},
		Type: goharborv1alpha2.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csrf,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.CSRFSecretKey: "12345678901234567890123456789012",
		},
		Type: goharborv1alpha2.SecretTypeCSRF,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryCtl,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: "the-registryctl-password",
		},
		Type: goharborv1alpha2.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      admin,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: "Harbor12345",
		},
		Type: goharborv1alpha2.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      core,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: "unsecure-core-secret",
		},
		Type: goharborv1alpha2.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobservice,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: "unsecure-jobservice-secret",
		},
		Type: goharborv1alpha2.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenCert,
			Namespace: ns,
		},
		Data: generateCertificate(),
		Type: goharborv1alpha2.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.RedisPasswordKey: "the-redis-password",
		},
		Type: goharborv1alpha2.SecretTypeRedis,
	})).To(Succeed())

	return database, encryption, csrf, registryCtl, admin, core, jobservice, tokenCert, redis
}

func setupValidCore(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	databasePasswordName, encryptionKeyName, csrfKey, registryCtlPassword, adminPassword, coreSecret, jobserviceSecret, tokenCertificate, redisPassword := setupCoreResourceDependencies(ctx, ns)

	name := newName("core")
	core := &goharborv1alpha2.Core{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.CoreSpec{
			Database: goharborv1alpha2.CoreDatabaseSpec{
				CorePostgresqlSpec: goharborv1alpha2.CorePostgresqlSpec{
					ExternalDatabaseSpec: goharborv1alpha2.ExternalDatabaseSpec{
						Host:        "the.database.url",
						PasswordRef: databasePasswordName,
					},
					Name: "core",
				},
				EncryptionKeyRef: encryptionKeyName,
			},
			CSRFKeyRef: csrfKey,
			CoreConfig: goharborv1alpha2.CoreConfig{
				AdminInitialPasswordRef: adminPassword,
				SecretRef:               coreSecret,
			},
			ExternalEndpoint: "https://the.public.url",
			ServiceToken: goharborv1alpha2.CoreServiceTokenSpec{
				CertificateRef: tokenCertificate,
			},
			Components: goharborv1alpha2.CoreComponentsSpec{
				TokenService: goharborv1alpha2.CoreComponentsTokenServiceSpec{
					URL: "https://the.public.url/service/token",
				},
				Registry: goharborv1alpha2.CoreComponentsRegistrySpec{
					ControllerURL: "http://the.registryctl.url",
					URL:           "http://the.registry.url",
					Redis: goharborv1alpha2.OpacifiedDSN{
						DSN:         "redis://registry-redis/2",
						PasswordRef: redisPassword,
					},
					Credentials: goharborv1alpha2.CoreComponentsRegistryCredentialsSpec{
						Username:    "admin",
						PasswordRef: registryCtlPassword,
					},
				},
				JobService: goharborv1alpha2.CoreComponentsJobServiceSpec{
					URL:       "http://the.jobservice.url",
					SecretRef: jobserviceSecret,
				},
				Portal: goharborv1alpha2.CoreComponentPortalSpec{
					URL: "https://the.public.url",
				},
			},
			Redis: goharborv1alpha2.CoreRedisSpec{
				OpacifiedDSN: goharborv1alpha2.OpacifiedDSN{
					DSN:         "redis://the.redis.url/0",
					PasswordRef: redisPassword,
				},
			},
		},
	}
	Expect(k8sClient.Create(ctx, core)).To(Succeed())

	return core, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateCore(ctx context.Context, object Resource) {
	core, ok := object.(*goharborv1alpha2.Core)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if core.Spec.Replicas != nil {
		replicas = *core.Spec.Replicas + 1
	}

	core.Spec.Replicas = &replicas
}

func getCoreStatusFunc(ctx context.Context, key client.ObjectKey) func() goharborv1alpha2.ComponentStatus {
	return func() goharborv1alpha2.ComponentStatus {
		var core goharborv1alpha2.Core

		err := k8sClient.Get(ctx, key, &core)

		Expect(err).ToNot(HaveOccurred())

		return core.Status
	}
}
