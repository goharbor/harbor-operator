package storage

import (
	"context"
	"fmt"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
)

type HealthChecker struct{}

func (c *HealthChecker) CheckHealth(ctx context.Context, svc *lcm.ServiceConfig, options ...lcm.Option) (*lcm.CheckResponse, error) {

	checkOpts := &lcm.CheckOptions{}

	for _, o := range options {
		o(checkOpts)
	}

	switch checkOpts.StorageDriver {
	case goharborv1.S3DriverName:
		return S3StorageHealthCheck(ctx, svc)
	case goharborv1.SwiftDriverName:
		return SwiftStorageHealthCheck(ctx, svc)
	case goharborv1.FileSystemDriverName:
		return &lcm.CheckResponse{
			Message: "skiped check",
			Status:  lcm.Healthy,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", checkOpts.StorageDriver)
	}
}

func S3StorageHealthCheck(ctx context.Context, svc *lcm.ServiceConfig) (*lcm.CheckResponse, error) {
	resp := &lcm.CheckResponse{}

	return resp, nil
}

func SwiftStorageHealthCheck(ctx context.Context, svc *lcm.ServiceConfig) (*lcm.CheckResponse, error) {
	resp := &lcm.CheckResponse{}

	return resp, nil
}
