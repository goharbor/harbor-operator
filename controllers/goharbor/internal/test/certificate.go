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

package test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func GenerateCertificate() map[string][]byte {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	//	x509.KeyUsageDigitalSignature
	var ca *x509.Certificate

	reader, writer := io.Pipe()

	go func() {
		defer ginkgo.GinkgoRecover()

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
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		ca, err = x509.ParseCertificate(certBytes)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		gomega.Expect(pem.Encode(writer, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})).
			To(gomega.Succeed())

		gomega.Expect(writer.Close()).To(gomega.Succeed())
	}()

	caPublicBytes, err := ioutil.ReadAll(reader)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	gomega.Expect(reader.Close()).To(gomega.Succeed())

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	reader, writer = io.Pipe()

	go func() {
		defer ginkgo.GinkgoRecover()

		privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		gomega.Expect(pem.Encode(writer, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})).
			To(gomega.Succeed())

		gomega.Expect(writer.Close()).To(gomega.Succeed())
	}()

	privBytes, err := ioutil.ReadAll(reader)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	gomega.Expect(reader.Close()).To(gomega.Succeed())

	reader, writer = io.Pipe()

	go func() {
		defer ginkgo.GinkgoRecover()

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
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		gomega.Expect(pem.Encode(writer, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})).
			To(gomega.Succeed())

		gomega.Expect(writer.Close()).To(gomega.Succeed())
	}()

	publicBytes, err := ioutil.ReadAll(reader)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	gomega.Expect(reader.Close()).To(gomega.Succeed())

	return map[string][]byte{
		corev1.TLSPrivateKeyKey:        privBytes,
		corev1.TLSCertKey:              publicBytes,
		corev1.ServiceAccountRootCAKey: caPublicBytes,
	}
}
