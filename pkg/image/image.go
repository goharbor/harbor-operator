package image

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
)

const (
	ConfigImageKey = "docker-image"
)

type metadataKind int

const (
	repositoryKind metadataKind = iota
	imageNameKind
	tagKind
)

type metadata struct {
	values map[metadataKind]map[*semver.Constraints]string
}

func (md *metadata) Add(kind metadataKind, value string, harborVersions ...string) {
	for _, harborVersion := range harborVersions {
		md.values[kind][mustConstraint(harborVersion)] = value
	}
}

func (md *metadata) Get(kind metadataKind, harborVersion string, defaultValues ...string) string {
	v, err := semver.NewVersion(harborVersion)
	if err == nil {
		for c, imageName := range md.values[kind] {
			if c.Check(v) {
				return imageName
			}
		}
	}

	if len(defaultValues) > 0 {
		return defaultValues[0]
	}

	return ""
}

func makeMetadata() *metadata {
	return &metadata{
		values: map[metadataKind]map[*semver.Constraints]string{
			repositoryKind: {},
			imageNameKind:  {},
			tagKind:        {},
		},
	}
}

func mustConstraint(c string) *semver.Constraints {
	o, err := semver.NewConstraint(c)
	if err != nil {
		panic(err)
	}

	return o
}

type components struct {
	lock sync.RWMutex
	mds  map[string]*metadata // key is the component name
}

func (cs *components) metadata(component string) *metadata {
	md, ok := cs.mds[component]
	if !ok {
		md = makeMetadata()
		cs.mds[component] = md
	}

	return md
}

func (cs *components) Get(component string, kind metadataKind, harborVersion string, defaultValues ...string) string {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.metadata(component).Get(kind, harborVersion, defaultValues...)
}

func (cs *components) Register(component string, kind metadataKind, value string, harborVersions ...string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cs.metadata(component).Add(kind, value, harborVersions...)
}

var knownCompoents = components{mds: map[string]*metadata{}}

func RegisterImageName(component, imageName string, harborVersions ...string) {
	knownCompoents.Register(component, imageNameKind, imageName, harborVersions...)
}

func RegisterRepository(component, repository string, harborVersions ...string) {
	knownCompoents.Register(component, repositoryKind, repository, harborVersions...)
}

func RegisterTag(component, tag string, harborVersions ...string) {
	knownCompoents.Register(component, tagKind, tag, harborVersions...)
}

func init() { // nolint:gochecknoinits
	// Register the harbor components
	harborComponentImageNames := map[string]string{
		"chartmuseum":  "chartmuseum-photon",
		"core":         "harbor-core",
		"exporter":     "harbor-exporter",
		"jobservice":   "harbor-jobservice",
		"notaryserver": "notary-server-photon",
		"notarysigner": "notary-signer-photon",
		"portal":       "harbor-portal",
		"registry":     "registry-photon",
		"registryctl":  "harbor-registryctl",
		"trivy":        "trivy-adapter-photon",
	}
	for component, imageName := range harborComponentImageNames {
		RegisterRepository(component, "goharbor", "*") // the goharbor repository of dockerhub
		RegisterImageName(component, imageName, "*")
	}

	// Register the cluster service components
	RegisterRepository("cluster-redis", "", "*") // the - repository of dockerhub
	RegisterImageName("cluster-redis", "redis", "*")
	RegisterTag("cluster-redis", "5.0-alpine", "~2.2.0")

	RegisterRepository("cluster-postgresql", "registry.opensource.zalan.do/acid", "*")
	RegisterImageName("cluster-postgresql", "spilo-12", "*")
	RegisterTag("cluster-postgresql", "1.6-p3", "~2.2.0")

	RegisterRepository("cluster-minio", "minio", "*") // the minio repository of dockerhub
	RegisterImageName("cluster-minio", "minio", "*")
	RegisterTag("cluster-minio", "RELEASE.2021-04-06T23-11-00Z", "~2.2.0")

	RegisterRepository("cluster-minio-init", "minio", "*") // the minio repository of dockerhub
	RegisterImageName("cluster-minio-init", "mc", "*")
	RegisterTag("cluster-minio-init", "RELEASE.2021-03-23T05-46-11Z", "~2.2.0")
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
		opts.repository = knownCompoents.Get(component, repositoryKind, opts.harborVersion)
	}

	if opts.configImageKey == "" {
		opts.configImageKey = ConfigImageKey
	}

	// imageFromSpec is the image from the spec of the component
	if opts.imageFromSpec != "" {
		return opts.imageFromSpec, nil
	}

	if opts.configStore != nil {
		// config image key with version, eg docker-image-2-1-0
		configImageKey := opts.configImageKey + "-" + strings.ReplaceAll(opts.harborVersion, ".", "-")

		image, err := opts.configStore.GetItemValue(configImageKey)
		if err == nil {
			return image, nil
		}

		if !config.IsNotFound(err, configImageKey) {
			return "", err
		}
	}

	imageName := knownCompoents.Get(component, imageNameKind, opts.harborVersion)
	if imageName == "" {
		return "", errors.Errorf("unknow component %s", component)
	}

	if opts.harborVersion == "" {
		return "", errors.Errorf("missing harbor version in component %s", component)
	}

	repository := opts.repository
	if repository != "" && !strings.HasSuffix(repository, "/") {
		repository += "/"
	}

	tag := knownCompoents.Get(component, tagKind, opts.harborVersion, "v"+opts.harborVersion)

	return fmt.Sprintf("%s%s:%s%s", repository, imageName, tag, opts.tagSuffix), nil
}
