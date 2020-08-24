package registry

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/ovh/configstore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	defaultKeyAlgorithm = certv1.RSAKeyAlgorithm
	defaultKeySize      = 4096
	registryCertName    = "registry-certificate"
)

type certificateEncryption struct {
	KeySize      int
	KeyAlgorithm certv1.KeyAlgorithm
}

func (r *Registry) GetCertificates(ctx context.Context) []*certv1.Certificate {
	operatorName := application.GetName(ctx)
	harborName := r.harbor.Name

	url := r.harbor.Spec.PublicURL

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

	return []*certv1.Certificate{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      r.harbor.NormalizeComponentName(registryCertName),
				Namespace: r.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.RegistryName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: certv1.CertificateSpec{
				CommonName:   url,
				Organization: []string{"Harbor Operator"},
				SecretName:   r.harbor.NormalizeComponentName(registryCertName),
				KeySize:      encryption.KeySize,
				KeyAlgorithm: encryption.KeyAlgorithm,
				// https://github.com/goharbor/harbor/blob/ba4764c61d7da76f584f808f7d16b017db576fb4/src/jobservice/generateCerts.sh#L24-L26
				KeyEncoding: certv1.PKCS1,
				DNSNames:    []string{url},
				IssuerRef:   r.harbor.Spec.CertificateIssuerRef,
			},
		},
	}
}
