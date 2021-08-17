package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	NetworkAnnotationName = "goharbor.io/network"
)

type Network struct {
	// +kubebuilder:validation:Optional
	IPFamilies []corev1.IPFamily `json:"ipFamilies,omitempty"`
}

func (network *Network) isIPFamilyEnabled(ipFamily corev1.IPFamily) bool {
	if network == nil {
		return true
	}

	for _, ipf := range network.IPFamilies {
		if ipf == ipFamily {
			return true
		}
	}

	return false
}

func (network *Network) IsIPv4Enabled() bool {
	return network.isIPFamilyEnabled(corev1.IPv4Protocol)
}

func (network *Network) IsIPv6Enabled() bool {
	return network.isIPFamilyEnabled(corev1.IPv6Protocol)
}

func (network *Network) Validate(rootPath *field.Path) *field.Error {
	if network == nil {
		return nil
	}

	if rootPath == nil {
		rootPath = field.NewPath("spec").Child("network")
	}

	for i, ipf := range network.IPFamilies {
		if ipf == corev1.IPv4Protocol || ipf == corev1.IPv6Protocol {
			continue
		}

		return field.Invalid(rootPath.Child("ipFamilies").Index(i), ipf, `valid value is "IPv4", "IPv6"`)
	}

	return nil
}
