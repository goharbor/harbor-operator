package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
)

const (
	ConfigImageKey = "docker-image"
)

var componentImageNames = map[string]string{
	"chartmuseum":  "chartmuseum-photon",
	"core":         "harbor-core",
	"jobservice":   "harbor-jobservice",
	"notaryserver": "notary-server-photon",
	"notarysigner": "notary-signer-photon",
	"portal":       "harbor-portal",
	"registry":     "registry-photon",
	"registryctl":  "harbor-registryctl",
	"trivy":        "goharbor/trivy-adapter-photon",
}

type Options struct {
	imageFromSpec  string
	configStore    *configstore.Store
	configImageKey string
	repository     string
	tag            string
	harborVersion  string
}

type Option func(*Options)

func WithImageFromSpec(imageFromSpec string) Option {
	return func(opts *Options) {
		opts.imageFromSpec = imageFromSpec
	}
}

func WithConfigstore(configStore *configstore.Store) Option {
	return func(opts *Options) {
		opts.configStore = configStore
	}
}

func WithConfigImageKey(configImageKey string) Option {
	return func(opts *Options) {
		opts.configImageKey = configImageKey
	}
}

func WithRepository(repository string) Option {
	return func(opts *Options) {
		opts.repository = repository
	}
}

func WithTag(tag string) Option {
	return func(opts *Options) {
		opts.tag = tag
	}
}

func WithHarborVersion(version string) Option {
	return func(opts *Options) {
		opts.harborVersion = version
	}
}

func GetImage(ctx context.Context, component string, options ...Option) (string, error) {
	opts := &Options{
		repository:     config.DefaultRegistry + "/goharbor",
		configImageKey: ConfigImageKey,
	}

	for _, o := range options {
		o(opts)
	}

	if opts.imageFromSpec != "" {
		return opts.imageFromSpec, nil
	}

	if opts.configStore != nil {
		image, err := opts.configStore.GetItemValue(opts.configImageKey)
		if err == nil {
			return image, nil
		}

		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return "", err
		}
	}

	imageName, ok := componentImageNames[component]
	if !ok {
		return "", fmt.Errorf("unknow component %s", component)
	}

	repository := strings.Trim(opts.repository, "/")

	if opts.tag != "" {
		return fmt.Sprintf("%s/%s:%s", repository, imageName, opts.tag), nil
	} else if opts.harborVersion != "" {
		return fmt.Sprintf("%s/%s:v%s", repository, imageName, opts.harborVersion), nil
	}

	return fmt.Sprintf("%s/%s:%s", repository, imageName, config.DefaultHarborVersion), nil
}
