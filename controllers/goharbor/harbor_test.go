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
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

var _ = Context("Harbor reconciler", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = logger.Context(log)
	})

	Describe("Creating resources with invalid public url", func() {
		It("Should raise an error", func() {
			harbor := &goharborv1alpha2.Harbor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "harbor-invalid-url",
					Namespace: ns.Name,
				},
				Spec: goharborv1alpha2.HarborSpec{
					HarborHelm1_4_0Spec: goharborv1alpha2.HarborHelm1_4_0Spec{
						ExternalURL: "123::bad::dns",
					},
				},
			}
			harbor.Default()

			Expect(k8sClient.Create(ctx, harbor)).
				Should(WithTransform(apierrs.IsInvalid, BeTrue()))
		})
	})
})

func newHarborController() controllerTest {
	return controllerTest{
		Setup:         setupValidHarbor,
		Update:        updateHarbor,
		GetStatusFunc: getHarborStatusFunc,
	}
}

func setupHarborResourceDependencies(ctx context.Context, ns string) (string, string) {
	adminSecretName := newName("admin-secret")

	err := k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      adminSecretName,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.SharedSecretKey: "th3Adm!nPa$$w0rd",
		},
		Type: goharborv1alpha2.SecretTypeSingle,
	})
	Expect(err).ToNot(HaveOccurred())

	tokenIssuerName := newName("token-issuer")

	err = k8sClient.Create(ctx, &certv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenIssuerName,
			Namespace: ns,
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				SelfSigned: &certv1.SelfSignedIssuer{},
			},
		},
	})
	Expect(err).ToNot(HaveOccurred())

	return adminSecretName, tokenIssuerName
}

func setupValidHarbor(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	adminSecretName, tokenIssuerName := setupHarborResourceDependencies(ctx, ns)

	corePG := setupPostgresql(ctx, ns)

	name := newName("harbor")
	publicURL := url.URL{
		Scheme: "http",
		Host:   "the.dns",
	}

	db, _, err := goharborv1alpha2.FromOpacifiedDSN(corePG)
	Expect(err).ToNot(HaveOccurred())

	harbor := &goharborv1alpha2.Harbor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.HarborSpec{
			HarborHelm1_4_0Spec: goharborv1alpha2.HarborHelm1_4_0Spec{
				ExternalURL:            publicURL.String(),
				HarborAdminPasswordRef: adminSecretName,
				EncryptionKeyRef:       "encryption-key",
				Persistence: goharborv1alpha2.HarborPersistenceSpec{
					ImageChartStorage: goharborv1alpha2.HarborPersistenceImageChartStorageSpec{
						FileSystem: &goharborv1alpha2.HarborPersistenceImageChartStorageFileSystemSpec{
							RootDirectory: "/harbor",
						},
					},
					PersistentVolumeClaim: goharborv1alpha2.HarborPersistencePersistentVolumeClaimComponentsSpec{
						Registry: goharborv1alpha2.HarborPersistencePersistentVolumeClaim5GSpec{
							HarborPersistencePersistentVolumeClaimComponentSpec: goharborv1alpha2.HarborPersistencePersistentVolumeClaimComponentSpec{
								ExistingClaim: "test-registry",
							},
						},
					},
				},
				HarborComponentsSpec: goharborv1alpha2.HarborComponentsSpec{
					Core: goharborv1alpha2.CoreComponentSpec{
						TokenIssuer: cmmeta.ObjectReference{
							Name: tokenIssuerName,
						},
					},
					Database: *db,
				},
			},
		},
	}
	harbor.Default()

	Expect(k8sClient.Create(ctx, harbor)).To(Succeed())

	return harbor, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateHarbor(ctx context.Context, object Resource) {
	harbor, ok := object.(*goharborv1alpha2.Harbor)
	Expect(ok).To(BeTrue())

	u, err := url.Parse(harbor.Spec.ExternalURL)
	Expect(err).ToNot(HaveOccurred())

	u.Host = "new." + u.Host
	harbor.Spec.ExternalURL = u.String()
}

func getHarborStatusFunc(ctx context.Context, key client.ObjectKey) func() goharborv1alpha2.ComponentStatus {
	return func() goharborv1alpha2.ComponentStatus {
		var harbor goharborv1alpha2.Harbor

		err := k8sClient.Get(ctx, key, &harbor)

		Expect(err).ToNot(HaveOccurred())

		return harbor.Status
	}
}
