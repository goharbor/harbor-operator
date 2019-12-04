package portal

import (
	"context"

	extv1 "k8s.io/api/extensions/v1beta1"
)

func (p *Portal) GetIngresses(ctx context.Context) []*extv1.Ingress {
	return []*extv1.Ingress{}
}
