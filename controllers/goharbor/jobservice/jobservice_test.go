package jobservice_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	harborcore "github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/harbor-core"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/pods"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/redis"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/pkg/config/inmemory"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("JobService", func() {
	var (
		ns         = test.InitNamespace(func() context.Context { return ctx })
		jobservice goharborv1.JobService
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		jobservice.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("jobservice"),
			Namespace: ns.GetName(),
			Annotations: test.AddVersionAnnotations(map[string]string{
				goharborv1.HarborClassAnnotation: className,
			}),
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, jobservice.GetName()),
			Namespace: jobservice.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			namespace := jobservice.GetNamespace()
			coreName := test.NewName("core")
			tokenServiceName := coreName
			registryName := test.NewName("registry")
			registryControllerName := test.NewName("registryctl")

			coreConfig := inmemory.NewInMemoryManager()

			coreConfig.Set(context.TODO(), common.AUTHMode, common.DBAuth)
			harborcore.DeployDatabase(ctx, namespace, coreConfig)

			coreConfig.Set(context.TODO(), common.CoreLocalURL, fmt.Sprintf("http://%s", test.NewName("core")))
			coreConfig.Set(context.TODO(), common.MetricPath, "/metrics")
			coreConfig.Set(context.TODO(), common.MetricPort, 8080)
			coreConfig.Set(context.TODO(), common.MetricEnable, false)

			jobservice.Spec = goharborv1.JobServiceSpec{
				SecretRef: test.NewName("secret"),
				Core: goharborv1.JobServiceCoreSpec{
					SecretRef: test.NewName("core"),
					URL:       harborcore.New(ctx, namespace, coreConfig).String(),
				},
				TokenService: goharborv1.JobServiceTokenSpec{
					URL: fmt.Sprintf("http://%s", tokenServiceName),
				},
				WorkerPool: goharborv1.JobServicePoolSpec{
					Redis: goharborv1.JobServicePoolRedisSpec{
						RedisConnection: redis.New(ctx, namespace),
					},
				},
				Registry: goharborv1.RegistryControllerConnectionSpec{
					Credentials: goharborv1.CoreComponentsRegistryCredentialsSpec{
						Username:    "jobservice",
						PasswordRef: test.NewName("registry"),
					},
					RegistryURL:   fmt.Sprintf("http://%s", registryName),
					ControllerURL: fmt.Sprintf("http://%s", registryControllerName),
				},
				JobLoggers: goharborv1.JobServiceLoggerConfigSpec{
					STDOUT: &goharborv1.JobServiceLoggerConfigSTDOUTSpec{
						Level: harbormetav1.JobServiceInfo,
					},
				},
				Loggers: goharborv1.JobServiceLoggerConfigSpec{
					STDOUT: &goharborv1.JobServiceLoggerConfigSTDOUTSpec{
						Level: harbormetav1.JobServiceInfo,
					},
				},
			}

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      jobservice.Spec.SecretRef,
				},
				StringData: map[string]string{
					harbormetav1.SharedSecretKey: "the-password",
				},
				Type: harbormetav1.SecretTypeSingle,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      jobservice.Spec.Registry.Credentials.PasswordRef,
				},
				StringData: map[string]string{
					harbormetav1.SharedSecretKey: "password4registry",
				},
				Type: harbormetav1.SecretTypeSingle,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      jobservice.Spec.Core.SecretRef,
				},
				StringData: map[string]string{
					harbormetav1.SharedSecretKey: "password4core",
				},
				Type: harbormetav1.SecretTypeSingle,
			})).To(Succeed())
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &jobservice)).
					Should(test.SuccessOrExists)

				Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&jobservice), &jobservice) }, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(jobservice.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &jobservice, time.Minute, 5*time.Second)

				IntegTest(ctx, &jobservice)
			})

			By("Updating resource spec", func() {
				oldGeneration := jobservice.GetGeneration()

				test.ScaleUp(ctx, &jobservice)

				Ω(jobservice.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&jobservice), &jobservice)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &jobservice, time.Minute, 5*time.Second)

				IntegTest(ctx, &jobservice)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &jobservice)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&jobservice), &jobservice)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

const healthPath = "/api/v1/stats"

type Worker struct {
	ID          string   `json:"worker_pool_id"`
	StartedAT   int      `json:"started_at"`
	HeartbeatAt int      `json:"heartbeat_at"`
	JobName     []string `json:"job_names"`
	Concurrency int      `json:"concurrency"`
	Status      string   `json:"status"`
}

type HealthResponse struct {
	WorkerPools []Worker `json:"worker_pools"`
}

func IntegTest(ctx context.Context, js *goharborv1.JobService) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, js.GetName()),
		Namespace: js.GetNamespace(),
	}

	proxyReq := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%s", namespacedName.Name, harbormetav1.CoreHTTPPortName)).
		SubResource("proxy").
		Suffix(healthPath).
		MaxRetries(0)

	Eventually(func() ([]byte, error) {
		return proxyReq.DoRaw(ctx)
	}).
		Should(WithTransform(func(result []byte) []Worker {
			var health HealthResponse

			Ω(json.Unmarshal(result, &health)).
				Should(Succeed())

			return health.WorkerPools
		}, MatchElements(func(element interface{}) string { return "A" }, AllowDuplicates, Elements{
			"A": MatchFields(IgnoreExtras, Fields{
				"Status": Equal("Healthy"),
			}),
		})))
}
