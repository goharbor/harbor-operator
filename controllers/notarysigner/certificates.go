package notarysigner

import (
	"context"
	"fmt"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

const (
	notaryCertificateName = "notary-certificate"
)

func (r *Reconciler) GetNotaryCertificate(ctx context.Context, notary *goharborv1alpha2.NotarySigner) (*certv1.Certificate, error) {
	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-notarysigner", notary.GetName()),
			Namespace: notary.GetNamespace(),
		},
		Spec: certv1.CertificateSpec{
			CommonName:   notary.Spec.CommonName,
			Organization: notary.Spec.Organization,
			SecretName:   fmt.Sprintf("%s-notarysigner", notary.GetName()),
			KeySize:      notary.Spec.KeySize,
			KeyAlgorithm: certv1.RSAKeyAlgorithm,
			KeyEncoding:  certv1.PKCS1,
			DNSNames: []string{
				fmt.Sprintf("%s-notarysigner", notary.GetName()),
				notary.Spec.PublicURL,
			},
			IssuerRef: notary.Spec.CertificateIssuerRef,
		},
	}, nil
}
