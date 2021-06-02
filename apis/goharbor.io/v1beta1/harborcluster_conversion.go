package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &HarborCluster{}

func (src *HarborCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.HarborCluster)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.HarborSpec.LogLevel = src.Spec.HarborSpec.LogLevel
	dst.Spec.HarborSpec.Version = src.Spec.HarborSpec.Version
	dst.Spec.HarborSpec.Core = v1alpha3.CoreComponentSpec{
		ComponentSpec:        harbormetav1.ComponentSpec{},
		CertificateInjection: v1alpha3.CertificateInjection{},
		TokenIssuer:          cmmeta.ObjectReference{},
		Metrics:              nil,
	}

	return nil
}

func (dst *HarborCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.HarborCluster)

	dst.ObjectMeta = src.ObjectMeta

	Convert_v1alph3_HarborClusterSpec_To_v1beta1_HarborClusterSpec(&src.Spec, &dst.Spec)

	Convert_v1alph3_HarborClusterStatus_To_v1beta1_HarborClusterStatus(&src.Status, &dst.Status)

	return nil

}

func Convert_v1alph3_HarborClusterSpec_To_v1beta1_HarborClusterSpec(src *v1alpha3.HarborClusterSpec, dst *HarborClusterSpec) {
	if src.InClusterCache != nil {
		Convert_v1alph3_Cache_To_v1beta1_Cache(src.InClusterCache, dst.Cache)
	}

	if src.InClusterStorage != nil {
		Convert_v1alph3_Storage_To_v1beta1_Storage(src.InClusterStorage, dst.Storage)
	}

	if src.InClusterDatabase != nil {
		Convert_v1alph3_Database_To_v1beta1_Database(src.InClusterDatabase, dst.Database)
	}

	Convert_v1alph3_HarborSpec_To_v1beta1_HarborSpec(&src.HarborSpec, &dst.HarborSpec)
}

func Convert_v1alph3_HarborSpec_To_v1beta1_HarborSpec(src *v1alpha3.HarborSpec, dst *HarborSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &HarborSpec{}
	}

	dst.ExternalURL = src.ExternalURL
	dst.InternalTLS = HarborInternalTLSSpec{
		Enabled: src.InternalTLS.Enabled,
	}
	//dst.ImageChartStorage
	dst.LogLevel = src.LogLevel
	dst.HarborAdminPasswordRef = src.HarborAdminPasswordRef
	dst.UpdateStrategyType = src.UpdateStrategyType
	dst.Version = src.Version

	if src.Proxy != nil {
		dst.Proxy = &HarborProxySpec{
			ProxySpec:  src.Proxy.ProxySpec,
			Components: src.Proxy.Components,
		}
	}

	Convert_v1alph3_ImageSourceSpec_To_v1beta1_ImageSourceSpec(src.ImageSource, dst.ImageSource)

	Convert_v1alph3_HarborExposeSpec_To_v1beta1_HarborExposeSpec(&src.Expose, &dst.Expose)

	Convert_v1alph3_HarborComponentSpec_To_v1beta1_HarborComponentSpec(&src.HarborComponentsSpec, &dst.HarborComponentsSpec)

}

func Convert_v1alph3_HarborComponentSpec_To_v1beta1_HarborComponentSpec(src *v1alpha3.HarborComponentsSpec, dst *HarborComponentsSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &HarborComponentsSpec{}
	}

	Convert_v1alph3_CoreComponentSpec_To_v1beta1_CoreComponentSpec(&src.Core, &dst.Core)

	Convert_v1alph3_RegistryComponentSpec_To_v1beta1_RegistryComponentSpec(&src.Registry, &dst.Registry)

	Convert_v1alph3_JobServiceComponentSpec_To_v1beta1_JobServiceComponentSpec(&src.JobService, &dst.JobService)

	Convert_v1alph3_ChartMuseumComponentSpec_To_v1beta1_ChartMuseumComponentSpec(src.ChartMuseum, dst.ChartMuseum)

	Convert_v1alph3_NotaryComponentSpec_To_v1beta1_NotaryComponentSpec(src.Notary, dst.Notary)

	Convert_v1alph3_TrivyComponentSpec_To_v1beta1_TrivyComponentSpec(src.Trivy, dst.Trivy)

	Convert_v1alph3_ExporterComponentSpec_To_v1beta1_ExporterComponentSpec(src.Exporter, dst.Exporter)
}

func Convert_v1alph3_CoreComponentSpec_To_v1beta1_CoreComponentSpec(src *v1alpha3.CoreComponentSpec, dst *CoreComponentSpec) {
	if dst == nil {
		dst = &CoreComponentSpec{}
	}

	if src == nil {
		return
	}

	dst.CertificateInjection = CertificateInjection{src.CertificateInjection.CertificateRefs}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.TokenIssuer = src.TokenIssuer

}

func Convert_v1alph3_RegistryComponentSpec_To_v1beta1_RegistryComponentSpec(src *v1alpha3.RegistryComponentSpec, dst *RegistryComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &RegistryComponentSpec{}
	}

	dst.CertificateInjection = CertificateInjection{src.CertificateInjection.CertificateRefs}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.RelativeURLs = src.RelativeURLs

	if src.StorageMiddlewares != nil {
		dst.StorageMiddlewares = make([]RegistryMiddlewareSpec, 0, len(src.StorageMiddlewares))
		for _, value := range src.StorageMiddlewares {
			dst.StorageMiddlewares = append(dst.StorageMiddlewares, RegistryMiddlewareSpec{
				Name:       value.Name,
				OptionsRef: value.OptionsRef,
			})
		}
	}
}

func Convert_v1alph3_JobServiceComponentSpec_To_v1beta1_JobServiceComponentSpec(src *v1alpha3.JobServiceComponentSpec, dst *JobServiceComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &JobServiceComponentSpec{}
	}

	dst.WorkerCount = src.WorkerCount
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{src.CertificateInjection.CertificateRefs}
}

func Convert_v1alph3_ChartMuseumComponentSpec_To_v1beta1_ChartMuseumComponentSpec(src *v1alpha3.ChartMuseumComponentSpec, dst *ChartMuseumComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ChartMuseumComponentSpec{}
	}

	dst.AbsoluteURL = src.AbsoluteURL
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{src.CertificateInjection.CertificateRefs}
}

func Convert_v1alph3_ExporterComponentSpec_To_v1beta1_ExporterComponentSpec(src *v1alpha3.ExporterComponentSpec, dst *ExporterComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &ExporterComponentSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.Port = src.Port
	dst.Path = src.Path

	Convert_v1alph3_HarborExporterCacheSpec_To_v1beta1_HarborExporterCacheSpec(&src.Cache, &dst.Cache)
}

func Convert_v1alph3_HarborExporterCacheSpec_To_v1beta1_HarborExporterCacheSpec(src *v1alpha3.HarborExporterCacheSpec, dst *HarborExporterCacheSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &HarborExporterCacheSpec{}
	}

	dst.Duration = src.Duration
	dst.CleanInterval = src.CleanInterval
}

func Convert_v1alph3_TrivyComponentSpec_To_v1beta1_TrivyComponentSpec(src *v1alpha3.TrivyComponentSpec, dst *TrivyComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &TrivyComponentSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{src.CertificateInjection.CertificateRefs}
	dst.GithubTokenRef = src.GithubTokenRef
	dst.SkipUpdate = src.SkipUpdate

	Convert_v1alph3_HarborStorageTrivyStorageSpec_To_v1beta1_HarborStorageTrivyStorageSpec(&src.Storage, &dst.Storage)

}

func Convert_v1alph3_HarborStorageTrivyStorageSpec_To_v1beta1_HarborStorageTrivyStorageSpec(src *v1alpha3.HarborStorageTrivyStorageSpec, dst *HarborStorageTrivyStorageSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &HarborStorageTrivyStorageSpec{}
	}

	if src.CachePersistentVolume != nil {
		dst.CachePersistentVolume = &HarborStoragePersistentVolumeSpec{}
		dst.CachePersistentVolume.Prefix = src.CachePersistentVolume.Prefix
		dst.CachePersistentVolume.PersistentVolumeClaimVolumeSource = src.CachePersistentVolume.PersistentVolumeClaimVolumeSource
	}

	if dst.ReportsPersistentVolume != nil {
		dst.ReportsPersistentVolume = &HarborStoragePersistentVolumeSpec{}
		dst.CachePersistentVolume.Prefix = src.CachePersistentVolume.Prefix
		dst.CachePersistentVolume.PersistentVolumeClaimVolumeSource = src.CachePersistentVolume.PersistentVolumeClaimVolumeSource
	}
}

func Convert_v1alph3_NotaryComponentSpec_To_v1beta1_NotaryComponentSpec(src *v1alpha3.NotaryComponentSpec, dst *NotaryComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &NotaryComponentSpec{}
	}

	dst.Server = src.Server
	dst.Signer = src.Signer
	dst.MigrationEnabled = src.MigrationEnabled
}

func Convert_v1alph3_ImageSourceSpec_To_v1beta1_ImageSourceSpec(src *v1alpha3.ImageSourceSpec, dst *ImageSourceSpec) {
	if dst == nil {
		dst = &ImageSourceSpec{}
	}

	if src == nil {
		return
	}

	dst.ImagePullPolicy = src.ImagePullPolicy
	dst.ImagePullSecrets = src.ImagePullSecrets
	dst.Repository = src.Repository
	dst.TagSuffix = src.TagSuffix
}

func Convert_v1alph3_Cache_To_v1beta1_Cache(src *v1alpha3.Cache, dst *Cache) {
	dst.Kind = src.Kind

	if src.RedisSpec != nil {
		Convert_v1alph3_RedisSpec_To_v1beta1_CacheSpec(src.RedisSpec, dst.Spec)
	}
}

func Convert_v1alph3_RedisSpec_To_v1beta1_CacheSpec(src *v1alpha3.RedisSpec, dst *CacheSpec) {
	dst = &CacheSpec{
		RedisFailover: &RedisFailoverSpec{},
	}

	if src.Server != nil {
		dst.RedisFailover.Server = &RedisServer{
			Replicas: src.Server.Replicas,
			Resources: corev1.ResourceRequirements{
				Limits:   src.Server.Resources.Limits,
				Requests: src.Server.Resources.Requests,
			},
			StorageClassName: src.Server.StorageClassName,
			Storage:          src.Server.Storage,
		}
	}

	if src.Sentinel != nil {
		dst.RedisFailover.Sentinel = &RedisSentinel{
			Replicas: src.Sentinel.Replicas,
		}
	}

	// TODO
	dst.RedisFailover.OperatorVersion = ""
	dst.RedisFailover.ImageSpec = src.ImageSpec
}

func Convert_v1alph3_Storage_To_v1beta1_Storage(src *v1alpha3.Storage, dst *Storage) {
	dst.Kind = src.Kind
	dst.Spec = StorageSpec{}

	if src.MinIOSpec != nil {
		Convert_v1alph3_MinIOSpec_to_v1beta1_MinIOSpec(src.MinIOSpec, dst.Spec.MinIO)
	}
}

func Convert_v1alph3_MinIOSpec_to_v1beta1_MinIOSpec(src *v1alpha3.MinIOSpec, dst *MinIOSpec) {
	dst = &MinIOSpec{}

	dst.OperatorVersion = ""

	dst.SecretRef = src.SecretRef
	dst.VolumeClaimTemplate = src.VolumeClaimTemplate
	dst.VolumesPerServer = src.VolumesPerServer

	dst.ImageSpec = src.ImageSpec

	dst.Replicas = src.Replicas
	dst.Resources = src.Resources

	dst.Redirect.Enable = src.Redirect.Enable
	if src.Redirect.Expose != nil {
		Convert_v1alph3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src.Redirect.Expose, dst.Redirect.Expose)
	}
}

func Convert_v1alph3_HarborExposeSpec_To_v1beta1_HarborExposeSpec(src *v1alpha3.HarborExposeSpec, dst *HarborExposeSpec) {
	if dst == nil {
		dst = &HarborExposeSpec{}
	}

	Convert_v1alph3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(&src.Core, &dst.Core)

	Convert_v1alph3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src.Notary, dst.Notary)
}

func Convert_v1alph3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src *v1alpha3.HarborExposeComponentSpec, dst *HarborExposeComponentSpec) {
	if dst == nil {
		dst = &HarborExposeComponentSpec{}
	}

	if src == nil {
		return
	}

	if src.Ingress != nil {
		Convert_v1alph3_HarborExposeIngressSpec_To_v1beta1_HarborExposeIngressSpec(src.Ingress, dst.Ingress)
	}

	if src.TLS != nil {
		dst.TLS = src.TLS
	}
}

func Convert_v1alph3_HarborExposeIngressSpec_To_v1beta1_HarborExposeIngressSpec(src *v1alpha3.HarborExposeIngressSpec, dst *HarborExposeIngressSpec) {
	if dst == nil {
		dst = &HarborExposeIngressSpec{}
	}

	if src == nil {
		return
	}

	dst.Host = src.Host
	dst.Controller = src.Controller
	dst.Annotations = src.Annotations
}

func Convert_v1alph3_Database_To_v1beta1_Database(src *v1alpha3.Database, dst *Database) {
	dst = &Database{}

	dst.Kind = src.Kind

	if src.PostgresSQLSpec != nil {
		dst.Spec.ZlandoPostgreSQL = &ZlandoPostgreSQLSpec{}
		Convert_v1alph3_PostgresSQLSpec_To_v1beta1_ZlandoPostgresSQLSpec(src.PostgresSQLSpec, dst.Spec.ZlandoPostgreSQL)
	}
}

func Convert_v1alph3_PostgresSQLSpec_To_v1beta1_ZlandoPostgresSQLSpec(src *v1alpha3.PostgresSQLSpec, dst *ZlandoPostgreSQLSpec) {
	dst.OperatorVersion = ""
	dst.Storage = src.Storage
	dst.Resources = src.Resources
	dst.Replicas = src.Replicas

	dst.ImageSpec = src.ImageSpec
	dst.StorageClassName = src.StorageClassName
}

func Convert_v1alph3_HarborClusterStatus_To_v1beta1_HarborClusterStatus(src *v1alpha3.HarborClusterStatus, dst *HarborClusterStatus) {
	dst.Operator = src.Operator
	dst.Status = ClusterStatus(src.Status)
	dst.ObservedGeneration = src.ObservedGeneration
	dst.Revision = src.Revision

	dst.Conditions = func() []HarborClusterCondition {
		conditions := make([]HarborClusterCondition, 0, len(src.Conditions))
		for _, cond := range src.Conditions {
			conditions = append(conditions, HarborClusterCondition{
				Type:               HarborClusterConditionType(cond.Type),
				Status:             cond.Status,
				LastTransitionTime: cond.LastTransitionTime,
				Reason:             cond.Reason,
				Message:            cond.Message,
			})
		}
		return conditions
	}()
}

//-----------------------------------------------------------

func Convert_v1beta1_HarborClusterSpec_To_v1alpha3_HarborClusterSpec(src *HarborClusterSpec, dst *v1alpha3.HarborClusterSpec) {
	if src.Cache != nil {
		Convert_v1beta1_Cache_To_v1alpha3_Cache(src.Cache, dst.InClusterCache)
	}

	if src.Storage != nil {
		Convert_v1beta1_Storage_To_v1alpha3_Storage(src.Storage, dst.InClusterStorage)
	}

	if src.Database != nil {
		Convert_v1beta1_Database_To_v1alpha3_Database(src.Database, dst.InClusterDatabase)
	}

	Convert_v1beta1_HarborSpec_To_v1alpha3_HarborSpec(&src.HarborSpec, &dst.HarborSpec)
}

func Convert_v1beta1_HarborSpec_To_v1alpha3_HarborSpec(src *HarborSpec, dst *v1alpha3.HarborSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.HarborSpec{}
	}

	dst.ExternalURL = src.ExternalURL
	dst.InternalTLS = v1alpha3.HarborInternalTLSSpec{
		Enabled: src.InternalTLS.Enabled,
	}
	//dst.ImageChartStorage
	dst.LogLevel = src.LogLevel
	dst.HarborAdminPasswordRef = src.HarborAdminPasswordRef
	dst.UpdateStrategyType = src.UpdateStrategyType
	dst.Version = src.Version

	if src.Proxy != nil {
		dst.Proxy = &v1alpha3.HarborProxySpec{
			ProxySpec:  src.Proxy.ProxySpec,
			Components: src.Proxy.Components,
		}
	}

	Convert_v1beta1_ImageSourceSpec_To_v1alpha3_ImageSourceSpec(src.ImageSource, dst.ImageSource)

	Convert_v1beta1_HarborExposeSpec_To_v1alpha3_HarborExposeSpec(&src.Expose, &dst.Expose)

	Convert_v1beta1_HarborComponentSpec_To_v1alpha3_HarborComponentSpec(&src.HarborComponentsSpec, &dst.HarborComponentsSpec)

}

func Convert_v1beta1_HarborComponentSpec_To_v1alpha3_HarborComponentSpec(src *HarborComponentsSpec, dst *v1alpha3.HarborComponentsSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.HarborComponentsSpec{}
	}

	Convert_v1beta1_CoreComponentSpec_To_v1alpha3_CoreComponentSpec(&src.Core, &dst.Core)

	Convert_v1beta1_RegistryComponentSpec_To_v1alpha3_RegistryComponentSpec(&src.Registry, &dst.Registry)

	Convert_v1beta1_JobServiceComponentSpec_To_v1alpha3_JobServiceComponentSpec(&src.JobService, &dst.JobService)

	Convert_v1beta1_ChartMuseumComponentSpec_To_v1alpha3_ChartMuseumComponentSpec(src.ChartMuseum, dst.ChartMuseum)

	Convert_v1beta1_NotaryComponentSpec_To_v1alpha3_NotaryComponentSpec(src.Notary, dst.Notary)

	Convert_v1beta1_TrivyComponentSpec_To_v1alpha3_TrivyComponentSpec(src.Trivy, dst.Trivy)

	Convert_v1beta1_ExporterComponentSpec_To_v1alpha3_ExporterComponentSpec(src.Exporter, dst.Exporter)
}

func Convert_v1beta1_CoreComponentSpec_To_v1alpha3_CoreComponentSpec(src *CoreComponentSpec, dst *v1alpha3.CoreComponentSpec) {
	if dst == nil {
		dst = &v1alpha3.CoreComponentSpec{}
	}

	if src == nil {
		return
	}

	dst.CertificateInjection = v1alpha3.CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.TokenIssuer = src.TokenIssuer

}

func Convert_v1beta1_RegistryComponentSpec_To_v1alpha3_RegistryComponentSpec(src *RegistryComponentSpec, dst *v1alpha3.RegistryComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.RegistryComponentSpec{}
	}

	dst.CertificateInjection = v1alpha3.CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.RelativeURLs = src.RelativeURLs

	if src.StorageMiddlewares != nil {
		dst.StorageMiddlewares = make([]v1alpha3.RegistryMiddlewareSpec, 0, len(src.StorageMiddlewares))
		for _, value := range src.StorageMiddlewares {
			dst.StorageMiddlewares = append(dst.StorageMiddlewares, v1alpha3.RegistryMiddlewareSpec{
				Name:       value.Name,
				OptionsRef: value.OptionsRef,
			})
		}
	}
}

func Convert_v1beta1_JobServiceComponentSpec_To_v1alpha3_JobServiceComponentSpec(src *JobServiceComponentSpec, dst *v1alpha3.JobServiceComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.JobServiceComponentSpec{}
	}

	dst.WorkerCount = src.WorkerCount
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1alpha3.CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
}

func Convert_v1beta1_ChartMuseumComponentSpec_To_v1alpha3_ChartMuseumComponentSpec(src *ChartMuseumComponentSpec, dst *v1alpha3.ChartMuseumComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ChartMuseumComponentSpec{}
	}

	dst.AbsoluteURL = src.AbsoluteURL
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1alpha3.CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
}

func Convert_v1beta1_ExporterComponentSpec_To_v1alpha3_ExporterComponentSpec(src *ExporterComponentSpec, dst *v1alpha3.ExporterComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ExporterComponentSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.Port = src.Port
	dst.Path = src.Path

	Convert_v1beta1_HarborExporterCacheSpec_To_v1alpha3_HarborExporterCacheSpec(&src.Cache, &dst.Cache)
}

func Convert_v1beta1_HarborExporterCacheSpec_To_v1alpha3_HarborExporterCacheSpec(src *HarborExporterCacheSpec, dst *v1alpha3.HarborExporterCacheSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.HarborExporterCacheSpec{}
	}

	dst.Duration = src.Duration
	dst.CleanInterval = src.CleanInterval
}

func Convert_v1beta1_TrivyComponentSpec_To_v1alpha3_TrivyComponentSpec(src *TrivyComponentSpec, dst *v1alpha3.TrivyComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.TrivyComponentSpec{}
	}

	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1alpha3.CertificateInjection{src.CertificateInjection.CertificateRefs}
	dst.GithubTokenRef = src.GithubTokenRef
	dst.SkipUpdate = src.SkipUpdate

	Convert_v1beta1_HarborStorageTrivyStorageSpec_To_v1alpha3_HarborStorageTrivyStorageSpec(&src.Storage, &dst.Storage)

}

func Convert_v1beta1_HarborStorageTrivyStorageSpec_To_v1alpha3_HarborStorageTrivyStorageSpec(src *HarborStorageTrivyStorageSpec, dst *v1alpha3.HarborStorageTrivyStorageSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.HarborStorageTrivyStorageSpec{}
	}

	if src.CachePersistentVolume != nil {
		dst.CachePersistentVolume = &v1alpha3.HarborStoragePersistentVolumeSpec{}
		dst.CachePersistentVolume.Prefix = src.CachePersistentVolume.Prefix
		dst.CachePersistentVolume.PersistentVolumeClaimVolumeSource = src.CachePersistentVolume.PersistentVolumeClaimVolumeSource
	}

	if dst.ReportsPersistentVolume != nil {
		dst.ReportsPersistentVolume = &v1alpha3.HarborStoragePersistentVolumeSpec{}
		dst.CachePersistentVolume.Prefix = src.CachePersistentVolume.Prefix
		dst.CachePersistentVolume.PersistentVolumeClaimVolumeSource = src.CachePersistentVolume.PersistentVolumeClaimVolumeSource
	}
}

func Convert_v1beta1_NotaryComponentSpec_To_v1alpha3_NotaryComponentSpec(src *NotaryComponentSpec, dst *v1alpha3.NotaryComponentSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.NotaryComponentSpec{}
	}

	dst.Server = src.Server
	dst.Signer = src.Signer
	dst.MigrationEnabled = src.MigrationEnabled
}

func Convert_v1beta1_ImageSourceSpec_To_v1alpha3_ImageSourceSpec(src *ImageSourceSpec, dst *v1alpha3.ImageSourceSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.ImageSourceSpec{}
	}

	dst.ImagePullPolicy = src.ImagePullPolicy
	dst.ImagePullSecrets = src.ImagePullSecrets
	dst.Repository = src.Repository
	dst.TagSuffix = src.TagSuffix
}

func Convert_v1beta1_Cache_To_v1alpha3_Cache(src *Cache, dst *v1alpha3.Cache) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.Cache{
			Kind: src.Kind,
		}
	}

	Convert_v1beta1_RedisSpec_To_v1alpha3_CacheSpec(src.Spec, dst.RedisSpec)
}

func Convert_v1beta1_RedisSpec_To_v1alpha3_CacheSpec(src *CacheSpec, dst *v1alpha3.RedisSpec) {
	dst = &v1alpha3.RedisSpec{}

	if src.RedisFailover != nil {

		if src.RedisFailover.Server != nil {
			dst.Server = &v1alpha3.RedisServer{
				Replicas:         src.RedisFailover.Server.Replicas,
				Resources:        src.RedisFailover.Server.Resources,
				StorageClassName: src.RedisFailover.Server.StorageClassName,
				Storage:          src.RedisFailover.Server.Storage,
			}
		}

		if src.RedisFailover.Sentinel != nil {
			dst.Sentinel = &v1alpha3.RedisSentinel{
				Replicas: src.RedisFailover.Sentinel.Replicas,
			}
		}

		dst.ImageSpec = src.RedisFailover.ImageSpec

	}
}

func Convert_v1beta1_Storage_To_v1alpha3_Storage(src *Storage, dst *v1alpha3.Storage) {
	dst.Kind = src.Kind

	if src.Spec.MinIO != nil {
		Convert_v1beta1_MinIOSpec_To_v1alpha3_MinIOSpec(src.Spec.MinIO, dst.MinIOSpec)
	}
}

func Convert_v1beta1_MinIOSpec_To_v1alpha3_MinIOSpec(src *MinIOSpec, dst *v1alpha3.MinIOSpec) {
	dst = &v1alpha3.MinIOSpec{}

	dst.SecretRef = src.SecretRef
	dst.VolumeClaimTemplate = src.VolumeClaimTemplate
	dst.VolumesPerServer = src.VolumesPerServer

	dst.ImageSpec = src.ImageSpec

	dst.Replicas = src.Replicas
	dst.Resources = src.Resources

	dst.Redirect.Enable = src.Redirect.Enable
	if src.Redirect.Expose != nil {
		Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src.Redirect.Expose, dst.Redirect.Expose)
	}
}

func Convert_v1beta1_HarborExposeSpec_To_v1alpha3_HarborExposeSpec(src *HarborExposeSpec, dst *v1alpha3.HarborExposeSpec) {
	if dst == nil {
		dst = &v1alpha3.HarborExposeSpec{}
	}

	Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(&src.Core, &dst.Core)

	Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src.Notary, dst.Notary)
}

func Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src *HarborExposeComponentSpec, dst *v1alpha3.HarborExposeComponentSpec) {
	if dst == nil {
		dst = &v1alpha3.HarborExposeComponentSpec{}
	}

	if src == nil {
		return
	}

	if src.Ingress != nil {
		Convert_v1beta1_HarborExposeIngressSpec_To_v1alpha3_HarborExposeIngressSpec(src.Ingress, dst.Ingress)
	}

	if src.TLS != nil {
		dst.TLS = src.TLS
	}
}

func Convert_v1beta1_HarborExposeIngressSpec_To_v1alpha3_HarborExposeIngressSpec(src *HarborExposeIngressSpec, dst *v1alpha3.HarborExposeIngressSpec) {
	if dst == nil {
		dst = &v1alpha3.HarborExposeIngressSpec{}
	}

	if src == nil {
		return
	}

	dst.Host = src.Host
	dst.Controller = src.Controller
	dst.Annotations = src.Annotations
}

func Convert_v1beta1_Database_To_v1alpha3_Database(src *Database, dst *v1alpha3.Database) {
	dst = &v1alpha3.Database{}

	dst.Kind = src.Kind

	if src.Spec.ZlandoPostgreSQL != nil {
		dst.PostgresSQLSpec = &v1alpha3.PostgresSQLSpec{}
		Convert_v1beta1_PostgresSQLSpec_To_v1alpha3_ZlandoPostgresSQLSpec(src.Spec.ZlandoPostgreSQL, dst.PostgresSQLSpec)
	}
}

func Convert_v1beta1_PostgresSQLSpec_To_v1alpha3_ZlandoPostgresSQLSpec(src *ZlandoPostgreSQLSpec, dst *v1alpha3.PostgresSQLSpec) {
	dst.Storage = src.Storage
	dst.Resources = src.Resources
	dst.Replicas = src.Replicas

	dst.ImageSpec = src.ImageSpec
	dst.StorageClassName = src.StorageClassName
}

func Convert_v1beta1_HarborClusterStatus_To_v1alpha3_HarborClusterStatus(src *HarborClusterStatus, dst *v1alpha3.HarborClusterStatus) {
	dst.Operator = src.Operator
	dst.Status = v1alpha3.ClusterStatus(src.Status)
	dst.ObservedGeneration = src.ObservedGeneration
	dst.Revision = src.Revision

	dst.Conditions = func() []v1alpha3.HarborClusterCondition {
		conditions := make([]v1alpha3.HarborClusterCondition, 0, len(src.Conditions))
		for _, cond := range src.Conditions {
			conditions = append(conditions, v1alpha3.HarborClusterCondition{
				Type:               v1alpha3.HarborClusterConditionType(cond.Type),
				Status:             cond.Status,
				LastTransitionTime: cond.LastTransitionTime,
				Reason:             cond.Reason,
				Message:            cond.Message,
			})
		}
		return conditions
	}()
}
