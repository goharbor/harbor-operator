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
	resp := &lcm.CheckResponse{}

	for _, o := range options {
		o(checkOpts)
	}

	switch checkOpts.StorageDriver {
	case goharborv1.S3DriverName:
		err := S3StorageHealthCheck(ctx, svc)
		return resp, err
	case goharborv1.SwiftDriverName:
		err := SwiftStorageHealthCheck(ctx, svc)
		return resp, err
	case goharborv1.FileSystemDriverName:
		resp.Message = "skiped check"
		resp.Status = lcm.Healthy
		return resp, nil
	default:
		return resp, fmt.Errorf("unsupported storage driver: %s", checkOpts.StorageDriver)
	}
}

func S3StorageHealthCheck(ctx context.Context, svc *lcm.ServiceConfig) error {
	return nil
}

func SwiftStorageHealthCheck(ctx context.Context, svc *lcm.ServiceConfig) error {
	return nil
}
