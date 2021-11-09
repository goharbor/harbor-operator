package registry_test

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

var _ = Describe("Registry", func() {
	var (
		ns       = test.InitNamespace(func() context.Context { return ctx })
		registry goharborv1.Registry
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		registry.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("registry"),
			Namespace: ns.GetName(),
			Annotations: test.AddVersionAnnotations(map[string]string{
				goharborv1.HarborClassAnnotation: className,
			}),
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, registry.GetName()),
			Namespace: registry.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			registry.Spec = goharborv1.RegistrySpec{
				RegistryConfig01: goharborv1.RegistryConfig01{
					Storage: goharborv1.RegistryStorageSpec{
						Driver: goharborv1.RegistryStorageDriverSpec{
							InMemory: &goharborv1.RegistryStorageDriverInmemorySpec{},
						},
					},
				},
			}

			registryCtl := &goharborv1.RegistryController{
				ObjectMeta: metav1.ObjectMeta{
					Name:        registry.GetName(),
					Namespace:   registry.GetNamespace(),
					Annotations: test.AddVersionAnnotations(nil),
				},
				Spec: goharborv1.RegistryControllerSpec{
					RegistryRef: registry.GetName(),
				},
			}
			Ω(test.GetClient(ctx).Create(ctx, registryCtl)).
				Should(test.SuccessOrExists)
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &registry)).
					Should(test.SuccessOrExists)

				Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&registry), &registry) }, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(registry.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &registry, time.Minute, 5*time.Second)

				IntegTest(ctx, &registry)
			})

			By("Updating resource spec", func() {
				oldGeneration := registry.GetGeneration()

				test.ScaleUp(ctx, &registry)

				Ω(registry.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&registry), &registry)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &registry, time.Minute, 5*time.Second)

				IntegTest(ctx, &registry)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &registry)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&registry), &registry)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

const healthPath = "/"

func IntegTest(ctx context.Context, registry *goharborv1.Registry) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, registry.GetName()),
		Namespace: registry.GetNamespace(),
	}

	proxyReq := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%s", namespacedName.Name, harbormetav1.RegistryAPIPortName)).
		SubResource("proxy").
		Suffix(healthPath).
		MaxRetries(0)

	Eventually(func() error {
		return proxyReq.Do(ctx).Error()
	}).ShouldNot(HaveOccurred())
}
