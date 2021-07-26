package goharbor_test

import (
	"context"
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/redis"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newExporterController() controllerTest {
	return controllerTest{
		Setup:         setupValidExporter,
		Update:        updateExporter,
		GetStatusFunc: getExporterStatusFunc,
	}
}

func setupValidExporter(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	coreRes, _ := setupValidCore(ctx, ns)
	core := coreRes.(*goharborv1.Core)

	redis := redis.New(ctx, ns)

	var replicas int32 = 1

	name := newName("exporter")

	exporter := &goharborv1.Exporter{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: test.AddVersionAnnotations(nil),
		},

		Spec: goharborv1.ExporterSpec{
			ComponentSpec: harbormetav1.ComponentSpec{
				Replicas: &replicas,
			},
			Core: goharborv1.ExporterCoreSpec{
				URL: fmt.Sprintf("http://%s-core", core.GetName()),
			},
			Database: goharborv1.ExporterDatabaseSpec{
				PostgresConnectionWithParameters: core.Spec.Database.PostgresConnectionWithParameters,
				EncryptionKeyRef:                 core.Spec.Database.EncryptionKeyRef,
			},
			JobService: &goharborv1.ExporterJobServiceSpec{
				Redis: &goharborv1.JobServicePoolRedisSpec{
					RedisConnection: redis,
				},
			},
		},
	}

	Expect(k8sClient.Create(ctx, exporter)).To(Succeed())

	return exporter, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateExporter(ctx context.Context, object Resource) {
	exporter, ok := object.(*goharborv1.Exporter)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if exporter.Spec.Replicas != nil {
		replicas = *exporter.Spec.Replicas + 1
	}

	exporter.Spec.Replicas = &replicas
}

func getExporterStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var exporter goharborv1.Exporter

		err := k8sClient.Get(ctx, key, &exporter)

		Expect(err).ToNot(HaveOccurred())

		return exporter.Status
	}
}
