// This package contains structure to describe a connection to a Postgres database
package v1alpha1

import (
	"fmt"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DatabaseName string

const (
	CoreDatabase         = "core"
	NotaryServerDatabase = "notaryserver"
	NotarySignerDatabase = "notarysigner"
)

type ErrPostgresNoHost bool

func (err *ErrPostgresNoHost) Error() string {
	return "postgres: no host found"
}

func NewErrPostgresNoHost() *ErrPostgresNoHost {
	err := ErrPostgresNoHost(false)

	return &err
}

type PostgresHostSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// Name of host to connect to.
	// If a host name begins with a slash, it specifies Unix-domain communication rather than
	// TCP/IP communication; the value is the name of the directory in which the socket file is stored.
	Host string `json:"host"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=true
	// Port number to connect to at the server host,
	// or socket file name extension for Unix-domain connections.
	// Zero, specifies the default port number established when PostgreSQL was built.
	Port int32 `json:"port,omitempty"`
}

type PostgresCredentials struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// PostgreSQL user name to connect as.
	// Defaults to be the same as the operating system name of the user running the application.
	Username string `json:"username,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// Secret containing the password to be used if the server demands password authentication.
	PasswordRef string `json:"passwordRef,omitempty"`
}

func (p *PostgresCredentials) GetPasswordEnvVarSource() *corev1.EnvVarSource {
	return &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			Key: PostgresqlPasswordKey,
			LocalObjectReference: corev1.LocalObjectReference{
				Name: p.PasswordRef,
			},
		},
	}
}

type PostgresConnection struct {
	PostgresCredentials `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Hosts []PostgresHostSpec `json:"hosts,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// The database name. Defaults to be the same as the user name.
	// In certain contexts, the value is checked for extended formats.
	Database string `json:"database,omitempty"`
}

const PostgresScheme = "postgres"

func (c *PostgresConnection) GetDSNNoCredentials() *url.URL {
	if c == nil {
		return nil
	}

	hosts := []string{}
	for _, host := range c.Hosts {
		hosts = append(hosts, fmt.Sprintf("%s:%d", host.Host, host.Port))
	}

	u := &url.URL{
		Scheme: PostgresScheme,
		Host:   strings.Join(hosts, ","),
		Path:   c.Database,
	}

	if c.Username != "" {
		u.User = url.User(c.Username)
	}

	return u
}

func (c *PostgresConnection) GetDSN(password string) *url.URL {
	if c == nil {
		return nil
	}

	u := c.GetDSNNoCredentials()

	if password != "" {
		u.User = url.UserPassword(u.User.Username(), password)
	}

	return u
}

func (c *PostgresConnection) GetDSNStringWithRawPassword(password string) string {
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

type PostgresConnectionWithParameters struct {
	PostgresConnection `json:",inline"`

	// +kubebuilder:validation:Optional
	// libpq parameters.
	Parameters map[string]string `json:"parameters,omitempty"`
}

func (c *PostgresConnectionWithParameters) GetDSNNoCredentials() *url.URL {
	if c == nil {
		return nil
	}

	u := c.PostgresConnection.GetDSNNoCredentials()

	return c.addParameters(u)
}

func (c *PostgresConnectionWithParameters) GetDSN(password string) *url.URL {
	if c == nil {
		return nil
	}

	u := c.PostgresConnection.GetDSN(password)

	return c.addParameters(u)
}

func (c *PostgresConnectionWithParameters) GetDSNStringWithRawPassword(password string) string {
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

func (c *PostgresConnectionWithParameters) addParameters(u *url.URL) *url.URL {
	if len(c.Parameters) == 0 {
		return u
	}

	query := u.Query()

	for key, value := range c.Parameters {
		query.Add(key, value)
	}

	u.RawQuery = query.Encode()
	u.ForceQuery = true

	return u
}

/*
  Following types describe libpq parameters keywords
  https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
*/

// +kubebuilder:validation:Optional
// Maximum wait for connection.
// Zero, negative, or not specified means wait indefinitely.
// The minimum allowed timeout is 2 seconds, therefore a value of 1 is interpreted as 2.
// This timeout applies separately to each host name or IP address.
// For example, if you specify two hosts and connect_timeout is 5,
// each host will time out if no connection is made within 5 seconds,
// so the total time spent waiting for a connection might be up to 10 seconds.
type PostgresConnectTimeout metav1.Duration

const PostgresConnectTimeoutKey = "connect_timeout"

func (p *PostgresConnectTimeout) Add(query url.Values) {
	if p == nil {
		return
	}

	query.Add(PostgresConnectTimeoutKey, fmt.Sprintf("%.0f", p.Duration.Seconds()))
}

// +kubebuilder:default="auto"
// This sets the client_encoding configuration parameter for this connection.
// In addition to the values accepted by the corresponding server option,
// you can use auto to determine the right encoding from the current locale
// in the client (LC_CTYPE environment variable on Unix systems).
type PostgresClientEncoding string

func (p *PostgresClientEncoding) Add(query url.Values) {
	if p == nil {
		return
	}

	query.Add(PostgresConnectTimeoutKey, string(*p))
}

// +kubebuilder:validation:Enum={"disable","allow","prefer","require","verify-ca","verify-full"}
// +kubebuilder:default="prefer"
// PostgreSQL has native support for using SSL connections to encrypt client/server communications for increased security.
type PostgresSSLMode string

const PostgresSSLModeKey = "sslmode"

const (
	PostgresSSLModeDisable    PostgresSSLMode = "disable"
	PostgresSSLModeAllow      PostgresSSLMode = "allow"
	PostgresSSLModePrefer     PostgresSSLMode = "prefer"
	PostgresSSLModeRequire    PostgresSSLMode = "require"
	PostgresSSLModeVerifyCA   PostgresSSLMode = "verify-ca"
	PostgresSSLModeVerifyFull PostgresSSLMode = "verify-full"
)

func (p *PostgresSSLMode) Add(query url.Values) {
	if p == nil {
		return
	}

	query.Add(PostgresConnectTimeoutKey, string(*p))
}
