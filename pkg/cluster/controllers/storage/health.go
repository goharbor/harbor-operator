package storage

import (
	"context"
	"fmt"
	"net/http"
	"time"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// HealthChecker for doing health checking for storage.
type HealthChecker struct{}

var _ lcm.HealthChecker = &HealthChecker{}

func (c *HealthChecker) CheckHealth(ctx context.Context, svc *lcm.ServiceConfig, options ...lcm.Option) (*lcm.CheckResponse, error) {

	checkOpts := &lcm.CheckOptions{}

	for _, o := range options {
		o(checkOpts)
	}

	switch checkOpts.StorageDriver {
	case goharborv1.S3DriverName:
		return S3StorageHealthCheck(ctx, svc, checkOpts)
	case goharborv1.SwiftDriverName:
		return SwiftStorageHealthCheck(ctx, svc, checkOpts)
	case goharborv1.FileSystemDriverName:
		return &lcm.CheckResponse{
			Message: "skiped check",
			Status:  lcm.Healthy,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", checkOpts.StorageDriver)
	}
}

func S3StorageHealthCheck(ctx context.Context, svc *lcm.ServiceConfig, options *lcm.CheckOptions) (*lcm.CheckResponse, error) {
	checkRes := &lcm.CheckResponse{}
	edp := svc.Endpoint.Host + string(svc.Endpoint.Port)
	bucket := &s3.HeadBucketInput{
		Bucket: aws.String(options.BucketName),
	}

	// Configure to use s3 Server, also can used for MinIO server
	s3Config := &aws.Config{
		Region:           aws.String(options.S3Region),
		Endpoint:         aws.String(edp),
		Credentials:      credentials.NewStaticCredentials(svc.Credentials.AccessKey, svc.Credentials.AccessSecret, ""),
		DisableSSL:       aws.Bool(options.SSLMode),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       &http.Client{Timeout: 10 * time.Second,},
		MaxRetries:       aws.Int(5),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}
	s3Client := s3.New(newSession)

	// check if the Bucket exists, If the condition is not met within the max attempt window, an error will
	// be returned.
	err = s3Client.WaitUntilBucketExists(bucket)
	if err != nil {
		return nil, err
	}

	checkRes.Status = lcm.Healthy
	return checkRes, nil
}

// TODO soulseen
func SwiftStorageHealthCheck(ctx context.Context, svc *lcm.ServiceConfig, options *lcm.CheckOptions) (*lcm.CheckResponse, error) {
	resp := &lcm.CheckResponse{
		Status: lcm.Healthy,
	}

	return resp, nil
}
