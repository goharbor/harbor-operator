package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *ChartMuseum) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.ChartMuseum)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_ChartMuseumSpec_To_v1alpha3_ChartMuseumSpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *ChartMuseum) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.ChartMuseum)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_ChartMuseumSpec_To_v1beta1_ChartMuseumSpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_ChartMuseumSpec_To_v1alpha3_ChartMuseumSpec(src *ChartMuseumSpec, dst *v1alpha3.ChartMuseumSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ChartMuseumSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec

	dst.CertificateInjection = v1alpha3.CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}

	Convert_v1beta1_ChartMuseumLogSpec_To_v1alpha3_ChartMuseumLogSpec(&src.Log, &dst.Log)

	Convert_v1beta1_ChartMuseumServerSpec_To_v1alpha3_ChartMuseumServerSpec(&src.Server, &dst.Server)

	Convert_v1beta1_ChartMuseumAuthSpec_To_v1alpha3_ChartMuseumAuthSpec(&src.Authentication, &dst.Authentication)

	Convert_v1beta1_ChartMuseumDisableSpec_To_v1alpha3_ChartMuseumDisableSpec(&src.Disable, &dst.Disable)

	Convert_v1beta1_ChartMuseumCacheSpec_To_v1alpha3_ChartMuseumCacheSpec(&src.Cache, &dst.Cache)

	Convert_v1beta1_ChartMuseumChartSpec_To_v1alpha3_ChartMuseumChartSpec(&src.Chart, &dst.Chart)
}

func Convert_v1alpha3_ChartMuseumSpec_To_v1beta1_ChartMuseumSpec(src *v1alpha3.ChartMuseumSpec, dst *ChartMuseumSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ChartMuseumSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec

	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}

	Convert_v1alpha3_ChartMuseumLogSpec_To_v1beta1_ChartMuseumLogSpec(&src.Log, &dst.Log)

	Convert_v1alpha3_ChartMuseumServerSpec_To_v1beta1_ChartMuseumServerSpec(&src.Server, &dst.Server)

	Convert_v1alpha3_ChartMuseumAuthSpec_To_v1beta1_ChartMuseumAuthSpec(&src.Authentication, &dst.Authentication)

	Convert_v1alpha3_ChartMuseumDisableSpec_To_v1beta1_ChartMuseumDisableSpec(&src.Disable, &dst.Disable)

	Convert_v1alpha3_ChartMuseumCacheSpec_To_v1beta1_ChartMuseumCacheSpec(&src.Cache, &dst.Cache)

	Convert_v1alpha3_ChartMuseumChartSpec_To_v1beta1_ChartMuseumChartSpec(&src.Chart, &dst.Chart)
}

func Convert_v1alpha3_ChartMuseumLogSpec_To_v1beta1_ChartMuseumLogSpec(src *v1alpha3.ChartMuseumLogSpec, dst *ChartMuseumLogSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ChartMuseumLogSpec{}
	}

	dst.Debug = src.Debug
	dst.Health = src.Health
	dst.JSON = src.JSON
	dst.LatencyInteger = src.LatencyInteger
}

func Convert_v1alpha3_ChartMuseumServerSpec_To_v1beta1_ChartMuseumServerSpec(src *v1alpha3.ChartMuseumServerSpec, dst *ChartMuseumServerSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ChartMuseumServerSpec{}
	}

	dst.TLS = src.TLS
	dst.CORSAllowOrigin = src.CORSAllowOrigin
	dst.MaxUploadSize = src.MaxUploadSize
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
}

func Convert_v1alpha3_ChartMuseumAuthSpec_To_v1beta1_ChartMuseumAuthSpec(src *v1alpha3.ChartMuseumAuthSpec, dst *ChartMuseumAuthSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ChartMuseumAuthSpec{}
	}

	dst.AnonymousGet = src.AnonymousGet
	dst.BasicAuthRef = src.BasicAuthRef

	if src.Bearer != nil {
		dst.Bearer = &ChartMuseumAuthBearerSpec{
			CertificateRef: src.Bearer.CertificateRef,
			Realm:          src.Bearer.Realm,
			Service:        src.Bearer.Service,
		}
	}
}

func Convert_v1alpha3_ChartMuseumDisableSpec_To_v1beta1_ChartMuseumDisableSpec(src *v1alpha3.ChartMuseumDisableSpec, dst *ChartMuseumDisableSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ChartMuseumDisableSpec{}
	}

	dst.Metrics = src.Metrics
	dst.Delete = src.Delete
	dst.API = src.API
	dst.ForceOverwrite = src.ForceOverwrite
	dst.StateFiles = src.StateFiles
}

func Convert_v1alpha3_ChartMuseumCacheSpec_To_v1beta1_ChartMuseumCacheSpec(src *v1alpha3.ChartMuseumCacheSpec, dst *ChartMuseumCacheSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ChartMuseumCacheSpec{}
	}

	dst.Redis = src.Redis
}

func Convert_v1alpha3_ChartMuseumChartSpec_To_v1beta1_ChartMuseumChartSpec(src *v1alpha3.ChartMuseumChartSpec, dst *ChartMuseumChartSpec) {

}

func Convert_v1beta1_ChartMuseumLogSpec_To_v1alpha3_ChartMuseumLogSpec(src *ChartMuseumLogSpec, dst *v1alpha3.ChartMuseumLogSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ChartMuseumLogSpec{}
	}

	dst.Debug = src.Debug
	dst.Health = src.Health
	dst.JSON = src.JSON
	dst.LatencyInteger = src.LatencyInteger
}

func Convert_v1beta1_ChartMuseumServerSpec_To_v1alpha3_ChartMuseumServerSpec(src *ChartMuseumServerSpec, dst *v1alpha3.ChartMuseumServerSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ChartMuseumServerSpec{}
	}

	dst.TLS = src.TLS
	dst.CORSAllowOrigin = src.CORSAllowOrigin
	dst.MaxUploadSize = src.MaxUploadSize
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
}

func Convert_v1beta1_ChartMuseumAuthSpec_To_v1alpha3_ChartMuseumAuthSpec(src *ChartMuseumAuthSpec, dst *v1alpha3.ChartMuseumAuthSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ChartMuseumAuthSpec{}
	}

	dst.AnonymousGet = src.AnonymousGet
	dst.BasicAuthRef = src.BasicAuthRef

	if src.Bearer != nil {
		dst.Bearer = &v1alpha3.ChartMuseumAuthBearerSpec{
			CertificateRef: src.Bearer.CertificateRef,
			Realm:          src.Bearer.Realm,
			Service:        src.Bearer.Service,
		}
	}
}

func Convert_v1beta1_ChartMuseumDisableSpec_To_v1alpha3_ChartMuseumDisableSpec(src *ChartMuseumDisableSpec, dst *v1alpha3.ChartMuseumDisableSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ChartMuseumDisableSpec{}
	}

	dst.Metrics = src.Metrics
	dst.Delete = src.Delete
	dst.API = src.API
	dst.ForceOverwrite = src.ForceOverwrite
	dst.StateFiles = src.StateFiles
}

func Convert_v1beta1_ChartMuseumCacheSpec_To_v1alpha3_ChartMuseumCacheSpec(src *ChartMuseumCacheSpec, dst *v1alpha3.ChartMuseumCacheSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ChartMuseumCacheSpec{}
	}

	dst.Redis = src.Redis
}

func Convert_v1beta1_ChartMuseumChartSpec_To_v1alpha3_ChartMuseumChartSpec(src *ChartMuseumChartSpec, dst *v1alpha3.ChartMuseumChartSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ChartMuseumChartSpec{}
	}

}
