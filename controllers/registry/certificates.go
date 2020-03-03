package registry

import (
	"context"
	"fmt"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/ovh/configstore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	defaultKeyAlgorithm = certv1.RSAKeyAlgorithm
	defaultKeySize      = 4096
)

type certificateEncryption struct {
	KeySize      int
	KeyAlgorithm certv1.KeyAlgorithm
}

func (r *Reconciler) GetCertificate(ctx context.Context, registry *goharborv1alpha2.Registry) (*certv1.Certificate, error) {
	encryption := &certificateEncryption{
		KeySize:      defaultKeySize,
		KeyAlgorithm: defaultKeyAlgorithm,
	}

	item, err := configstore.Filter().Slice("certificate-encryption").Unmarshal(func() interface{} { return &certificateEncryption{} }).GetFirstItem()
	if err == nil {
		l := logger.Get(ctx)

		// todo
		encryptionConfig, err := item.Unmarshaled()
		if err != nil {
			l.Error(err, "Invalid encryption certificate config: use default value")
		} else {
			encryption = encryptionConfig.(*certificateEncryption)
		}
	}

	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-registry", registry.GetName()),
			Namespace: registry.GetNamespace(),
		},
		Spec: certv1.CertificateSpec{
			CommonName:   registry.Spec.PublicURL,
			Organization: []string{"Harbor Operator"},
			SecretName:   fmt.Sprintf("%s-registry-certificate", registry.GetName()),
			KeySize:      encryption.KeySize,
			KeyAlgorithm: encryption.KeyAlgorithm,
			// https://github.com/goharbor/harbor/blob/ba4764c61d7da76f584f808f7d16b017db576fb4/src/jobservice/generateCerts.sh#L24-L26
			KeyEncoding: certv1.PKCS1,
			DNSNames:    []string{registry.Spec.PublicURL},
			IssuerRef:   registry.Spec.CertificateIssuerRef,
		},
	}, nil
}
