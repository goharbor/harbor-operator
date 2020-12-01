package cache

import (
	"context"
	"fmt"
	"strconv"

	rediscli "github.com/go-redis/redis"

	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
)

var _ lcm.HealthChecker = &RedisHealthChecker{}

// RedisHealthChecker check health for redis service.
type RedisHealthChecker struct{}

// CheckHealth implements lcm.HealthChecker
func (c *RedisHealthChecker) CheckHealth(ctx context.Context, svc *lcm.ServiceConfig, options ...lcm.Option) (*lcm.CheckResponse, error) {
	if svc == nil || svc.Endpoint == nil {
		return nil, fmt.Errorf("serviceConfig or endpoint can not be nil")
	}
	// apply options
	checkOpts := &lcm.CheckOptions{}
	for _, o := range options {
		o(checkOpts)
	}

	var client *rediscli.Client
	// build redis client
	redisConn := &RedisConnect{
		Endpoints: []string{svc.Endpoint.Host},
		Port:      strconv.Itoa(int(svc.Endpoint.Port)),
	}
	// check password
	if svc.Credentials != nil {
		redisConn.Password = svc.Credentials.AccessSecret
	}
	// check mode
	if checkOpts.Sentinel {
		redisConn.GroupName = RedisSentinelConnGroup
		client = redisConn.NewRedisPool()
	} else {
		client = redisConn.NewRedisClient()
	}

	resp := &lcm.CheckResponse{}
	defer client.Close()
	err := client.Ping().Err()
	if err != nil {
		resp.Status = lcm.UnHealthy
		resp.Message = err.Error()
		return resp, err
	}

	resp.Status = lcm.Healthy
	return resp, nil
}
