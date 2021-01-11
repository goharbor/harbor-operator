package test

import (
	"context"
	"fmt"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func Logs(ctx context.Context, deployment types.NamespacedName) map[string][]byte {
	config := NewRestConfig(ctx)
	config.APIPath = "apis"
	config.GroupVersion = &appsv1.SchemeGroupVersion

	client, err := rest.UnversionedRESTClientFor(config)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	var deploymentResource appsv1.Deployment

	gomega.Expect(
		client.Get().
			Resource("deployments").
			Namespace(deployment.Namespace).
			Name(deployment.Name).
			Do(ctx).
			Into(&deploymentResource)).
		To(gomega.Succeed())

	gomega.Expect(deploymentResource.Spec.Selector.MatchLabels).ToNot(gomega.HaveLen(0))

	labelSelectors := make([]string, 0, len(deploymentResource.Spec.Selector.MatchLabels))
	for label, value := range deploymentResource.Spec.Selector.MatchLabels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", label, value))
	}

	client, err = rest.UnversionedRESTClientFor(NewRestConfig(ctx))
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	var pods corev1.PodList

	gomega.Expect(
		client.Get().
			Resource("pods").
			Namespace(deploymentResource.GetNamespace()).
			Param("labelSelector", strings.Join(labelSelectors, ",")).
			Do(ctx).Into(&pods)).
		To(gomega.Succeed())

	gomega.Expect(pods.Items).ToNot(gomega.HaveLen(0))

	results := make(map[string][]byte, len(pods.Items))

	for _, pod := range pods.Items {
		result, err := client.Get().
			Resource("pods").
			Namespace(pod.GetNamespace()).
			Name(pod.GetName()).
			SubResource("log").
			Param("pretty", "true").
			DoRaw(ctx)
		//gomega.Expect(err).ToNot(gomega.HaveOccurred())
		if err != nil {
			results[pod.GetName()] = []byte(fmt.Sprintf("%v: status %s", err, pod.Status.Phase))
			continue
		}

		results[pod.GetName()] = result
	}

	return results
}

func LogsAll(ctx *context.Context, name func() types.NamespacedName) interface{} {
	return func(done ginkgo.Done) {
		defer close(done)

		if !ginkgo.CurrentGinkgoTestDescription().Failed {
			return
		}

		defer ginkgo.GinkgoRecover()

		ginkgo.By("Fetching logs after failure", func() {
			for name, logs := range Logs(*ctx, name()) {
				fmt.Fprintf(ginkgo.GinkgoWriter, "\n### Logs of %s ###\n%s\n", name, string(logs))
			}
		})
	}
}
