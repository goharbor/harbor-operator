package goharbor_test

import (
	"context"

	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/certificate"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/postgresql"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		Data: certificate.NewCA().NewCert().ToMap(),
		Type: corev1.SecretTypeTLS,
	})).ToNot(HaveOccurred())

	return authCertName, aliasesName
}

func setupValidNotarySigner(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	database := postgresql.New(ctx, ns)
	authCertName, aliasesName := setupNotarySignerResourceDependencies(ctx, ns)

	name := newName("notary-signer")
	notarySigner := &goharborv1.NotarySigner{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: test.AddVersionAnnotations(nil),
		},
		Spec: goharborv1.NotarySignerSpec{
			Storage: goharborv1.NotarySignerStorageSpec{
				NotaryStorageSpec: goharborv1.NotaryStorageSpec{
					Postgres: database,
				},
				AliasesRef: aliasesName,
			},
			Authentication: goharborv1.NotarySignerAuthenticationSpec{
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
	notarySigner, ok := object.(*goharborv1.NotarySigner)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if notarySigner.Spec.Replicas != nil {
		replicas = *notarySigner.Spec.Replicas + 1
	}

	notarySigner.Spec.Replicas = &replicas
}

func getNotarySignerStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var notarySigner goharborv1.NotarySigner

		err := k8sClient.Get(ctx, key, &notarySigner)

		Expect(err).ToNot(HaveOccurred())

		return notarySigner.Status
	}
}
