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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
)

func newNotarySignerController() controllerTest {
	return controllerTest{
		Setup:         setupValidNotarySigner,
		Update:        updateNotarySigner,
		GetStatusFunc: getNotarySignerStatusFunc,
	}
}

func setupNotarySignerResourceDependencies(ctx context.Context, ns string) (string, string) {
	aliasesName := newName("aliases")
	authCertName := newName("authentication-certificate")

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aliasesName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.DefaultAliasSecretKey: "abcde_012345_ABCDE",
		},
		Type: harbormetav1.SecretTypeNotarySignerAliases,
	})).ToNot(HaveOccurred())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      authCertName,
			Namespace: ns,
		},
		Data: generateCertificate(),
		Type: corev1.SecretTypeTLS,
	})).ToNot(HaveOccurred())

	return authCertName, aliasesName
}

func setupValidNotarySigner(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	database := setupPostgresql(ctx, ns)
	authCertName, aliasesName := setupNotarySignerResourceDependencies(ctx, ns)

	name := newName("notary-signer")
	notarySigner := &goharborv1alpha2.NotarySigner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.NotarySignerSpec{
			Storage: goharborv1alpha2.NotarySignerStorageSpec{
				NotaryStorageSpec: goharborv1alpha2.NotaryStorageSpec{
					Postgres: database,
				},
				AliasesRef: aliasesName,
			},
			Authentication: goharborv1alpha2.NotarySignerAuthenticationSpec{
				CertificateRef: authCertName,
			},
		},
	}

	Expect(k8sClient.Create(ctx, notarySigner)).To(Succeed())

	return notarySigner, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateNotarySigner(ctx context.Context, object Resource) {
	notarySigner, ok := object.(*goharborv1alpha2.NotarySigner)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if notarySigner.Spec.Replicas != nil {
		replicas = *notarySigner.Spec.Replicas + 1
	}

	notarySigner.Spec.Replicas = &replicas
}

func getNotarySignerStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var notarySigner goharborv1alpha2.NotarySigner

		err := k8sClient.Get(ctx, key, &notarySigner)

		Expect(err).ToNot(HaveOccurred())

		return notarySigner.Status
	}
}
