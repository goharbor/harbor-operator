package goharbor_test

import (
	"context"

	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newPortalController() controllerTest {
	return controllerTest{
		Setup:         setupValidPortal,
		Update:        updatePortal,
		GetStatusFunc: getPortalStatusFunc,
	}
}

func setupValidPortal(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	name := newName("portal")
	portal := &goharborv1.Portal{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: test.AddVersionAnnotations(nil),
		},
	}

	Expect(k8sClient.Create(ctx, portal)).To(Succeed())

	return portal, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updatePortal(ctx context.Context, object Resource) {
	portal, ok := object.(*goharborv1.Portal)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if portal.Spec.Replicas != nil {
		replicas = *portal.Spec.Replicas + 1
	}

	portal.Spec.Replicas = &replicas
}

func getPortalStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var portal goharborv1.Portal

		err := k8sClient.Get(ctx, key, &portal)

		Expect(err).ToNot(HaveOccurred())

		return portal.Status
	}
}
