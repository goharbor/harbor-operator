package harbor

import (
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
)

const NotaryMigrationGithubCredentialsConfigKey = "notary-migration-github-credentials"

const NotaryMigrationGithubSourceConfigKey = "notary-migration-github-source"

type GithubSource struct {
	Owner      string `json:"owner"`
	Repository string `json:"repository"`
	Path       string `json:"path"`
	Reference  string `json:"reference"`
}

func (r *Reconciler) GetDefaultNotaryMigrationSource() (*GithubSource, error) {
	defaultSource := GithubSource{
		Owner:      "theupdateframework",
		Repository: "notary",
		Path:       "/migrations/server/postgresql",
		Reference:  "v0.6.1",
	}

	item, err := configstore.Filter().
		Slice(NotaryMigrationGithubSourceConfigKey).
		Unmarshal(func() interface{} {
			copy := defaultSource

			return &copy
		}).
		GetFirstItem()
	if err != nil {
		if config.IsNotFound(err, NotaryMigrationGithubSourceConfigKey) {
			return &defaultSource, nil
		}

		return nil, err
	}

	config, err := item.Unmarshaled()
	if err != nil {
		return nil, errors.Wrap(err, "invalid")
	}

	return config.(*GithubSource), nil
}
