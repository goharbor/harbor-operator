package core

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
)

func (c *HarborCore) GetCertificates(ctx context.Context) []*certv1.Certificate {
	return []*certv1.Certificate{}
}
