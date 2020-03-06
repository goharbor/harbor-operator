package jobservice

import (
	"context"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
)

type JobService struct {
	harbor *containerregistryv1alpha1.Harbor
	Option Option
}

type Option interface {
	GetPriority() *int32
}

func New(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, opt Option) (*JobService, error) {
	return &JobService{
		harbor: harbor,
		Option: opt,
	}, nil
}
