package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/ovh/configstore"
)

const (
	ConfigImageKeyPrefix = "docker-image"
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
	"trivy":        "trivy-adapter-photon",
}

type Options struct {
	imageFromSpec  string
	configStore    *configstore.Store
	configImageKey string
	repository     string
	tagSuffix      string
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

func WithTagSuffix(tagSuffix string) Option {
	return func(opts *Options) {
		opts.tagSuffix = tagSuffix
	}
}

func WithHarborVersion(version string) Option {
	return func(opts *Options) {
		opts.harborVersion = version
	}
}

func GetImage(ctx context.Context, component string, options ...Option) (string, error) {
	opts := &Options{}

	for _, o := range options {
		o(opts)
	}

	if opts.repository == "" {
		opts.repository = "goharbor"
	}

	if opts.harborVersion == "" {
		opts.harborVersion = version.Default()
	}

	if opts.configImageKey == "" {
		// config image key with version, eg docker-image-2-1-0
		opts.configImageKey = ConfigImageKeyPrefix + "-" + strings.ReplaceAll(opts.harborVersion, ".", "-")
	}

	// imageFromSpec is the image from the spec of the component
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

	repository := strings.TrimSuffix(opts.repository, "/")

	return fmt.Sprintf("%s/%s:v%s%s", repository, imageName, opts.harborVersion, opts.tagSuffix), nil
}
