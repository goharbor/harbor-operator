package registryctl_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/pods"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("RegistryController", func() {
	var (
		ns          = test.InitNamespace(func() context.Context { return ctx })
		registryCtl goharborv1alpha2.RegistryController
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		registryCtl.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("registryctl"),
			Namespace: ns.GetName(),
			Annotations: map[string]string{
				goharborv1alpha2.HarborClassAnnotation: className,
			},
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, registryCtl.GetName()),
			Namespace: registryCtl.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			className, err := registryReconciler.GetClassName(ctx)
			Expect(err).ToNot(HaveOccurred())

			registry := &goharborv1alpha2.Registry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      test.NewName("registry"),
					Namespace: registryCtl.GetNamespace(),
					Annotations: map[string]string{
						goharborv1alpha2.HarborClassAnnotation: className,
					},
				},
				Spec: goharborv1alpha2.RegistrySpec{
					RegistryConfig01: goharborv1alpha2.RegistryConfig01{
						Storage: goharborv1alpha2.RegistryStorageSpec{
							Driver: goharborv1alpha2.RegistryStorageDriverSpec{
								InMemory: &goharborv1alpha2.RegistryStorageDriverInmemorySpec{},
							},
						},
					},
				},
			}

			Expect(test.GetClient(ctx).Create(ctx, registry)).To(Succeed())
			test.EnsureReady(ctx, registry, time.Minute, 5*time.Second)

			registryCtl.Spec = goharborv1alpha2.RegistryControllerSpec{
				RegistryRef: registry.GetName(),
			}
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &registryCtl)).
					Should(test.SuccessOrExists)

				Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&registryCtl), &registryCtl) }, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(registryCtl.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &registryCtl, time.Minute, 5*time.Second)

				IntegTest(ctx, &registryCtl)
			})

			By("Updating resource spec", func() {
				oldGeneration := registryCtl.GetGeneration()

				test.ScaleUp(ctx, &registryCtl)

				Ω(registryCtl.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&registryCtl), &registryCtl)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &registryCtl, time.Minute, 5*time.Second)

				IntegTest(ctx, &registryCtl)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &registryCtl)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&registryCtl), &registryCtl)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

const healthPath = "/api/health"

func IntegTest(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, registryCtl.GetName()),
		Namespace: registryCtl.GetNamespace(),
	}

	proxyReq := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%s", namespacedName.Name, harbormetav1.RegistryControllerHTTPPortName)).
		SubResource("proxy").
		Suffix(healthPath)

	Ω(proxyReq.Do(ctx).Error()).ShouldNot(HaveOccurred())
}
