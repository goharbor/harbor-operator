package registry

import (
	"context"

	netv1 "k8s.io/api/networking/v1beta1"
)

func (r *Registry) GetIngresses(ctx context.Context) []*netv1.Ingress {
	return []*netv1.Ingress{}
}
