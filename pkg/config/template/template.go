package template

import (
	"net/http"
	"os"
	"path"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

const ConfigTemplateKey = "template-content"

type ConfigTemplate struct {
	Path      string
	Key       string
	Priority  int64
	refreshed bool
}

func New(path string) *ConfigTemplate {
	return &ConfigTemplate{
		Path:      path,
		Key:       ConfigTemplateKey,
		Priority:  config.DefaultPriority,
		refreshed: false,
	}
}

const ConfigTemplatePathKey = "template-path"

func FromConfigStore(store *configstore.Store, defaultFileName string) (*ConfigTemplate, error) {
	templateDir, err := config.GetString(store, config.TemplateDirectoryKey, config.DefaultTemplateDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "directory")
	}

	defaultTemplatePath := path.Join(templateDir, defaultFileName)

	configTemplatePath, err := config.GetString(store, ConfigTemplatePathKey, defaultTemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "path")
	}

	fileStat, err := os.Stat(configTemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "stat")
	}

	if fileStat.IsDir() {
		return nil, &ErrNotValidFile{Path: configTemplatePath}
	}

	return New(configTemplatePath), nil
}

func (t *ConfigTemplate) Register(store *configstore.Store) {
	store.FileCustomRefresh(t.Path, func(data []byte) ([]configstore.Item, error) {
		t.refreshed = true

		return []configstore.Item{configstore.NewItem(t.Key, string(data), t.Priority)}, nil
	})
}

var _ healthz.Checker = (&ConfigTemplate{}).ReadyzCheck

func (t *ConfigTemplate) ReadyzCheck(req *http.Request) error {
	if t.refreshed {
		return nil
	}

	return ErrNotYetRefreshed{}
}

var _ healthz.Checker = (&ConfigTemplate{}).HealthzCheck

func (t *ConfigTemplate) HealthzCheck(req *http.Request) error {
	fileStat, err := os.Stat(t.Path)
	if err != nil {
		return errors.Wrap(err, "stat")
	}

	if fileStat.IsDir() {
		return &ErrNotValidFile{Path: t.Path}
	}

	return nil
}
