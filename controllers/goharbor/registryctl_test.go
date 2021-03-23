package goharbor_test

import (
	"context"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"
)

func newRegistryCtlController() controllerTest {
	return controllerTest{
		Setup:         setupValidRegistryCtl,
		Update:        updateRegistryCtl,
		GetStatusFunc: getRegistryCtlStatusFunc,
	}
}

func setupValidRegistryCtl(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	registryObject, key := setupValidRegistry(ctx, ns)

	// TODO remove this once the controller owning registryCtl watch Registries' events.
	Eventually(getRegistryStatusFunc(ctx, key), time.Minute, 2*time.Second).
		Should(MatchFields(IgnoreExtras, Fields{
			"Conditions": ContainElements(MatchFields(IgnoreExtras, Fields{
				"Type":   BeEquivalentTo(status.ConditionInProgress),
				"Status": BeEquivalentTo(corev1.ConditionFalse),
			}), MatchFields(IgnoreExtras, Fields{
				"Type":   BeEquivalentTo(status.ConditionFailed),
				"Status": BeEquivalentTo(corev1.ConditionFalse),
			})),
		}), "registry should be ready")

	registry := registryObject.(*goharborv1.Registry)

	name := newName("registryctl")
	registryctl := &goharborv1.RegistryController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1.RegistryControllerSpec{
			RegistryRef: registry.GetName(),
		},
	}

	Expect(k8sClient.Create(ctx, registryctl)).To(Succeed())

	return registryctl, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateRegistryCtl(ctx context.Context, object Resource) {
	registryctl, ok := object.(*goharborv1.RegistryController)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if registryctl.Spec.Replicas != nil {
		replicas = *registryctl.Spec.Replicas + 1
	}

	registryctl.Spec.Replicas = &replicas
}

func getRegistryCtlStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var registryctl goharborv1.RegistryController

		err := k8sClient.Get(ctx, key, &registryctl)

		Expect(err).ToNot(HaveOccurred())

		return registryctl.Status
	}
}
