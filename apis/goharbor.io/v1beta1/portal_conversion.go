package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Portal) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Portal)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_PortalSpec_To_v1alpha3_PortalSpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *Portal) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Portal)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_PortalSpec_To_v1beta1_PortalSpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_PortalSpec_To_v1alpha3_PortalSpec(src *PortalSpec, dst *v1alpha3.PortalSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.PortalSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.MaxConnections = src.MaxConnections
	dst.TLS = src.TLS
}

func Convert_v1alpha3_PortalSpec_To_v1beta1_PortalSpec(src *v1alpha3.PortalSpec, dst *PortalSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &PortalSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.MaxConnections = src.MaxConnections
	dst.TLS = src.TLS
}
