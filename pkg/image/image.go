package image

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/ovh/configstore"
)

const (
	ConfigImageKey = "docker-image"
)

type metadata struct {
	Repositories map[string]string // key is the harbor version, value is the repository
	ImageNames   map[string]string // key is the harbor version, value is the image name
	Tags         map[string]string // key is the harbor version, value is the image tag
}

func makeMetadata() *metadata {
	return &metadata{
		Repositories: map[string]string{},
		ImageNames:   map[string]string{},
		Tags:         map[string]string{},
	}
}

type components struct {
	lock sync.RWMutex
	mds  map[string]*metadata // key is the component name
}

func (cs *components) GetImageName(component string, harborVersion string) (string, bool) {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	md, ok := cs.mds[component]
	if ok {
		for _, version := range []string{harborVersion, "*"} {
			if imageName, ok := md.ImageNames[version]; ok {
				return imageName, true
			}
		}
	}

	return "", false
}

func (cs *components) GetRepository(component string, harborVersion string) string {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	md, ok := cs.mds[component]
	if ok {
		for _, version := range []string{harborVersion, "*"} {
			if repository, ok := md.Repositories[version]; ok {
				return repository
			}
		}
	}

	return ""
}

func (cs *components) GetTag(component string, harborVersion string) string {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	md, ok := cs.mds[component]
	if ok {
		for _, version := range []string{harborVersion, "*"} {
			if tag, ok := md.Tags[version]; ok {
				return tag
			}
		}
	}

	return "v" + harborVersion
}

func (cs *components) RegisterImageName(component string, imageName string, harborVersions ...string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	md, ok := cs.mds[component]
	if !ok {
		md = makeMetadata()
		cs.mds[component] = md
	}

	for _, harborVersion := range harborVersions {
		md.ImageNames[harborVersion] = imageName
	}
}

func (cs *components) RegisterRepository(component, repository string, harborVersions ...string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	md, ok := cs.mds[component]
	if !ok {
		md = makeMetadata()
		cs.mds[component] = md
	}

	for _, harborVersion := range harborVersions {
		md.Repositories[harborVersion] = repository
	}
}

func (cs *components) RegisterTag(component string, tag string, harborVersions ...string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	md, ok := cs.mds[component]
	if !ok {
		md = makeMetadata()
		cs.mds[component] = md
	}

	for _, harborVersion := range harborVersions {
		md.Tags[harborVersion] = tag
	}
}

var knowCompoents = components{mds: map[string]*metadata{}}

func RegisterImageName(component, imageName string, harborVersions ...string) {
	knowCompoents.RegisterImageName(component, imageName, harborVersions...)
}

func RegisterRepsitory(component, repository string, harborVersions ...string) {
	knowCompoents.RegisterRepository(component, repository, harborVersions...)
}

func RegisterTag(component, tag string, harborVersions ...string) {
	knowCompoents.RegisterTag(component, tag, harborVersions...)
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
		RegisterRepsitory(component, "goharbor", "*") // the goharbor repository of dockerhub
		RegisterImageName(component, imageName, "*")
	}

	// Register the cluster service components
	RegisterRepsitory("cluster-redis", "", "*") // the - repository of dockerhub
	RegisterImageName("cluster-redis", "redis", "*")
	RegisterTag("cluster-redis", "5.0-alpine", "2.2.1")

	RegisterRepsitory("cluster-postgresql", "registry.opensource.zalan.do/acid", "*")
	RegisterImageName("cluster-postgresql", "spilo-12", "*")
	RegisterTag("cluster-postgresql", "1.6-p3", "2.2.1")

	RegisterRepsitory("cluster-minio", "minio", "*") // the minio repository of dockerhub
	RegisterImageName("cluster-minio", "minio", "*")
	RegisterTag("cluster-minio", "RELEASE.2021-04-06T23-11-00Z", "2.2.1")
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

	if opts.harborVersion == "" {
		opts.harborVersion = version.Default()
	}

	if opts.repository == "" {
		opts.repository = knowCompoents.GetRepository(component, opts.harborVersion)
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

	imageName, ok := knowCompoents.GetImageName(component, opts.harborVersion)
	if !ok {
		return "", fmt.Errorf("unknow component %s", component)
	}

	repository := opts.repository
	if repository != "" && !strings.HasSuffix(repository, "/") {
		repository += "/"
	}

	return fmt.Sprintf("%s%s:%s%s", repository, imageName, knowCompoents.GetTag(component, opts.harborVersion), opts.tagSuffix), nil
}
