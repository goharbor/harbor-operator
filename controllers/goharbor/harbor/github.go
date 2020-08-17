package harbor

import (
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
)

const GithubCredentialsConfigKey = "github-credentials"

type GithubCredentials struct {
	User  string `json:"user"`
	Token string `json:"token"`
}

func (r *Reconciler) GetDefaultGithubCredentials() (*GithubCredentials, error) {
	item, err := configstore.Filter().
		Slice(GithubCredentialsConfigKey).
		Unmarshal(func() interface{} {
			return &GithubCredentials{}
		}).
		GetFirstItem()
	if err != nil {
		return nil, err
	}

	config, err := item.Unmarshaled()
	if err != nil {
		return nil, errors.Wrap(err, "invalid")
	}

	return config.(*GithubCredentials), nil
}

func (r *Reconciler) GetGithubCredentials(configKey string) (*GithubCredentials, error) {
	item, err := configstore.Filter().
		Slice(configKey).
		Unmarshal(func() interface{} {
			return &GithubCredentials{}
		}).
		GetFirstItem()
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			return r.GetDefaultGithubCredentials()
		}

		return nil, err
	}

	config, err := item.Unmarshaled()
	if err != nil {
		return nil, errors.Wrap(err, "invalid")
	}

	return config.(*GithubCredentials), nil
}
