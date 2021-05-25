package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Exporter) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Exporter)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_ExporterSpec_To_v1alpha3_ExporterSpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *Exporter) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Exporter)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_ExporterSpec_To_v1beta1_ExporterSpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_ExporterSpec_To_v1alpha3_ExporterSpec(src *ExporterSpec, dst *v1alpha3.ExporterSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ExporterSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.TLS = src.TLS
	dst.Port = src.Port
	dst.Path = src.Path

	Convert_v1beta1_ExporterLogSpec_To_v1alpha3_ExporterLogSpec(&src.Log, &dst.Log)

	Convert_v1beta1_ExporterCacheSpec_To_v1alpha3_ExporterCacheSpec(&src.Cache, &dst.Cache)

	Convert_v1beta1_ExporterCoreSpec_To_v1alpha3_ExporterCoreSpec(&src.Core, &dst.Core)

	Convert_v1beta1_ExporterDatabaseSpec_To_v1alpha3_ExporterDatabaseSpec(&src.Database, &dst.Database)
}

func Convert_v1beta1_ExporterLogSpec_To_v1alpha3_ExporterLogSpec(src *ExporterLogSpec, dst *v1alpha3.ExporterLogSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ExporterLogSpec{}
	}

	dst.Level = src.Level
}

func Convert_v1beta1_ExporterCoreSpec_To_v1alpha3_ExporterCoreSpec(src *ExporterCoreSpec, dst *v1alpha3.ExporterCoreSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ExporterCoreSpec{}
	}

	dst.URL = src.URL
}

func Convert_v1beta1_ExporterDatabaseSpec_To_v1alpha3_ExporterDatabaseSpec(src *ExporterDatabaseSpec, dst *v1alpha3.ExporterDatabaseSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ExporterDatabaseSpec{}
	}

	dst.PostgresConnectionWithParameters = src.PostgresConnectionWithParameters
	dst.MaxOpenConnections = src.MaxOpenConnections
	dst.MaxIdleConnections = src.MaxIdleConnections
	dst.EncryptionKeyRef = src.EncryptionKeyRef
}

func Convert_v1beta1_ExporterCacheSpec_To_v1alpha3_ExporterCacheSpec(src *ExporterCacheSpec, dst *v1alpha3.ExporterCacheSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ExporterCacheSpec{}
	}

	dst.CleanInterval = src.CleanInterval
	dst.Duration = src.Duration
}

func Convert_v1alpha3_ExporterSpec_To_v1beta1_ExporterSpec(src *v1alpha3.ExporterSpec, dst *ExporterSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ExporterSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.TLS = src.TLS
	dst.Port = src.Port
	dst.Path = src.Path

	Convert_v1alpha3_ExporterLogSpec_To_v1beta1_ExporterLogSpec(&src.Log, &dst.Log)

	Convert_v1alpha3_ExporterCacheSpec_To_v1beta1_ExporterCacheSpec(&src.Cache, &dst.Cache)

	Convert_v1alpha3_ExporterCoreSpec_To_v1beta1_ExporterCoreSpec(&src.Core, &dst.Core)

	Convert_v1alpha3_ExporterDatabaseSpec_To_v1beta1_ExporterDatabaseSpec(&src.Database, &dst.Database)
}

func Convert_v1alpha3_ExporterLogSpec_To_v1beta1_ExporterLogSpec(src *v1alpha3.ExporterLogSpec, dst *ExporterLogSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ExporterLogSpec{}
	}

	dst.Level = src.Level
}

func Convert_v1alpha3_ExporterCoreSpec_To_v1beta1_ExporterCoreSpec(src *v1alpha3.ExporterCoreSpec, dst *ExporterCoreSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ExporterCoreSpec{}
	}

	dst.URL = src.URL
}

func Convert_v1alpha3_ExporterDatabaseSpec_To_v1beta1_ExporterDatabaseSpec(src *v1alpha3.ExporterDatabaseSpec, dst *ExporterDatabaseSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ExporterDatabaseSpec{}
	}

	dst.PostgresConnectionWithParameters = src.PostgresConnectionWithParameters
	dst.MaxOpenConnections = src.MaxOpenConnections
	dst.MaxIdleConnections = src.MaxIdleConnections
	dst.EncryptionKeyRef = src.EncryptionKeyRef
}

func Convert_v1alpha3_ExporterCacheSpec_To_v1beta1_ExporterCacheSpec(src *v1alpha3.ExporterCacheSpec, dst *ExporterCacheSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ExporterCacheSpec{}
	}

	dst.CleanInterval = src.CleanInterval
	dst.Duration = src.Duration
}
