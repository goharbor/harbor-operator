package pods

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func (pods Pods) Logs(ctx context.Context) map[string][]byte {
	config := test.GetRestConfig(ctx)
	config.APIPath = "apis"
	config.GroupVersion = &appsv1.SchemeGroupVersion

	client, err := rest.UnversionedRESTClientFor(config)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	results := make(map[string][]byte, len(pods))

	for _, pod := range pods {
		result, err := client.Get().
			Resource("pods").
			Namespace(pod.GetNamespace()).
			Name(pod.GetName()).
			SubResource("log").
			Param("pretty", "true").
			DoRaw(ctx)
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
			for name, logs := range List(*ctx, name()).Logs(*ctx) {
				fmt.Fprintf(ginkgo.GinkgoWriter, "\n### Logs of %s ###\n%s\n", name, string(logs))
			}
		})
	}
}
