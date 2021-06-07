package portal_test

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("Portal", func() {
	var (
		ns     = test.InitNamespace(func() context.Context { return ctx })
		portal goharborv1.Portal
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		portal.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("portal"),
			Namespace: ns.GetName(),
			Annotations: test.AddVersionAnnotations(map[string]string{
				goharborv1.HarborClassAnnotation: className,
			}),
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, portal.GetName()),
			Namespace: portal.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			portal.Spec = goharborv1.PortalSpec{}
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &portal)).
					Should(test.SuccessOrExists)

				Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&portal), &portal) }, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(portal.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &portal, time.Minute, 5*time.Second)

				IntegTest(ctx, &portal)
			})

			By("Updating resource spec", func() {
				oldGeneration := portal.GetGeneration()

				test.ScaleUp(ctx, &portal)

				Ω(portal.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&portal), &portal)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &portal, time.Minute, 5*time.Second)

				IntegTest(ctx, &portal)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &portal)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&portal), &portal)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

const healthPath = "/"

func IntegTest(ctx context.Context, portal *goharborv1.Portal) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, portal.GetName()),
		Namespace: portal.GetNamespace(),
	}

	proxyReq := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%s", namespacedName.Name, harbormetav1.PortalHTTPPortName)).
		SubResource("proxy").
		Suffix(healthPath).
		MaxRetries(0)

	Eventually(func() error {
		return proxyReq.Do(ctx).Error()
	}).ShouldNot(HaveOccurred())
}
