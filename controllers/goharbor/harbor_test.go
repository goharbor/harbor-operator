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
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
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

			err := k8sClient.Create(ctx, harbor)
			Expect(err).To(HaveOccurred())
			Expect(err).To(WithTransform(apierrs.IsInvalid, BeTrue()))
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

func setupHarborResourceDependencies(ctx context.Context, ns string) (string, string, string) {
	adminSecretName := newName("admin-secret")

	err := k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      adminSecretName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: "th3Adm!nPa$$w0rd",
		},
		Type: harbormetav1.SecretTypeSingle,
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

	pvcName := newName("pvc")

	err = k8sClient.Create(ctx, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: ns,
		},
	})
	Expect(err).ToNot(HaveOccurred())

	return pvcName, adminSecretName, tokenIssuerName
}

func setupValidHarbor(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	pvcName, adminSecretName, tokenIssuerName := setupHarborResourceDependencies(ctx, ns)

	database := setupPostgresql(ctx, ns)

	name := newName("harbor")
	publicURL := url.URL{
		Scheme: "http",
		Host:   "the.dns",
	}

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
				ImageChartStorage: goharborv1alpha2.HarborStorageImageChartStorageSpec{
					FileSystem: &goharborv1alpha2.HarborStorageImageChartStorageFileSystemSpec{
						RegistryPersistentVolume: goharborv1alpha2.HarborStorageRegistryPersistentVolumeSpec{
							HarborStoragePersistentVolumeSpec: goharborv1alpha2.HarborStoragePersistentVolumeSpec{
								PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
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
					Database: goharborv1alpha2.HarborDatabaseSpec{
						PostgresCredentials: database.PostgresCredentials,
						Hosts:               database.Hosts,
						SSLMode:             harbormetav1.PostgresSSLMode(database.Parameters[harbormetav1.PostgresSSLModeKey]),
					},
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

func getHarborStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var harbor goharborv1alpha2.Harbor

		err := k8sClient.Get(ctx, key, &harbor)

		Expect(err).ToNot(HaveOccurred())

		return harbor.Status
	}
}
