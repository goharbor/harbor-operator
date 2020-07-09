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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"io/ioutil"
	"math/big"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

func newNotarySignerController() controllerTest {
	return controllerTest{
		Setup:         setupValidNotarySigner,
		Update:        updateNotarySigner,
		GetStatusFunc: getNotarySignerStatusFunc,
	}
}

func generateCertificate() map[string][]byte {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).ToNot(HaveOccurred())

	//	x509.KeyUsageDigitalSignature
	var ca *x509.Certificate

	reader, writer := io.Pipe()

	go func() {
		defer GinkgoRecover()

		now := time.Now()
		template := x509.Certificate{
			SerialNumber: new(big.Int).Lsh(big.NewInt(1), 128),
			Subject: pkix.Name{
				Organization: []string{"goharbor.io"},
			},
			NotBefore:             now,
			NotAfter:              now.Add(time.Minute * 30),
			KeyUsage:              x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
			IsCA:                  true,
		}

		certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &caKey.PublicKey, caKey)
		Expect(err).ToNot(HaveOccurred())

		ca, err = x509.ParseCertificate(certBytes)
		Expect(err).ToNot(HaveOccurred())

		Expect(pem.Encode(writer, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})).
			To(Succeed())

		Expect(writer.Close()).To(Succeed())
	}()

	caPublicBytes, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())

	Expect(reader.Close()).To(Succeed())

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).ToNot(HaveOccurred())

	reader, writer = io.Pipe()

	go func() {
		defer GinkgoRecover()

		privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
		Expect(err).ToNot(HaveOccurred())

		Expect(pem.Encode(writer, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})).
			To(Succeed())

		Expect(writer.Close()).To(Succeed())
	}()

	privBytes, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())

	Expect(reader.Close()).To(Succeed())

	reader, writer = io.Pipe()

	go func() {
		defer GinkgoRecover()

		now := time.Now()
		template := x509.Certificate{
			SerialNumber: new(big.Int).Lsh(big.NewInt(1), 128),
			Subject: pkix.Name{
				Organization: []string{"goharbor.io"},
			},
			NotBefore:             now,
			NotAfter:              now.Add(time.Minute * 30),
			KeyUsage:              x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
			IsCA:                  true,
		}

		certBytes, err := x509.CreateCertificate(rand.Reader, &template, ca, &priv.PublicKey, priv)
		Expect(err).ToNot(HaveOccurred())

		Expect(pem.Encode(writer, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})).
			To(Succeed())

		Expect(writer.Close()).To(Succeed())
	}()

	publicBytes, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())

	Expect(reader.Close()).To(Succeed())

	return map[string][]byte{
		corev1.TLSPrivateKeyKey:        privBytes,
		corev1.TLSCertKey:              publicBytes,
		corev1.ServiceAccountRootCAKey: caPublicBytes,
	}
}

func setupNotarySignerResourceDependencies(ctx context.Context, ns string) (string, string, string, string) {
	pgPasswordName := newName("pg-password")
	aliasesName := newName("aliases")
	signingCertName := newName("signing-certificate")
	httpsCertName := newName("https-certificate")

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgPasswordName,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.PostgresqlPasswordKey: "th3Adm!nPa$$w0rd",
		},
		Type: goharborv1alpha2.SecretTypePostgresql,
	})).ToNot(HaveOccurred())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aliasesName,
			Namespace: ns,
		},
		StringData: map[string]string{
			goharborv1alpha2.DefaultAliasSecretKey: "abcde_012345_ABCDE",
		},
		Type: goharborv1alpha2.SecretTypeNotarySignerAliases,
	})).ToNot(HaveOccurred())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      signingCertName,
			Namespace: ns,
		},
		Data: generateCertificate(),
		Type: corev1.SecretTypeTLS,
	})).ToNot(HaveOccurred())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      httpsCertName,
			Namespace: ns,
		},
		Data: generateCertificate(),
		Type: corev1.SecretTypeTLS,
	})).ToNot(HaveOccurred())

	return pgPasswordName, signingCertName, httpsCertName, aliasesName
}

func setupValidNotarySigner(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	pgPasswordName, signingCertName, httpsCertName, aliasesName := setupNotarySignerResourceDependencies(ctx, ns)

	name := newName("notary-signer")
	notarySigner := &goharborv1alpha2.NotarySigner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1alpha2.NotarySignerSpec{
			Storage: goharborv1alpha2.NotarySignerStorageSpec{
				NotaryStorageSpec: goharborv1alpha2.NotaryStorageSpec{
					OpacifiedDSN: goharborv1alpha2.OpacifiedDSN{
						DSN:         "postgres://postgres:password@the.database/notarysigner",
						PasswordRef: pgPasswordName,
					},
					Type: "postgres",
				},
				AliasesRef: aliasesName,
			},
			PublicURL:      "https://notary.url",
			CertificateRef: signingCertName,
			HTTPS: goharborv1alpha2.NotaryHTTPSSpec{
				CertificateRef: httpsCertName,
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

func getNotarySignerStatusFunc(ctx context.Context, key client.ObjectKey) func() goharborv1alpha2.ComponentStatus {
	return func() goharborv1alpha2.ComponentStatus {
		var notarySigner goharborv1alpha2.NotarySigner

		err := k8sClient.Get(ctx, key, &notarySigner)

		Expect(err).ToNot(HaveOccurred())

		return notarySigner.Status
	}
}
