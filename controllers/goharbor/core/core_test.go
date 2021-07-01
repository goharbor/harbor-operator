package core_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/certificate"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/pods"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/postgresql"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/redis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("Core", func() {
	var (
		ns   = test.InitNamespace(func() context.Context { return ctx })
		core goharborv1.Core
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		core.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("core"),
			Namespace: ns.GetName(),
			Annotations: test.AddVersionAnnotations(map[string]string{
				goharborv1.HarborClassAnnotation: className,
			}),
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, core.GetName()),
			Namespace: core.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			namespace := core.GetNamespace()

			tokenServiceName := test.NewName("token-service")
			jobserviceName := test.NewName("jobservice")
			portalName := test.NewName("portal")
			registryName := test.NewName("registry")
			registryControllerName := test.NewName("registryctl")

			core.Spec = goharborv1.CoreSpec{
				Components: goharborv1.CoreComponentsSpec{
					TokenService: goharborv1.CoreComponentsTokenServiceSpec{
						URL:            fmt.Sprintf("http://%s", tokenServiceName),
						CertificateRef: tokenServiceName,
					},
					JobService: goharborv1.CoreComponentsJobServiceSpec{
						URL:       fmt.Sprintf("http://%s", jobserviceName),
						SecretRef: jobserviceName,
					},
					Portal: goharborv1.CoreComponentPortalSpec{
						URL: fmt.Sprintf("http://%s", portalName),
					},
					Registry: goharborv1.CoreComponentsRegistrySpec{
						RegistryControllerConnectionSpec: goharborv1.RegistryControllerConnectionSpec{
							RegistryURL:   fmt.Sprintf("http://%s", registryName),
							ControllerURL: fmt.Sprintf("http://%s", registryControllerName),
							Credentials: goharborv1.CoreComponentsRegistryCredentialsSpec{
								PasswordRef: test.NewName("registry-core"),
								Username:    "core",
							},
						},
					},
				},
				ExternalEndpoint: fmt.Sprintf("http://%s", core.GetName()),
				Redis: goharborv1.CoreRedisSpec{
					RedisConnection: redis.New(ctx, namespace),
				},
				CoreConfig: goharborv1.CoreConfig{
					SecretRef:               namespace,
					AdminInitialPasswordRef: test.NewName("initial-password"),
				},
				CSRFKeyRef: test.NewName("csrf"),
				Database: goharborv1.CoreDatabaseSpec{
					EncryptionKeyRef:                 test.NewName("encryption-key"),
					PostgresConnectionWithParameters: postgresql.New(ctx, core.GetNamespace()),
				},
			}

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      core.Spec.Components.TokenService.CertificateRef,
				},
				Data: certificate.NewCA().NewCert(tokenServiceName).ToMap(),
				Type: corev1.SecretTypeTLS,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      core.Spec.Components.JobService.SecretRef,
				},
				StringData: map[string]string{
					harbormetav1.SharedSecretKey: "the-key",
				},
				Type: harbormetav1.SecretTypeSingle,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      core.Spec.Components.Registry.Credentials.PasswordRef,
				},
				StringData: map[string]string{
					harbormetav1.SharedSecretKey: "the-registry-password",
				},
				Type: harbormetav1.SecretTypeSingle,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      core.Spec.CSRFKeyRef,
				},
				StringData: map[string]string{
					harbormetav1.CSRFSecretKey: strings.Repeat("a", 16),
				},
				Type: harbormetav1.CSRFSecretKey,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      core.Spec.AdminInitialPasswordRef,
				},
				StringData: map[string]string{
					harbormetav1.SharedSecretKey: "Harbor12345",
				},
				Type: harbormetav1.SecretTypeSingle,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      core.Spec.Database.EncryptionKeyRef,
				},
				StringData: map[string]string{
					harbormetav1.SharedSecretKey: "my-encryption-key",
				},
				Type: harbormetav1.SecretTypeSingle,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      core.Spec.SecretRef,
				},
				StringData: map[string]string{
					harbormetav1.SharedSecretKey: "the-password",
				},
				Type: harbormetav1.SecretTypeSingle,
			})).To(Succeed())
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &core)).
					Should(test.SuccessOrExists)

				Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&core), &core) }, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(core.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &core, time.Minute, 5*time.Second)

				IntegTest(ctx, &core)
			})

			By("Updating resource spec", func() {
				oldGeneration := core.GetGeneration()

				test.ScaleUp(ctx, &core)

				Ω(core.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&core), &core)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &core, time.Minute, 5*time.Second)

				IntegTest(ctx, &core)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &core)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&core), &core)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

const healthPath = "api/v2.0/health"

func IntegTest(ctx context.Context, core *goharborv1.Core) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, core.GetName()),
		Namespace: core.GetNamespace(),
	}

	proxyReq := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%s", namespacedName.Name, harbormetav1.CoreHTTPPortName)).
		SubResource("proxy").
		Suffix(healthPath).
		MaxRetries(0)

	type ComponentStatus struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}

	Eventually(func() ([]byte, error) {
		return proxyReq.DoRaw(ctx)
	}).
		Should(WithTransform(func(result []byte) []ComponentStatus {
			var health struct {
				Status     string            `json:"status"`
				Components []ComponentStatus `json:"components"`
			}

			Ω(json.Unmarshal(result, &health)).
				Should(Succeed())

			return health.Components
		}, WithTransform(func(components []ComponentStatus) string {
			for _, component := range components {
				if component.Name == "core" {
					return component.Status
				}
			}

			panic("core component not found")
		}, Equal("healthy"))))
}
