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

package goharbor_test

import (
	"context"

	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

func newRegistryCtlController() controllerTest {
	return controllerTest{
		Setup:         setupValidRegistryCtl,
		Update:        updateRegistryCtl,
		GetStatusFunc: getRegistryCtlStatusFunc,
	}
}

func setupRegistryCtlResourceDependencies(ctx context.Context, ns string) string {
	registryName := newName("registry")

	var replicas int32 = 1

	err := k8sClient.Create(ctx, &goharborv1alpha2.Registry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryName,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.RegistrySpec{
			ComponentSpec: goharborv1alpha2.ComponentSpec{
				Replicas: &replicas,
			},
			RegistryConfig01: goharborv1alpha2.RegistryConfig01{
				Storage: goharborv1alpha2.RegistryStorageSpec{
					Driver: goharborv1alpha2.RegistryStorageDriverSpec{
						InMemory: &goharborv1alpha2.RegistryStorageDriverInmemorySpec{},
					},
				},
			},
		},
	})
	Expect(err).ToNot(HaveOccurred())

	return registryName
}

func setupValidRegistryCtl(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	registryName := setupRegistryCtlResourceDependencies(ctx, ns)

	name := newName("registryctl")
	registryctl := &goharborv1alpha2.RegistryController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.RegistryControllerSpec{
			RegistryRef: registryName,
		},
	}

	Expect(k8sClient.Create(ctx, registryctl)).To(Succeed())

	return registryctl, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateRegistryCtl(ctx context.Context, object Resource) {
	registryctl, ok := object.(*goharborv1alpha2.RegistryController)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if registryctl.Spec.Replicas != nil {
		replicas = *registryctl.Spec.Replicas + 1
	}

	registryctl.Spec.Replicas = &replicas
}

func getRegistryCtlStatusFunc(ctx context.Context, key client.ObjectKey) func() goharborv1alpha2.ComponentStatus {
	return func() goharborv1alpha2.ComponentStatus {
		var registryctl goharborv1alpha2.RegistryController

		err := k8sClient.Get(ctx, key, &registryctl)

		Expect(err).ToNot(HaveOccurred())

		return registryctl.Status
	}
}
