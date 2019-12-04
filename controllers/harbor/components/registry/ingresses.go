package registry

import (
	"context"

	extv1 "k8s.io/api/extensions/v1beta1"
)

func (r *Registry) GetIngresses(ctx context.Context) []*extv1.Ingress {
	return []*extv1.Ingress{}
}
