package cache

import (
	"strings"
	"time"

	rediscli "github.com/go-redis/redis"
)

const (
	RedisSentinelConnPort  = 26379
	RedisSentinelConnGroup = "mymaster"
)

type RedisConnect struct {
	Schema    string
	Endpoints []string
	Port      string
	Password  string
	GroupName string
}

// NewRedisPool returns redis sentinel client.
func (c *RedisConnect) NewRedisPool() *rediscli.Client {
	return BuildRedisPool(c.Endpoints, c.Port, c.Password, c.GroupName, 0)
}

// NewRedisClient returns redis client.
func (c *RedisConnect) NewRedisClient() *rediscli.Client {
	return BuildRedisClient(c.Endpoints, c.Port, c.Password, 0)
}

// BuildRedisPool returns redis connection pool client.
func BuildRedisPool(redisSentinelIP []string, redisSentinelPort, redisSentinelPassword, redisGroupName string, redisIndex int) *rediscli.Client {
	sentinelsInfo := GenHostInfo(redisSentinelIP, redisSentinelPort)

	const (
		PoolSize     = 100
		DialTimeout  = 10
		ReadTimeout  = 30
		WriteTimeout = 30
		PoolTimeout  = 30
	)

	options := &rediscli.FailoverOptions{
		MasterName:         redisGroupName,
		SentinelAddrs:      sentinelsInfo,
		Password:           redisSentinelPassword,
		DB:                 redisIndex,
		PoolSize:           PoolSize,
		DialTimeout:        DialTimeout * time.Second,
		ReadTimeout:        ReadTimeout * time.Second,
		WriteTimeout:       WriteTimeout * time.Second,
		PoolTimeout:        PoolTimeout * time.Second,
		IdleTimeout:        time.Millisecond,
		IdleCheckFrequency: time.Millisecond,
	}

	client := rediscli.NewFailoverClient(options)

	return client
}

// BuildRedisClient returns redis connection client.
func BuildRedisClient(host []string, port, password string, index int) *rediscli.Client {
	hostInfo := GenHostInfo(host, port)
	options := &rediscli.Options{
		Addr:     strings.Join(hostInfo, ","),
		Password: password,
		DB:       index,
	}
	client := rediscli.NewClient(options)

	return client
}

// GenHostInfo splice host and port.
func GenHostInfo(endpoint []string, port string) []string {
	hostInfo := make([]string, 0, len(endpoint))

	for _, s := range endpoint {
		sp := s + ":" + port
		hostInfo = append(hostInfo, sp)
	}

	return hostInfo
}
