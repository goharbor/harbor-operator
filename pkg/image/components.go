package image

import (
	"sync"

	"github.com/Masterminds/semver"
)

type metadataKind int

const (
	repositoryKind metadataKind = iota
	imageNameKind
	tagKind
)

func mustConstraint(c string) *semver.Constraints {
	o, err := semver.NewConstraint(c)
	if err != nil {
		panic(err)
	}

	return o
}

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
		for c, value := range md.values[kind] {
			if c.Check(v) {
				return value
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

var knownComponents = components{mds: map[string]*metadata{}}

func RegisterImageName(component, imageName string, harborVersions ...string) {
	knownComponents.Register(component, imageNameKind, imageName, harborVersions...)
}

func RegisterRepository(component, repository string, harborVersions ...string) {
	knownComponents.Register(component, repositoryKind, repository, harborVersions...)
}

func RegisterTag(component, tag string, harborVersions ...string) {
	knownComponents.Register(component, tagKind, tag, harborVersions...)
}

func init() { //nolint:gochecknoinits
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
	RegisterTag("cluster-redis", "5.0-alpine", "~2.2.0", "~2.3.0", "~2.4.0", "~2.5.0", "~2.6.0")

	RegisterRepository("cluster-postgresql", "registry.opensource.zalan.do/acid", "*")
	RegisterImageName("cluster-postgresql", "spilo-13", "*")
	RegisterTag("cluster-postgresql", "2.1-p1", "~2.2.0", "~2.3.0", "~2.4.0", "~2.5.0", "~2.6.0")

	RegisterRepository("cluster-minio", "minio", "*") // the minio repository of dockerhub
	RegisterImageName("cluster-minio", "minio", "*")
	RegisterTag("cluster-minio", "RELEASE.2022-08-26T19-53-15Z", "~2.2.0", "~2.3.0", "~2.4.0", "~2.5.0", "~2.6.0")

	RegisterRepository("cluster-minio-init", "minio", "*") // the minio repository of dockerhub
	RegisterImageName("cluster-minio-init", "mc", "*")
	RegisterTag("cluster-minio-init", "RELEASE.2022-08-23T05-45-20Z", "~2.2.0", "~2.3.0", "~2.4.0", "~2.5.0", "~2.6.0")
}
