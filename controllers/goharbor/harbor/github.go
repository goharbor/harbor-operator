package harbor

import (
	"github.com/goharbor/harbor-operator/pkg/config"
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
		if config.IsNotFound(err, configKey) {
			return r.GetDefaultGithubToken()
		}

		return "", err
	}

	return token, nil
}
