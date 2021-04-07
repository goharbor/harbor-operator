package v1alpha1

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type ComponentWithRedis Component

const (
	CoreRedis        = ComponentWithRedis(CoreComponent)
	JobServiceRedis  = ComponentWithRedis(JobServiceComponent)
	RegistryRedis    = ComponentWithRedis(RegistryComponent)
	ChartMuseumRedis = ComponentWithRedis(ChartMuseumComponent)
	TrivyRedis       = ComponentWithRedis(TrivyComponent)
)

type RedisHostSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// Server hostname.
	Host string `json:"host"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=true
	// Server port.
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Optional
	// for Sentinel MasterSet.
	SentinelMasterSet string `json:"sentinelMasterSet,omitempty"`
}

type RedisCredentials struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Secret containing the password to use when connecting to the server.
	PasswordRef string `json:"passwordRef,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Secret containing the client certificate to authenticate with.
	CertificateRef string `json:"certificateRef,omitempty"`
}

type RedisConnection struct {
	RedisHostSpec    `json:",inline"`
	RedisCredentials `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=8
	// +kubebuilder:default=0
	// The database number.
	Database int32 `json:"database,omitempty"`
}

const (
	RedisScheme         = "redis"
	RedisSentinelScheme = "redis+sentinel"
)

func (c *RedisConnection) GetDSNNoCredentials() *url.URL {
	if c == nil {
		return nil
	}

	u := &url.URL{
		Host: fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path: strconv.Itoa(int(c.Database)),
	}
	if c.SentinelMasterSet == "" {
		u.Scheme = RedisScheme
	} else {
		u.Scheme = RedisSentinelScheme
		u.Path = c.SentinelMasterSet + "/" + u.Path
	}

	return u
}

func (c *RedisConnection) GetDSN(password string) *url.URL {
	if c == nil {
		return nil
	}

	u := c.GetDSNNoCredentials()

	if password != "" {
		u.User = url.UserPassword(u.User.Username(), password)
	}

	return u
}

func (c *RedisConnection) GetDSNStringWithRawPassword(password string) string {
	if c == nil {
		return ""
	}

	u := c.GetDSNNoCredentials()

	username := u.User.Username()
	u.User = nil
	noPrefix := strings.TrimPrefix(u.String(), fmt.Sprintf("%s://", u.Scheme))
	result := fmt.Sprintf("%s://%s:%s@%s", u.Scheme, url.User(username).String(), password, noPrefix)

	return result
}

const (
	coreRedisDatabaseIndex        = 0
	registryRedisDatabaseIndex    = 1
	jobServiceRedisDatabaseIndex  = 2
	chartMuseumRedisDatabaseIndex = 3
	trivyRedisDatabaseIndex       = 5
)

func (r ComponentWithRedis) Index() int32 {
	return map[ComponentWithRedis]int32{
		CoreRedis:        coreRedisDatabaseIndex,
		JobServiceRedis:  jobServiceRedisDatabaseIndex,
		RegistryRedis:    registryRedisDatabaseIndex,
		ChartMuseumRedis: chartMuseumRedisDatabaseIndex,
		TrivyRedis:       trivyRedisDatabaseIndex,
	}[r]
}

func (r ComponentWithRedis) String() string {
	return Component(r).String()
}
