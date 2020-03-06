package portal

import (
	"context"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
)

type Portal struct {
	harbor *goharborv1alpha1.Harbor
	Option Option
}

type Option interface {
	GetPriority() *int32
}

func New(ctx context.Context, harbor *goharborv1alpha1.Harbor, opt Option) (*Portal, error) {
	return &Portal{
		harbor: harbor,
		Option: opt,
	}, nil
}
