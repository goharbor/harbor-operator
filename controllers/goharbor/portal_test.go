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

func newPortalController() controllerTest {
	return controllerTest{
		Setup:         setupValidPortal,
		Update:        updatePortal,
		GetStatusFunc: getPortalStatusFunc,
	}
}

func setupValidPortal(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	name := newName("portal")
	portal := &goharborv1alpha2.Portal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}

	Expect(k8sClient.Create(ctx, portal)).To(Succeed())

	return portal, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updatePortal(ctx context.Context, object Resource) {
	portal, ok := object.(*goharborv1alpha2.Portal)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if portal.Spec.Replicas != nil {
		replicas = *portal.Spec.Replicas + 1
	}

	portal.Spec.Replicas = &replicas
}

func getPortalStatusFunc(ctx context.Context, key client.ObjectKey) func() goharborv1alpha2.ComponentStatus {
	return func() goharborv1alpha2.ComponentStatus {
		var portal goharborv1alpha2.Portal

		err := k8sClient.Get(ctx, key, &portal)

		Expect(err).ToNot(HaveOccurred())

		return portal.Status
	}
}
