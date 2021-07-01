package goharbor_test

import (
	"context"

	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/postgresql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newNotaryServerController() controllerTest {
	return controllerTest{
		Setup:         setupValidNotaryServer,
		Update:        updateNotaryServer,
		GetStatusFunc: getNotaryServerStatusFunc,
	}
}

func setupValidNotaryServer(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	database := postgresql.New(ctx, ns)

	name := newName("notary-server")
	notaryServer := &goharborv1.NotaryServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: test.AddVersionAnnotations(nil),
		},
		Spec: goharborv1.NotaryServerSpec{
			Storage: goharborv1.NotaryStorageSpec{
				Postgres: database,
			},
		},
	}

	Expect(k8sClient.Create(ctx, notaryServer)).To(Succeed())

	return notaryServer, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateNotaryServer(ctx context.Context, object Resource) {
	notaryServer, ok := object.(*goharborv1.NotaryServer)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if notaryServer.Spec.Replicas != nil {
		replicas = *notaryServer.Spec.Replicas + 1
	}

	notaryServer.Spec.Replicas = &replicas
}

func getNotaryServerStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var notaryServer goharborv1.NotaryServer

		err := k8sClient.Get(ctx, key, &notaryServer)

		Expect(err).ToNot(HaveOccurred())

		return notaryServer.Status
	}
}
