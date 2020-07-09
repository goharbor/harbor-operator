package v1alpha2

import (
	"fmt"
	"net/url"
	"strings"
)

type OpacifiedDSN struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=".+://.+"
	DSN string `json:"dsn"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// If password is absent from dsn field, please specify it here
	PasswordRef string `json:"passwordRef,omitempty"`
}

func (n *OpacifiedDSN) GetDSN(password string) (*url.URL, error) {
	u, err := url.Parse(n.DSN)
	if err != nil {
		return nil, err
	}

	if password != "" {
		u.User = url.UserPassword(u.User.Username(), password)
	}

	return u, nil
}

func (n *OpacifiedDSN) GetDSNStringWithRawPassword(passwordEnvName string) (string, error) {
	u, err := url.Parse(n.DSN)
	if err != nil {
		return "", err
	}

	username := u.User.Username()
	u.User = nil
	noPrefix := strings.TrimPrefix(u.String(), fmt.Sprintf("%s://", u.Scheme))
	result := fmt.Sprintf("%s://%s:%s@%s", u.Scheme, url.User(username).String(), passwordEnvName, noPrefix)

	return result, nil
}
