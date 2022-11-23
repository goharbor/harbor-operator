package version

import (
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

const (
	versionAnnotationKey = "harbor.goharbor.io/version"
)

var (
	knownConstraints []*semver.Constraints
	latestConstraint *semver.Constraints
)

func init() { //nolint:gochecknoinits
	RegisterKnownConstraints(
		"~2.2.x",
		"~2.3.x",
		"~2.4.x",
		"~2.5.x",
		"~2.6.x",
	)
}

func parseVersion(version string) (*semver.Version, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}

	for _, knownConstraint := range knownConstraints {
		if knownConstraint.Check(v) {
			return v, nil
		}
	}

	return nil, errors.Errorf("unknown version %s", version)
}

// RegisterKnownConstraints register the know constraints.
// NOTE: the parameter constraints must be in increasing order.
func RegisterKnownConstraints(versions ...string) {
	knownConstraints = []*semver.Constraints{}

	for _, version := range versions {
		c, err := semver.NewConstraint(version)
		if err != nil {
			panic(c)
		}

		knownConstraints = append(knownConstraints, c)
	}

	if len(knownConstraints) > 0 {
		latestConstraint = knownConstraints[len(knownConstraints)-1]
	}
}

// Validate returns nil when version is the default version.
func Validate(version string) error {
	v, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	if valid, errs := latestConstraint.Validate(v); !valid {
		return errs[0]
	}

	return nil
}

// UpgradeAllowed returns nil when upgrade allowed.
func UpgradeAllowed(from, to string) error {
	fromVersion, err := parseVersion(from)
	if err != nil {
		return err
	}

	// only allowed to upgrade to latest major and minor version
	toVersion, err := parseVersion(to)
	if err != nil {
		return err
	}

	if !fromVersion.Equal(toVersion) {
		if valid, errs := latestConstraint.Validate(toVersion); !valid {
			return errors.Errorf("upgrade from %s to %s is not allowed, error: %v", from, to, errs[0])
		}

		if fromVersion.GreaterThan(toVersion) {
			return errors.Errorf("downgrade from %s to %s is not allowed", from, to)
		}
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
