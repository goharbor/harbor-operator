package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v4"

	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
)

var _ lcm.HealthChecker = &PostgreSQLHealthChecker{}

// PostgreSQLHealthChecker check health for postgresql service.
type PostgreSQLHealthChecker struct{}

// CheckHealth implements lcm.HealthChecker
func (p *PostgreSQLHealthChecker) CheckHealth(ctx context.Context, svc *lcm.ServiceConfig, options ...lcm.Option) (*lcm.CheckResponse, error) {
	if svc == nil || svc.Endpoint == nil {
		return nil, fmt.Errorf("serviceConfig or endpoint can not be nil")
	}
	// apply options
	checkOpts := &lcm.CheckOptions{}
	for _, o := range options {
		o(checkOpts)
	}

	var (
		client *pgx.Conn
		err    error
	)
	conn := Connect{
		Host:     svc.Endpoint.Host,
		Port:     strconv.Itoa(int(svc.Endpoint.Port)),
		Database: InClusterDatabaseName,
	}

	if svc.Credentials != nil {
		conn.Password = svc.Credentials.AccessSecret
		conn.Username = svc.Credentials.AccessKey
	}

	url := conn.GenDatabaseUrl()
	resp := &lcm.CheckResponse{}

	client, err = pgx.Connect(ctx, url)
	if err != nil {
		return nil, err
	}

	defer client.Close(ctx)

	if err := client.Ping(ctx); err != nil {
		resp.Status = lcm.UnHealthy
		resp.Message = err.Error()
		return resp, err
	}

	resp.Status = lcm.Healthy
	return resp, nil

}
