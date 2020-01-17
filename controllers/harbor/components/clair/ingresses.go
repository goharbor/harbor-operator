package clair

import (
	"context"

	netv1 "k8s.io/api/networking/v1beta1"
)

func (*Clair) GetIngresses(ctx context.Context) []*netv1.Ingress {
	return []*netv1.Ingress{}
}
