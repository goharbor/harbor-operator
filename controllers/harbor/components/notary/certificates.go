package notary

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/ovh/configstore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	defaultKeyAlgorithm   = certv1.RSAKeyAlgorithm
	defaultKeySize        = 4096
	notaryCertificateName = "notary-certificate"
)

var (
	notaryCertificateKeyAlgorithm = defaultKeyAlgorithm
	notaryCertificateKeySize      = defaultKeySize
)

type certificateEncryption struct {
	KeySize      int
	KeyAlgorithm certv1.KeyAlgorithm
}

func (n *Notary) GetCertificates(ctx context.Context) []*certv1.Certificate {
	operatorName := application.GetName(ctx)
	harborName := n.harbor.Name

	url := n.harbor.Spec.Components.Notary.PublicURL

	encryption := &certificateEncryption{
		KeySize:      notaryCertificateKeySize,
		KeyAlgorithm: notaryCertificateKeyAlgorithm,
	}

	item, err := configstore.Filter().Slice("certificate-encryption").Unmarshal(func() interface{} { return &certificateEncryption{} }).GetFirstItem()
	if err == nil {
		l := logger.Get(ctx)

		encryptionConfig, err := item.Unmarshaled()
		if err != nil {
			l.Error(err, "Invalid encryption certificate config: use default value")
		} else {
			encryption = encryptionConfig.(*certificateEncryption)
		}
	}

	return []*certv1.Certificate{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      n.harbor.NormalizeComponentName(notaryCertificateName),
				Namespace: n.harbor.Namespace,
				Labels: map[string]string{
					"app":      notaryCertificateName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: certv1.CertificateSpec{
				CommonName:   url,
				Organization: []string{"Harbor Operator"},
				SecretName:   n.harbor.NormalizeComponentName(notaryCertificateName),
				KeySize:      encryption.KeySize,
				KeyAlgorithm: encryption.KeyAlgorithm,
				KeyEncoding:  certv1.PKCS1,
				DNSNames:     []string{n.harbor.NormalizeComponentName(NotarySignerName)},
				IssuerRef:    n.harbor.Spec.CertificateIssuerRef,
			},
		},
	}
}
