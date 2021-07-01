package goharbor_test

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/certificate"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/postgresql"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/redis"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newCoreController() controllerTest {
	return controllerTest{
		Setup:         setupValidCore,
		Update:        updateCore,
		GetStatusFunc: getCoreStatusFunc,
	}
}

func setupCoreResourceDependencies(ctx context.Context, ns string) (string, string, string, string, string, string, string) {
	encryption := newName("encryption")
	csrf := newName("csrf")
	registryCtl := newName("registryctl")
	admin := newName("admin-password")
	core := newName("core-secret")
	jobservice := newName("jobservice-secret")
	tokenCert := newName("token-certificate")

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      encryption,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: "1234567890123456",
		},
		Type: harbormetav1.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csrf,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.CSRFSecretKey: "12345678901234567890123456789012",
		},
		Type: harbormetav1.SecretTypeCSRF,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryCtl,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: "the-registryctl-password",
		},
		Type: harbormetav1.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      admin,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: "Harbor12345",
		},
		Type: harbormetav1.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      core,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: "unsecure-core-secret",
		},
		Type: harbormetav1.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobservice,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: "unsecure-jobservice-secret",
		},
		Type: harbormetav1.SecretTypeSingle,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenCert,
			Namespace: ns,
		},
		Data: certificate.NewCA().NewCert().ToMap(),
		Type: harbormetav1.SecretTypeSingle,
	})).To(Succeed())

	return encryption, csrf, registryCtl, admin, core, jobservice, tokenCert
}

func setupValidCore(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	encryptionKeyName, csrfKey, registryCtlPassword, adminPassword, coreSecret, jobserviceSecret, tokenCertificate := setupCoreResourceDependencies(ctx, ns)

	database := postgresql.New(ctx, ns)
	redis := redis.New(ctx, ns)

	name := newName("core")
	core := &goharborv1.Core{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: test.AddVersionAnnotations(nil),
		},
		Spec: goharborv1.CoreSpec{
			Database: goharborv1.CoreDatabaseSpec{
				PostgresConnectionWithParameters: database,
				EncryptionKeyRef:                 encryptionKeyName,
			},
			CSRFKeyRef: csrfKey,
			CoreConfig: goharborv1.CoreConfig{
				AdminInitialPasswordRef: adminPassword,
				SecretRef:               coreSecret,
			},
			ExternalEndpoint: "https://the.public.url",
			Components: goharborv1.CoreComponentsSpec{
				TokenService: goharborv1.CoreComponentsTokenServiceSpec{
					URL:            "https://the.public.url/service/token",
					CertificateRef: tokenCertificate,
				},
				Registry: goharborv1.CoreComponentsRegistrySpec{
					RegistryControllerConnectionSpec: goharborv1.RegistryControllerConnectionSpec{
						ControllerURL: "http://the.registryctl.url",
						RegistryURL:   "http://the.registry.url",
						Credentials: goharborv1.CoreComponentsRegistryCredentialsSpec{
							Username:    "admin",
							PasswordRef: registryCtlPassword,
						},
					},
					Redis: &harbormetav1.RedisConnection{
						RedisHostSpec: harbormetav1.RedisHostSpec{
							Host: "registry-redis",
						},
						Database: 2,
					},
				},
				JobService: goharborv1.CoreComponentsJobServiceSpec{
					URL:       "http://the.jobservice.url",
					SecretRef: jobserviceSecret,
				},
				Portal: goharborv1.CoreComponentPortalSpec{
					URL: "https://the.public.url",
				},
			},
			Redis: goharborv1.CoreRedisSpec{
				RedisConnection: redis,
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
	core, ok := object.(*goharborv1.Core)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if core.Spec.Replicas != nil {
		replicas = *core.Spec.Replicas + 1
	}

	core.Spec.Replicas = &replicas
}

func getCoreStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var core goharborv1.Core

		err := k8sClient.Get(ctx, key, &core)

		Expect(err).ToNot(HaveOccurred())

		return core.Status
	}
}
