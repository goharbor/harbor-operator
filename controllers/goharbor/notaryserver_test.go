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

func newNotaryServerController() controllerTest {
	return controllerTest{
		Setup:         setupValidNotaryServer,
		Update:        updateNotaryServer,
		GetStatusFunc: getNotaryServerStatusFunc,
	}
}

func setupValidNotaryServer(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	database := setupPostgresql(ctx, ns)

	name := newName("notary-server")
	notaryServer := &goharborv1alpha2.NotaryServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.NotaryServerSpec{
			Storage: goharborv1alpha2.NotaryStorageSpec{
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
	notaryServer, ok := object.(*goharborv1alpha2.NotaryServer)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if notaryServer.Spec.Replicas != nil {
		replicas = *notaryServer.Spec.Replicas + 1
	}

	notaryServer.Spec.Replicas = &replicas
}

func getNotaryServerStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var notaryServer goharborv1alpha2.NotaryServer

		err := k8sClient.Get(ctx, key, &notaryServer)

		Expect(err).ToNot(HaveOccurred())

		return notaryServer.Status
	}
}
