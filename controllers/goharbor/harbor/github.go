package harbor

import (
	"github.com/ovh/configstore"
)

const GithubCredentialsConfigKey = "github-token"

func (r *Reconciler) GetDefaultGithubToken() (string, error) {
	token, err := configstore.Filter().GetItemValue(GithubCredentialsConfigKey)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *Reconciler) GetGithubToken(configKey string) (string, error) {
	token, err := configstore.Filter().GetItemValue(configKey)
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); ok {
			return r.GetDefaultGithubToken()
		}

		return "", err
	}

	return token, nil
}
