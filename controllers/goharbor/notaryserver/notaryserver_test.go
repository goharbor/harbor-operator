package notaryserver_test

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
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/postgresql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("NotaryServer", func() {
	var (
		ns           = test.InitNamespace(func() context.Context { return ctx })
		notaryserver goharborv1.NotaryServer
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		notaryserver.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("notaryserver"),
			Namespace: ns.GetName(),
			Annotations: test.AddVersionAnnotations(map[string]string{
				goharborv1.HarborClassAnnotation: className,
			}),
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, notaryserver.GetName()),
			Namespace: notaryserver.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			namespace := notaryserver.GetNamespace()

			notaryserver.Spec = goharborv1.NotaryServerSpec{
				Storage: goharborv1.NotaryStorageSpec{
					Postgres: postgresql.New(ctx, namespace),
				},
			}
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &notaryserver)).
					Should(test.SuccessOrExists)

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&notaryserver), &notaryserver)
				}, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(notaryserver.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &notaryserver, time.Minute, 5*time.Second)

				IntegTest(ctx, &notaryserver)
			})

			By("Updating resource spec", func() {
				oldGeneration := notaryserver.GetGeneration()

				test.ScaleUp(ctx, &notaryserver)

				Ω(notaryserver.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&notaryserver), &notaryserver)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &notaryserver, time.Minute, 5*time.Second)

				IntegTest(ctx, &notaryserver)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &notaryserver)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&notaryserver), &notaryserver)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

const healthPath = "/_notary_server/health"

func IntegTest(ctx context.Context, notaryserver *goharborv1.NotaryServer) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, notaryserver.GetName()),
		Namespace: notaryserver.GetNamespace(),
	}

	proxyReq := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%s", namespacedName.Name, harbormetav1.NotaryServerAPIPortName)).
		SubResource("proxy").
		Suffix(healthPath).
		MaxRetries(0)

	Eventually(func() error {
		return proxyReq.Do(ctx).Error()
	}).ShouldNot(HaveOccurred())
}
