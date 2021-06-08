package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	dst.Debug = src.Debug
	dst.Health = src.Health
	dst.JSON = src.JSON
	dst.LatencyInteger = src.LatencyInteger
}

func Convert_v1alpha3_ChartMuseumServerSpec_To_v1beta1_ChartMuseumServerSpec(src *v1alpha3.ChartMuseumServerSpec, dst *ChartMuseumServerSpec) {
	dst.TLS = src.TLS
	dst.CORSAllowOrigin = src.CORSAllowOrigin
	dst.MaxUploadSize = src.MaxUploadSize
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
}

func Convert_v1alpha3_ChartMuseumAuthSpec_To_v1beta1_ChartMuseumAuthSpec(src *v1alpha3.ChartMuseumAuthSpec, dst *ChartMuseumAuthSpec) {
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
	dst.Metrics = src.Metrics
	dst.Delete = src.Delete
	dst.API = src.API
	dst.ForceOverwrite = src.ForceOverwrite
	dst.StateFiles = src.StateFiles
}

func Convert_v1alpha3_ChartMuseumCacheSpec_To_v1beta1_ChartMuseumCacheSpec(src *v1alpha3.ChartMuseumCacheSpec, dst *ChartMuseumCacheSpec) {
	dst.Redis = src.Redis
}

func Convert_v1alpha3_ChartMuseumChartSpec_To_v1beta1_ChartMuseumChartSpec(src *v1alpha3.ChartMuseumChartSpec, dst *ChartMuseumChartSpec) {
	dst.PostFormFieldName = ChartMuseumPostFormFieldNameSpec{
		Chart:      src.PostFormFieldName.Chart,
		Provenance: src.PostFormFieldName.Provenance,
	}

	dst.URL = src.URL
	dst.AllowOverwrite = src.AllowOverwrite
	dst.SemanticVersioning2Only = src.SemanticVersioning2Only

	Convert_v1alpha3_ChartMuseumChartStorageSpec_To_v1beta1_ChartMuseumChartStorageSpec(&src.Storage, &dst.Storage)

	Convert_v1alpha3_ChartMuseumChartIndexSpec_To_v1beta1_ChartMuseumChartIndexSpec(&src.Index, &dst.Index)

	Convert_v1alpha3_ChartMuseumChartRepoSpec_To_v1beta1_ChartMuseumChartRepoSpec(&src.Repo, &dst.Repo)

}

func Convert_v1alpha3_ChartMuseumChartStorageSpec_To_v1beta1_ChartMuseumChartStorageSpec(src *v1alpha3.ChartMuseumChartStorageSpec, dst *ChartMuseumChartStorageSpec) {
	if src.MaxStorageObjects != nil {
		dst.MaxStorageObjects = new(int64)
		dst.MaxStorageObjects = src.MaxStorageObjects
	}

	Convert_v1alpha3_ChartMuseumChartStorageDriverSpec_To_v1beta1_ChartMuseumChartStorageDriverSpec(&src.ChartMuseumChartStorageDriverSpec, &dst.ChartMuseumChartStorageDriverSpec)

}

func Convert_v1alpha3_ChartMuseumChartStorageDriverSpec_To_v1beta1_ChartMuseumChartStorageDriverSpec(src *v1alpha3.ChartMuseumChartStorageDriverSpec, dst *ChartMuseumChartStorageDriverSpec) {
	if src.FileSystem != nil {
		dst.FileSystem = &ChartMuseumChartStorageDriverFilesystemSpec{
			VolumeSource: src.FileSystem.VolumeSource,
			Prefix:       src.FileSystem.Prefix,
		}
	}

	if src.Amazon != nil {
		dst.Amazon = &ChartMuseumChartStorageDriverAmazonSpec{
			Bucket:               src.Amazon.Bucket,
			Endpoint:             src.Amazon.Endpoint,
			Prefix:               src.Amazon.Prefix,
			Region:               src.Amazon.Region,
			ServerSideEncryption: src.Amazon.ServerSideEncryption,
			AccessKeyID:          src.Amazon.AccessKeyID,
			AccessSecretRef:      src.Amazon.AccessSecretRef,
		}
	}

	if src.OpenStack != nil {
		dst.OpenStack = &ChartMuseumChartStorageDriverOpenStackSpec{
			Container:         src.OpenStack.Container,
			Prefix:            src.OpenStack.Prefix,
			Region:            src.OpenStack.Region,
			AuthenticationURL: src.OpenStack.AuthenticationURL,
			Tenant:            src.OpenStack.Tenant,
			TenantID:          src.OpenStack.TenantID,
			Domain:            src.OpenStack.Domain,
			DomainID:          src.OpenStack.DomainID,
			Username:          src.OpenStack.Username,
			UserID:            src.OpenStack.UserID,
			PasswordRef:       src.OpenStack.PasswordRef,
		}
	}
}

func Convert_v1alpha3_ChartMuseumChartIndexSpec_To_v1beta1_ChartMuseumChartIndexSpec(src *v1alpha3.ChartMuseumChartIndexSpec, dst *ChartMuseumChartIndexSpec) {
	if src.ParallelLimit != nil {
		dst.ParallelLimit = new(int32)
		*dst.ParallelLimit = *src.ParallelLimit
	}

	if src.StorageTimestampTolerance != nil {
		dst.StorageTimestampTolerance = &metav1.Duration{
			Duration: src.StorageTimestampTolerance.Duration,
		}
	}
}

func Convert_v1alpha3_ChartMuseumChartRepoSpec_To_v1beta1_ChartMuseumChartRepoSpec(src *v1alpha3.ChartMuseumChartRepoSpec, dst *ChartMuseumChartRepoSpec) {
	if src.Depth != nil {
		dst.Depth = new(int32)
		*dst.Depth = *src.Depth
	}

	dst.DepthDynamic = src.DepthDynamic
}

func Convert_v1beta1_ChartMuseumLogSpec_To_v1alpha3_ChartMuseumLogSpec(src *ChartMuseumLogSpec, dst *v1alpha3.ChartMuseumLogSpec) {
	dst.Debug = src.Debug
	dst.Health = src.Health
	dst.JSON = src.JSON
	dst.LatencyInteger = src.LatencyInteger
}

func Convert_v1beta1_ChartMuseumServerSpec_To_v1alpha3_ChartMuseumServerSpec(src *ChartMuseumServerSpec, dst *v1alpha3.ChartMuseumServerSpec) {
	dst.TLS = src.TLS
	dst.CORSAllowOrigin = src.CORSAllowOrigin
	dst.MaxUploadSize = src.MaxUploadSize
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
}

func Convert_v1beta1_ChartMuseumAuthSpec_To_v1alpha3_ChartMuseumAuthSpec(src *ChartMuseumAuthSpec, dst *v1alpha3.ChartMuseumAuthSpec) {
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
	dst.Metrics = src.Metrics
	dst.Delete = src.Delete
	dst.API = src.API
	dst.ForceOverwrite = src.ForceOverwrite
	dst.StateFiles = src.StateFiles
}

func Convert_v1beta1_ChartMuseumCacheSpec_To_v1alpha3_ChartMuseumCacheSpec(src *ChartMuseumCacheSpec, dst *v1alpha3.ChartMuseumCacheSpec) {
	dst.Redis = src.Redis
}

func Convert_v1beta1_ChartMuseumChartSpec_To_v1alpha3_ChartMuseumChartSpec(src *ChartMuseumChartSpec, dst *v1alpha3.ChartMuseumChartSpec) {
	dst.PostFormFieldName = v1alpha3.ChartMuseumPostFormFieldNameSpec{
		Chart:      src.PostFormFieldName.Chart,
		Provenance: src.PostFormFieldName.Provenance,
	}

	dst.URL = src.URL
	dst.AllowOverwrite = src.AllowOverwrite
	dst.SemanticVersioning2Only = src.SemanticVersioning2Only

	Convert_v1beta1_ChartMuseumChartStorageSpec_To_v1alpha3_ChartMuseumChartStorageSpec(&src.Storage, &dst.Storage)

	Convert_v1beta1_ChartMuseumChartIndexSpec_To_v1alpha3_ChartMuseumChartIndexSpec(&src.Index, &dst.Index)

	Convert_v1beta1_ChartMuseumChartRepoSpec_To_v1alpha3_ChartMuseumChartRepoSpec(&src.Repo, &dst.Repo)

}

func Convert_v1beta1_ChartMuseumChartStorageSpec_To_v1alpha3_ChartMuseumChartStorageSpec(src *ChartMuseumChartStorageSpec, dst *v1alpha3.ChartMuseumChartStorageSpec) {
	if src.MaxStorageObjects != nil {
		dst.MaxStorageObjects = new(int64)
		dst.MaxStorageObjects = src.MaxStorageObjects
	}

	Convert_v1beta1_ChartMuseumChartStorageDriverSpec_To_v1alpha3_ChartMuseumChartStorageDriverSpec(&src.ChartMuseumChartStorageDriverSpec, &dst.ChartMuseumChartStorageDriverSpec)

}

func Convert_v1beta1_ChartMuseumChartStorageDriverSpec_To_v1alpha3_ChartMuseumChartStorageDriverSpec(src *ChartMuseumChartStorageDriverSpec, dst *v1alpha3.ChartMuseumChartStorageDriverSpec) {
	if src.FileSystem != nil {
		dst.FileSystem = &v1alpha3.ChartMuseumChartStorageDriverFilesystemSpec{
			VolumeSource: src.FileSystem.VolumeSource,
			Prefix:       src.FileSystem.Prefix,
		}
	}

	if src.Amazon != nil {
		dst.Amazon = &v1alpha3.ChartMuseumChartStorageDriverAmazonSpec{
			Bucket:               src.Amazon.Bucket,
			Endpoint:             src.Amazon.Endpoint,
			Prefix:               src.Amazon.Prefix,
			Region:               src.Amazon.Region,
			ServerSideEncryption: src.Amazon.ServerSideEncryption,
			AccessKeyID:          src.Amazon.AccessKeyID,
			AccessSecretRef:      src.Amazon.AccessSecretRef,
		}
	}

	if src.OpenStack != nil {
		dst.OpenStack = &v1alpha3.ChartMuseumChartStorageDriverOpenStackSpec{
			Container:         src.OpenStack.Container,
			Prefix:            src.OpenStack.Prefix,
			Region:            src.OpenStack.Region,
			AuthenticationURL: src.OpenStack.AuthenticationURL,
			Tenant:            src.OpenStack.Tenant,
			TenantID:          src.OpenStack.TenantID,
			Domain:            src.OpenStack.Domain,
			DomainID:          src.OpenStack.DomainID,
			Username:          src.OpenStack.Username,
			UserID:            src.OpenStack.UserID,
			PasswordRef:       src.OpenStack.PasswordRef,
		}
	}
}

func Convert_v1beta1_ChartMuseumChartIndexSpec_To_v1alpha3_ChartMuseumChartIndexSpec(src *ChartMuseumChartIndexSpec, dst *v1alpha3.ChartMuseumChartIndexSpec) {
	if src.ParallelLimit != nil {
		dst.ParallelLimit = new(int32)
		*dst.ParallelLimit = *src.ParallelLimit
	}

	if src.StorageTimestampTolerance != nil {
		dst.StorageTimestampTolerance = &metav1.Duration{
			Duration: src.StorageTimestampTolerance.Duration,
		}
	}
}

func Convert_v1beta1_ChartMuseumChartRepoSpec_To_v1alpha3_ChartMuseumChartRepoSpec(src *ChartMuseumChartRepoSpec, dst *v1alpha3.ChartMuseumChartRepoSpec) {
	if src.Depth != nil {
		dst.Depth = new(int32)
		*dst.Depth = *src.Depth
	}

	dst.DepthDynamic = src.DepthDynamic
}
