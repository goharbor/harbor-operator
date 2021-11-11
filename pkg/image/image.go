package image

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	ImageSourceRepositoryEnvKey = "IMAGE_SOURCE_REPOSITORY"
	ImageSourceTagSuffixEnvKey  = "IMAGE_SOURCE_TAG_SUFFIX"
)

type Options struct {
	imageFromSpec string
	repository    string
	tagSuffix     string
	harborVersion string
}

type Option func(*Options)

func WithImageFromSpec(imageFromSpec string) Option {
	return func(opts *Options) {
		opts.imageFromSpec = imageFromSpec
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

	// imageFromSpec is the image from the spec of the component
	if opts.imageFromSpec != "" {
		return opts.imageFromSpec, nil
	}

	if opts.harborVersion == "" {
		return "", errors.Errorf("missing harbor version in component %s", component)
	}

	if opts.repository == "" {
		repository := os.Getenv(ImageSourceRepositoryEnvKey)
		if repository != "" {
			opts.repository = repository
		} else {
			opts.repository = knownComponents.Get(component, repositoryKind, opts.harborVersion)
		}
	}

	if opts.tagSuffix == "" {
		tagSuffix := os.Getenv(ImageSourceTagSuffixEnvKey)
		if tagSuffix != "" {
			opts.tagSuffix = tagSuffix
		}
	}

	imageName := knownComponents.Get(component, imageNameKind, opts.harborVersion)
	if imageName == "" {
		return "", errors.Errorf("unknown component %s", component)
	}

	repository := opts.repository
	if repository != "" && !strings.HasSuffix(repository, "/") {
		repository += "/"
	}

	tag := knownComponents.Get(component, tagKind, opts.harborVersion, "v"+opts.harborVersion)

	return fmt.Sprintf("%s%s:%s%s", repository, imageName, tag, opts.tagSuffix), nil
}
