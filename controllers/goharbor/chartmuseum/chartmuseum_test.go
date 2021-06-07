package chartmuseum_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/pods"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("ChartMuseum", func() {
	var (
		ns          = test.InitNamespace(func() context.Context { return ctx })
		chartMuseum goharborv1.ChartMuseum
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		chartMuseum.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("chartmuseum"),
			Namespace: ns.GetName(),
			Annotations: test.AddVersionAnnotations(map[string]string{
				goharborv1.HarborClassAnnotation: className,
			}),
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, chartMuseum.GetName()),
			Namespace: chartMuseum.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			chartMuseum.Spec = goharborv1.ChartMuseumSpec{
				Chart: goharborv1.ChartMuseumChartSpec{
					Storage: goharborv1.ChartMuseumChartStorageSpec{
						ChartMuseumChartStorageDriverSpec: goharborv1.ChartMuseumChartStorageDriverSpec{
							FileSystem: &goharborv1.ChartMuseumChartStorageDriverFilesystemSpec{
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
					URL: "http://the.chartserver.url",
				},
			}
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &chartMuseum)).
					Should(test.SuccessOrExists)

				Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&chartMuseum), &chartMuseum) }, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(chartMuseum.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &chartMuseum, time.Minute, 5*time.Second)

				IntegTest(ctx, &chartMuseum)
			})

			By("Updating resource spec", func() {
				oldGeneration := chartMuseum.GetGeneration()

				test.ScaleUp(ctx, &chartMuseum)

				Ω(chartMuseum.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&chartMuseum), &chartMuseum)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &chartMuseum, time.Minute, 5*time.Second)

				IntegTest(ctx, &chartMuseum)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &chartMuseum)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&chartMuseum), &chartMuseum)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

func IntegTest(ctx context.Context, chartMuseum *goharborv1.ChartMuseum) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, chartMuseum.GetName()),
		Namespace: chartMuseum.GetNamespace(),
	}

	proxyReq := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%s", namespacedName.Name, harbormetav1.ChartMuseumHTTPPortName)).
		SubResource("proxy").
		Suffix("health").
		MaxRetries(0)

	Eventually(func() ([]byte, error) {
		return proxyReq.DoRaw(ctx)
	}).
		Should(WithTransform(func(result []byte) bool {
			var health struct {
				Healthy bool `json:"healthy"`
			}

			Ω(json.Unmarshal(result, &health)).
				Should(Succeed())

			return health.Healthy
		}, BeTrue()))
}
