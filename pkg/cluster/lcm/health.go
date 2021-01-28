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

package lcm

import (
	"context"
	"errors"
)

const (
	Healthy   HealthStatus = "healthy"
	UnHealthy HealthStatus = "unhealthy"
	Unknown   HealthStatus = "unknown"
)

var (
	UnHealthError = errors.New("service health checking error")
)

type HealthStatus string

// HealthChecker defines health check methods to check the health status of the related services.
type HealthChecker interface {
	// CheckHealth checks the health of the specified service, including:
	// database postgresql /
	// cache redis /
	// storage minIO /
	CheckHealth(ctx context.Context, svc *ServiceConfig, options ...Option) (*CheckResponse, error)
}

// ServiceConfig contains the relevant service configurations that can be used to do health check.
type ServiceConfig struct {
	// Endpoint of the service
	// Required
	Endpoint *Endpoint
	// Credentials used to connect service
	Credentials *Credentials
}

// Endpoint of the service.
type Endpoint struct {
	Host string
	Port uint
}

// Credentials for connecting to the services.
type Credentials struct {
	// Access key or username
	// Optional
	AccessKey string
	// Access secret or password
	AccessSecret string
}

// CheckResponse represents the response returned by the health check method.
type CheckResponse struct {
	Status HealthStatus

	// Extra message if needed
	// Optional
	Message string
}

// CheckOptions keep options for doing health checking.
type CheckOptions struct {
	// Enable SSL mode
	// Applicable for Postgresql
	SSLMode bool

	// Whether connecting to Redis with sentinel mode
	// Applicable for Redis
	Sentinel bool

	// Name of the storage driver
	// Applicable for minIO
	StorageDriver string
	S3Options
	SwiftOptions
}

// For s3 options.
type S3Options struct {
	S3Region   string
	BucketName string
}

// For Swift storage options.
type SwiftOptions struct {
	AuthURL   string
	Tenant    string
	TenantID  string
	Domain    string
	DomainID  string
	Region    string
	Container string
}

// Option with function way.
type Option func(options *CheckOptions)

// WithSSL sets ssl mode option.
func WithSSL(sslMode bool) Option {
	return func(options *CheckOptions) {
		options.SSLMode = sslMode
	}
}

// WithSentinel sets sentinel mode.
func WithSentinel(sentinel bool) Option {
	return func(options *CheckOptions) {
		options.Sentinel = sentinel
	}
}

// WithStorage sets the storage driver.
func WithStorage(driver string) Option {
	return func(options *CheckOptions) {
		options.StorageDriver = driver
	}
}
