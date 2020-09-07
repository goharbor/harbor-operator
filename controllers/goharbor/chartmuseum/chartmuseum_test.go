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

package chartmuseum_test

import (
	"bytes"
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/html"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("ChartMuseum", func() {
	var (
		ns          = test.InitNamespace(func() context.Context { return ctx })
		chartMuseum goharborv1alpha2.ChartMuseum
	)

	BeforeEach(func() {
		chartMuseum.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("chartmuseum"),
			Namespace: ns.GetName(),
		}
	})

	JustAfterEach(test.LogOnFailureFunc(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, chartMuseum.GetName()),
			Namespace: chartMuseum.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			chartMuseum.Spec = goharborv1alpha2.ChartMuseumSpec{
				Chart: goharborv1alpha2.ChartMuseumChartSpec{
					Storage: goharborv1alpha2.ChartMuseumChartStorageSpec{
						ChartMuseumChartStorageDriverSpec: goharborv1alpha2.ChartMuseumChartStorageDriverSpec{
							FileSystem: &goharborv1alpha2.ChartMuseumChartStorageDriverFilesystemSpec{
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
					URL: "https://the.chartserver.url",
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

				Eventually(func() (int64, error) {
					err := test.GetClient(ctx).Create(ctx, &chartMuseum)
					if err != nil {
						return 0, err
					}

					return chartMuseum.Status.ObservedGeneration, nil
				}, 5*time.Second).
					Should(Equal(chartMuseum.GetGeneration()), "ObservedGeneration should be updated")

				test.EnsureReady(ctx, &chartMuseum, time.Minute, 5*time.Second)

				IntegTest(ctx, &chartMuseum)
			})

			By("Updating resource spec", func() {
				test.ScaleUp(ctx, &chartMuseum)

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&chartMuseum), &chartMuseum)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &chartMuseum, time.Minute, 5*time.Second)

				IntegTest(ctx, &chartMuseum)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &chartMuseum)).Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&chartMuseum), &chartMuseum)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

func IntegTest(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) {
	client, err := rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
	Expect(err).ToNot(HaveOccurred())

	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, chartMuseum.GetName()),
		Namespace: chartMuseum.GetNamespace(),
	}

	/*
		Eventually(func() error {
			endpoint := &corev1.Endpoints{}

			err := test.GetClient(ctx).Get(ctx, namespacedName, endpoint)
			if err != nil {
				return err
			}

			ok, err := statuscheck.EndpointCheck(ctx, endpoint, harbormetav1.ChartMuseumHTTPPortName)
			if err != nil {
				return err
			}

			if !ok {
				return errors.New("not ready") // nolint:goerr113
			}

			return nil
		}, 30*time.Second, 100*time.Millisecond).Should(Succeed())

		var pods corev1.PodList
		Ω(
			client.Get().
				Resource("pods").
				Namespace(chartMuseum.GetNamespace()).
				Param("labelSelector", strings.Join([]string{
					fmt.Sprintf("%s=%s", reconciler.Label("name"), reconciler.NormalizeName(ctx, chartMuseum.Name)),
					fmt.Sprintf("%s=%s", controller.OperatorNameLabel, reconciler.GetName()),
					fmt.Sprintf("%s=%s", controller.OperatorVersionLabel, reconciler.GetVersion()),
				}, ",")).
				Do(ctx).
				Into(&pods)).
			Should(Succeed())

		Ω(pods.Items).ShouldNot(HaveLen(0))
	*/

	result, err := client.Get().
		Resource("services").
		Namespace(namespacedName.Namespace).
		Name(fmt.Sprintf("%s:%d", namespacedName.Name, harbormetav1.HTTPPort)).
		SubResource("proxy").
		Suffix("/").
		DoRaw(ctx)
	Ω(err).Should(Succeed())

	Ω(html.Parse(bytes.NewReader(result))).
		Should(WithTransform(func(node *html.Node) string {
			return node.FirstChild.NextSibling.FirstChild.FirstChild.NextSibling.FirstChild.Data
		}, Equal("Welcome to ChartMuseum!")))
}
