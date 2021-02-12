package goharbor_test

import (
	"context"

	. "github.com/onsi/gomega"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newChartMuseumController() controllerTest {
	return controllerTest{
		Setup:         setupValidChartMuseum,
		Update:        updateChartMuseum,
		GetStatusFunc: getChartMuseumStatusFunc,
	}
}

func setupValidChartMuseum(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	name := newName("chartmuseum")
	chartmuseum := &goharborv1alpha2.ChartMuseum{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.ChartMuseumSpec{
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
		},
	}

	Expect(k8sClient.Create(ctx, chartmuseum)).To(Succeed())

	return chartmuseum, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateChartMuseum(ctx context.Context, object Resource) {
	chartmuseum, ok := object.(*goharborv1alpha2.ChartMuseum)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if chartmuseum.Spec.Replicas != nil {
		replicas = *chartmuseum.Spec.Replicas + 1
	}

	chartmuseum.Spec.Replicas = &replicas
}

func getChartMuseumStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var chartmuseum goharborv1alpha2.ChartMuseum

		err := k8sClient.Get(ctx, key, &chartmuseum)

		Expect(err).ToNot(HaveOccurred())

		return chartmuseum.Status
	}
}
