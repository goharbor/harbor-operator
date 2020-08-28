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

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newRegistryController() controllerTest {
	return controllerTest{
		Setup:         setupValidRegistry,
		Update:        updateRegistry,
		GetStatusFunc: getRegistryStatusFunc,
	}
}

func setupValidRegistry(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	var replicas int32 = 1

	name := newName("registry")
	registry := &goharborv1alpha2.Registry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.RegistrySpec{
			ComponentSpec: harbormetav1.ComponentSpec{
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
	}

	Expect(k8sClient.Create(ctx, registry)).To(Succeed())

	return registry, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateRegistry(ctx context.Context, object Resource) {
	registry, ok := object.(*goharborv1alpha2.Registry)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if registry.Spec.Replicas != nil {
		replicas = *registry.Spec.Replicas + 1
	}

	registry.Spec.Replicas = &replicas
}

func getRegistryStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var registry goharborv1alpha2.Registry

		err := k8sClient.Get(ctx, key, &registry)

		Expect(err).ToNot(HaveOccurred())

		return registry.Status
	}
}
