package portal

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
)

func (*Portal) GetCertificates(ctx context.Context) []*certv1.Certificate {
	return []*certv1.Certificate{}
}
