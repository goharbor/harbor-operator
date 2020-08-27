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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
)

const (
	defaultGenerationNumber int64 = 1
)

var (
	ns = SetupTest()
)

var _ = Describe("ChartMuseum", func() {
	var chartMuseum *goharborv1alpha2.ChartMuseum

	BeforeEach(func() {
		chartMuseum = &goharborv1alpha2.ChartMuseum{
			ObjectMeta: metav1.ObjectMeta{
				Name:      test.NewName("chartmuseum"),
				Namespace: ns.GetName(),
			},
		}
	})

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
			By("Creating new resource")

			Expect(test.GetClient(ctx).Create(ctx, chartMuseum)).
				To(Succeed())

			Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(chartMuseum), chartMuseum) }, time.Minute, 5*time.Second).
				Should(Succeed(), "resource should exists")

			Expect(chartMuseum.GetGeneration()).
				Should(Equal(defaultGenerationNumber), "ObservedGeneration should not be updated")

			test.EnsureReady(ctx, chartMuseum, time.Minute, 5*time.Second)

			IntegTest(ctx, chartMuseum)

			By("Updating resource spec")

			test.ScaleUp(ctx, chartMuseum)

			Expect(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(chartMuseum), chartMuseum)).
				To(Succeed(), "resource should still be accessible")

			test.EnsureReady(ctx, chartMuseum, time.Minute, 5*time.Second)

			IntegTest(ctx, chartMuseum)

			By("Deleting resource")

			Expect(test.GetClient(ctx).Delete(ctx, chartMuseum)).To(Succeed())

			Eventually(func() error { return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(chartMuseum), chartMuseum) }, time.Minute, 5*time.Second).
				ShouldNot(Succeed(), "Resource should no more exist")
		})
	})
})

func IntegTest(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) {
	config := rest.CopyConfig(cfg)
	config.APIPath = "api"
	config = rest.AddUserAgent(config, fmt.Sprintf("%s(%s)", application.GetName(ctx), application.GetVersion(ctx)))
	config.NegotiatedSerializer = serializer.NewCodecFactory(test.GetScheme(ctx))
	config.GroupVersion = &corev1.SchemeGroupVersion

	client, err := rest.UnversionedRESTClientFor(config)
	Expect(err).ToNot(HaveOccurred())

	name := reconciler.NormalizeName(ctx, chartMuseum.GetName())
	namespace := ns.GetName()

	Eventually(func() error {
		return test.GetClient(ctx).Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}, &corev1.Endpoints{})
	}, 5*time.Second, 200*time.Millisecond).Should(Succeed())

	time.Sleep(10 * time.Second)

	result, err := client.Get().
		Resource("services").
		Namespace(namespace).
		Name(fmt.Sprintf("%s:%d", name, harbormetav1.HTTPPort)).
		SubResource("proxy").
		Suffix("/").
		DoRaw(ctx)

	Expect(err).ToNot(HaveOccurred())

	_, err = html.Parse(bytes.NewReader(result))
	Expect(err).ToNot(HaveOccurred())
}
