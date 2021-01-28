// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/ncw/swift"
)

// HealthChecker for doing health checking for storage.
type HealthChecker struct{}

var _ lcm.HealthChecker = &HealthChecker{}

// CheckHealth implements lcm.HealthChecker interface for checking health of the storage.
func (c *HealthChecker) CheckHealth(ctx context.Context, svc *lcm.ServiceConfig, options ...lcm.Option) (*lcm.CheckResponse, error) {
	checkOpts := &lcm.CheckOptions{}

	for _, o := range options {
		o(checkOpts)
	}

	var res *lcm.CheckResponse

	var err error

	switch checkOpts.StorageDriver {
	case goharborv1.S3DriverName:
		res, err = S3StorageHealthCheck(ctx, svc, checkOpts)
	case goharborv1.SwiftDriverName:
		res, err = SwiftStorageHealthCheck(ctx, svc, checkOpts)
	case goharborv1.FileSystemDriverName:
		res, err = &lcm.CheckResponse{
			Message: "skipped check",
			Status:  lcm.Healthy,
		}, nil
	default:
		res, err = nil, fmt.Errorf("unsupported storage driver: %s", checkOpts.StorageDriver)
	}

	if err != nil {

		return res, fmt.Errorf("storage: %w, %v", lcm.ErrUnHealth, err)
	}
	return res, nil
}

func S3StorageHealthCheck(ctx context.Context, svc *lcm.ServiceConfig, options *lcm.CheckOptions) (*lcm.CheckResponse, error) {
	checkRes := &lcm.CheckResponse{}
	bucket := &s3.HeadBucketInput{
		Bucket: aws.String(options.S3Options.BucketName),
	}

	// Configure to use s3 Server, also can used for MinIO server.
	// For s3 the Host contains the Port already.
	s3Config := &aws.Config{
		Region:           aws.String(options.S3Options.S3Region),
		Endpoint:         aws.String(svc.Endpoint.Host),
		Credentials:      credentials.NewStaticCredentials(svc.Credentials.AccessKey, svc.Credentials.AccessSecret, ""),
		DisableSSL:       aws.Bool(options.SSLMode),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       &http.Client{Timeout: 10 * time.Second},
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

func SwiftStorageHealthCheck(ctx context.Context, svc *lcm.ServiceConfig, options *lcm.CheckOptions) (*lcm.CheckResponse, error) {
	checkRes := &lcm.CheckResponse{}

	// Create a connection
	c := swift.Connection{
		UserName: svc.Credentials.AccessKey,
		ApiKey:   svc.Credentials.AccessSecret,
		AuthUrl:  options.AuthURL,
		Domain:   options.Domain, // Name of the domain (v3 auth only)
		DomainId: options.DomainID,
		Tenant:   options.Tenant, // Name of the tenant (v2 auth only)
		TenantId: options.TenantID,
		Region:   options.Region,
	}

	// Authenticate
	err := c.Authenticate()
	if err != nil {
		return nil, err
	}

	containerInfo, _, err := c.Container(options.Container)
	if err != nil {
		return nil, err
	}

	checkRes.Status = lcm.Healthy
	checkRes.Message = fmt.Sprintf("Check container %s successful", containerInfo.Name)

	return checkRes, nil
}
