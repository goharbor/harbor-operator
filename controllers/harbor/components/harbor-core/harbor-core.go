package core

import (
	"context"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
)

type HarborCore struct {
	harbor *containerregistryv1alpha1.Harbor
	Option Option
}

type Option interface {
	GetPriority() *int32
}

func New(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, opt Option) (*HarborCore, error) {
	return &HarborCore{
		harbor: harbor,
		Option: opt,
	}, nil
}
