package trivy_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/pods"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/redis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("Trivy", func() {
	var (
		ns    = test.InitNamespace(func() context.Context { return ctx })
		trivy goharborv1.Trivy
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		trivy.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("trivy"),
			Namespace: ns.GetName(),
			Annotations: test.AddVersionAnnotations(map[string]string{
				goharborv1.HarborClassAnnotation: className,
			}),
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, trivy.GetName()),
			Namespace: trivy.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			namespace := trivy.GetNamespace()

			trivy.Spec = goharborv1.TrivySpec{
				Redis: goharborv1.TrivyRedisSpec{
					RedisConnection: redis.New(ctx, namespace),
				},
			}
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &trivy)).
					Should(test.SuccessOrExists)

				Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&trivy), &trivy) }, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(trivy.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &trivy, time.Minute, 5*time.Second)

				IntegTest(ctx, &trivy)
			})

			By("Updating resource spec", func() {
				oldGeneration := trivy.GetGeneration()

				test.ScaleUp(ctx, &trivy)

				Ω(trivy.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&trivy), &trivy)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &trivy, time.Minute, 5*time.Second)

				IntegTest(ctx, &trivy)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &trivy)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&trivy), &trivy)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

const healthPath = "/probe/ready"

func IntegTest(ctx context.Context, trivy *goharborv1.Trivy) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, trivy.GetName()),
		Namespace: trivy.GetNamespace(),
	}

	proxyReq := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%s", namespacedName.Name, harbormetav1.TrivyHTTPPortName)).
		SubResource("proxy").
		Suffix(healthPath).
		MaxRetries(0)

	Eventually(func() error {
		return proxyReq.Do(ctx).Error()
	}).ShouldNot(HaveOccurred())
}
