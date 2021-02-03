package pods

import (
	"context"
	"fmt"
	"strings"

	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

type Pods []corev1.Pod

func List(ctx context.Context, deployment types.NamespacedName) Pods {
	config := test.NewRestConfig(ctx)
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

	client, err = rest.UnversionedRESTClientFor(test.NewRestConfig(ctx))
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

	return Pods(pods.Items)
}

func (pods Pods) Ready(ctx context.Context) Pods {
	var result []corev1.Pod

	for _, pod := range pods {
		pod := pod

		ok, err := statuscheck.BasicCheck(ctx, &pod)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		if ok {
			result = append(result, pod)
		}
	}

	return Pods(result)
}

func (pods Pods) Latest(ctx context.Context) *corev1.Pod {
	var latest *corev1.Pod

	for _, pod := range pods {
		pod := pod

		if latest == nil {
			latest = &pod

			break
		}

		if pod.CreationTimestamp.After(latest.CreationTimestamp.Time) {
			latest = &pod
		}
	}

	return latest
}
