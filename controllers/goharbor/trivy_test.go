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
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
)

func newTrivyController() controllerTest {
	return controllerTest{
		Setup:         setupValidTrivy,
		Update:        updateTrivy,
		GetStatusFunc: getTrivyStatusFunc,
	}
}

func setupValidTrivy(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	var replicas int32 = 1

	name := newName("trivy")
	trivy := &goharborv1alpha2.Trivy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},

		Spec: goharborv1alpha2.TrivySpec{
			ComponentSpec: harbormetav1.ComponentSpec{
				Replicas: &replicas,
			},

			Server: goharborv1alpha2.TrivyServerSpec{
				Address:    ":8080",
				CacheDir:   "/home/scanner/.cache/trivy",
				ReportsDir: "/home/scanner/.cache/reports",
			},

			Cache: goharborv1alpha2.TrivyCacheSpec{
				Redis: goharborv1alpha2.OpacifiedDSN{
					DSN: "redis://10.2.1.14:6379/5",
				},
			},

			Log: goharborv1alpha2.TrivyLogSpec{
				Level: harbormetav1.TrivyDefaultLevel,
			},
		},
	}

	Expect(k8sClient.Create(ctx, trivy)).To(Succeed())

	return trivy, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateTrivy(ctx context.Context, object Resource) {
	trivy, ok := object.(*goharborv1alpha2.Trivy)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if trivy.Spec.Replicas != nil {
		replicas = *trivy.Spec.Replicas + 1
	}

	trivy.Spec.Replicas = &replicas
}

func getTrivyStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var trivy goharborv1alpha2.Trivy

		err := k8sClient.Get(ctx, key, &trivy)

		Expect(err).ToNot(HaveOccurred())

		return trivy.Status
	}
}
