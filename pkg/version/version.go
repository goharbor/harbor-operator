package version

import (
	"github.com/pkg/errors"
)

const (
	versionAnnotationKey = "harbor.goharbor.io/version"
)

var (
	knowVersions       = []string{}
	knowVersionIndexes = map[string]int{}
	defaultVersion     = ""
)

func init() { // nolint:gochecknoinits
	RegisterKnowVersions(
		"2.2.1",
	)
}

// RegisterKnowVersions register the know versions.
// NOTE: the paramater versions must be in increasing order.
func RegisterKnowVersions(versions ...string) {
	for i, version := range versions {
		knowVersions = append(knowVersions, version)
		knowVersionIndexes[version] = i
	}

	if len(knowVersions) > 0 {
		defaultVersion = knowVersions[len(knowVersions)-1]
	}
}

// Default returns the default version.
func Default() string {
	return defaultVersion
}

// Validate returns nil when version is the default version.
func Validate(version string) error {
	if version != Default() {
		return errors.Errorf("version %s not support, please use version %s", version, Default())
	}

	return nil
}

// UpgradeAllowed returns nil when upgrade allowed.
func UpgradeAllowed(from, to string) error {
	fromIndex, ok := knowVersionIndexes[from]
	if !ok {
		return errors.Errorf("unknow version %s", from)
	}

	toIndex, ok := knowVersionIndexes[to]
	if !ok {
		return errors.Errorf("unknow version %s", to)
	}

	if toIndex < fromIndex {
		return errors.Errorf("downgrade from version %s to %s is not allowed", from, to)
	}

	return nil
}

// GetVersion returns version from the annotations.
func GetVersion(annotations map[string]string) string {
	if annotations == nil {
		return ""
	}

	return annotations[versionAnnotationKey]
}

// SetVersion set version to the annotations.
func SetVersion(annotations map[string]string, version string) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[versionAnnotationKey] = version

	return annotations
}

// NewVersionAnnotations returns annotations only with version annotation when the version annotation found in from annotations.
func NewVersionAnnotations(from map[string]string) map[string]string {
	v := GetVersion(from)
	if v == "" {
		return nil
	}

	return map[string]string{
		versionAnnotationKey: v,
	}
}
