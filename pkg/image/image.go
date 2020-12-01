package image

import (
	"fmt"

	"github.com/blang/semver"
)

// Getter will proxy the Locator.
type Getter interface {
	Locator
}

// GetterImpl contains the concrete Locator instance,
// if registry is not null, all the methods in Getter will be wrapped to add the registry prefix.
type GetterImpl struct {
	locator       Locator
	registry      *string
	harborVersion string
}

func NewImageGetter(registry *string, harborVersion string) (Getter, error) {
	// The version should be validated at the spec level to make sure it's in the supported list
	// or keep the current returns
	hv, err := semver.Parse(harborVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid harbor version in the CR: %w", err)
	}

	versionRange, err := semver.ParseRange(">1.10.0 <2.0.0")
	if err != nil {
		return nil, fmt.Errorf("invalid harbor version range: %w", err)
	}

	// As "1.10.0" is compatible with semver, ignore the make error
	vM1m10p0, _ := semver.Make("1.10.0")

	var locator Getter

	switch {
	case hv.Compare(vM1m10p0) == 0:
		locator = &harborV1_10_0ImageLocator{}
	case versionRange(hv):
		locator = &harborVM1m10pxImageLocator{
			HarborVersion: harborVersion,
		}
	default:
		return nil, fmt.Errorf("failed to get relate images with harbor version %s, only support '1.10.x'", harborVersion)
	}

	return &GetterImpl{
		locator:       locator,
		registry:      registry,
		harborVersion: harborVersion,
	}, nil
}

func (i *GetterImpl) CoreImage() string {
	return GetImage(i.registry, i.locator.CoreImage())
}

func (i *GetterImpl) ChartMuseumImage() string {
	return GetImage(i.registry, i.locator.ChartMuseumImage())
}

func (i *GetterImpl) ClairImage() string {
	return GetImage(i.registry, i.locator.ClairImage())
}

func (i *GetterImpl) ClairAdapterImage() string {
	return GetImage(i.registry, i.locator.ClairAdapterImage())
}

func (i *GetterImpl) JobServiceImage() string {
	return GetImage(i.registry, i.locator.JobServiceImage())
}

func (i *GetterImpl) NotaryServerImage() string {
	return GetImage(i.registry, i.locator.NotaryServerImage())
}

func (i *GetterImpl) NotarySingerImage() string {
	return GetImage(i.registry, i.locator.NotarySingerImage())
}

func (i *GetterImpl) NotaryDBMigratorImage() string {
	return GetImage(i.registry, i.locator.NotaryDBMigratorImage())
}

func (i *GetterImpl) PortalImage() string {
	return GetImage(i.registry, i.locator.PortalImage())
}

func (i *GetterImpl) RegistryImage() string {
	return GetImage(i.registry, i.locator.RegistryImage())
}

func (i *GetterImpl) RegistryControllerImage() string {
	return GetImage(i.registry, i.locator.RegistryControllerImage())
}

// Locator provider method to get harbor component image.
type Locator interface {
	CoreImage() string
	ChartMuseumImage() string
	ClairImage() string
	ClairAdapterImage() string
	JobServiceImage() string
	NotaryServerImage() string
	NotarySingerImage() string
	NotaryDBMigratorImage() string
	PortalImage() string
	RegistryImage() string
	RegistryControllerImage() string
}

func GetImage(registry *string, image string) string {
	var imageAddr string
	if registry == nil {
		imageAddr = image
	} else {
		imageAddr = fmt.Sprintf("%s/%s", *registry, image)
	}

	return imageAddr
}

func String(value string) *string {
	return &value
}
