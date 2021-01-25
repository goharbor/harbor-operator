package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
)

func bytesToPEM(certBytes []byte, certPrivKey *rsa.PrivateKey) (*bytes.Buffer, *bytes.Buffer) {
	certPEM := new(bytes.Buffer)
	err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	return certPEM, certPrivKeyPEM
}

func verifyCertificate(caPEM *bytes.Buffer, certPEM *bytes.Buffer, dnsNames ...string) {
	roots := x509.NewCertPool()
	gomega.Expect(roots.AppendCertsFromPEM(caPEM.Bytes())).To(gomega.BeTrue())

	block, _ := pem.Decode(certPEM.Bytes())
	gomega.Expect(block).ToNot(gomega.BeNil())

	cert, err := x509.ParseCertificate(block.Bytes)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	_, err = cert.Verify(x509.VerifyOptions{
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
		Roots:                     roots,
		CurrentTime:               time.Now(),
		MaxConstraintComparisions: 0,
	})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	for _, dnsName := range dnsNames {
		_, err = cert.Verify(x509.VerifyOptions{
			Roots:   roots,
			DNSName: dnsName,
		})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	}
}

func New(dnsNames ...string) map[string][]byte {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"goharbor.io"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Minute * 30),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	caPEM, _ := bytesToPEM(caBytes, caPrivKey)

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"goharbor.io"},
		},
		DNSNames:     dnsNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Minute * 30),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	certPEM, certPrivKeyPEM := bytesToPEM(certBytes, certPrivKey)
	verifyCertificate(caPEM, certPEM, dnsNames...)

	return map[string][]byte{
		corev1.TLSPrivateKeyKey:        certPrivKeyPEM.Bytes(),
		corev1.TLSCertKey:              certPEM.Bytes(),
		corev1.ServiceAccountRootCAKey: caPEM.Bytes(),
	}
}
