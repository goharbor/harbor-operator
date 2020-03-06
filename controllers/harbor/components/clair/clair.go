package clair

import (
	"context"

	containerregistryv1alpha1 "github.com/goharbor/harbor-core-operator/api/v1alpha1"
)

type Clair struct {
	harbor *containerregistryv1alpha1.Harbor
	Option Option
}

type Option interface {
	GetPriority() *int32
}

func New(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, opt Option) (*Clair, error) {
	return &Clair{
		harbor: harbor,
		Option: opt,
	}, nil
}
