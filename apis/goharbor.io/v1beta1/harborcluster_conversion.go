package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &HarborCluster{}

func (h *HarborCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.HarborCluster)

	dst.ObjectMeta = h.ObjectMeta

	Convert_v1beta1_HarborClusterSpec_To_v1alpha3_HarborClusterSpec(&h.Spec, &dst.Spec)

	Convert_v1beta1_HarborClusterStatus_To_v1alpha3_HarborClusterStatus(&h.Status, &dst.Status)

	return nil
}

func (h *HarborCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.HarborCluster)

	h.ObjectMeta = src.ObjectMeta

	Convert_v1alpha3_HarborClusterSpec_To_v1beta1_HarborClusterSpec(&src.Spec, &h.Spec)

	Convert_v1alpha3_HarborClusterStatus_To_v1beta1_HarborClusterStatus(&src.Status, &h.Status)

	return nil
}

func Convert_v1alpha3_HarborClusterSpec_To_v1beta1_HarborClusterSpec(src *v1alpha3.HarborClusterSpec, dst *HarborClusterSpec) { // nolint
	if src.InClusterCache != nil {
		Convert_v1alpha3_Cache_To_v1beta1_Cache(src.InClusterCache, &dst.Cache)
	} else if src.Redis != nil {
		dst.Cache = Cache{
			Kind: "Redis",
			Spec: &CacheSpec{
				Redis: &ExternalRedisSpec{
					RedisHostSpec:    src.Redis.RedisHostSpec,
					RedisCredentials: src.Redis.RedisCredentials,
				},
			},
		}
	}

	if src.InClusterStorage != nil {
		dst.Storage = Storage{}
		Convert_v1alpha3_Storage_To_v1beta1_Storage(src.InClusterStorage, &dst.Storage)
	} else if src.ImageChartStorage != nil {
		dst.Storage = Storage{}
		Convert_v1alpha3_HarborStorageImageChartStorageSpec_To_v1beta1_Storage(src.ImageChartStorage, &dst.Storage)
	}

	if src.InClusterDatabase != nil {
		dst.Database = Database{}
		Convert_v1alpha3_Database_To_v1beta1_Database(src.InClusterDatabase, &dst.Database)
	} else if src.Database != nil {
		dst.Database = Database{}
		Convert_v1alpha3_HarborDatabaseSpec_To_v1beta1_Database(src.Database, &dst.Database)
	}

	Convert_v1alpha3_HarborSpec_To_v1beta1_HarborSpec(&src.HarborSpec, &dst.EmbeddedHarborSpec)
}

func Convert_v1alpha3_HarborSpec_To_v1beta1_HarborSpec(src *v1alpha3.HarborSpec, dst *EmbeddedHarborSpec) { // nolint
	dst.ExternalURL = src.ExternalURL
	dst.InternalTLS = HarborInternalTLSSpec{
		Enabled: src.InternalTLS.Enabled,
	}
	// dst.ImageChartStorage
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

	if src.ImageSource != nil {
		dst.ImageSource = &ImageSourceSpec{}
		Convert_v1alpha3_ImageSourceSpec_To_v1beta1_ImageSourceSpec(src.ImageSource, dst.ImageSource)
	}

	Convert_v1alpha3_HarborExposeSpec_To_v1beta1_HarborExposeSpec(&src.Expose, &dst.Expose)

	Convert_v1alpha3_HarborComponentSpec_To_v1beta1_HarborComponentSpec(&src.HarborComponentsSpec, &dst.HarborComponentsSpec)
}

func Convert_v1alpha3_HarborComponentSpec_To_v1beta1_HarborComponentSpec(src *v1alpha3.HarborComponentsSpec, dst *HarborComponentsSpec) { // nolint
	Convert_v1alpha3_CoreComponentSpec_To_v1beta1_CoreComponentSpec(&src.Core, &dst.Core)

	Convert_v1alpha3_RegistryComponentSpec_To_v1beta1_RegistryComponentSpec(&src.Registry, &dst.Registry)

	Convert_v1alpha3_JobServiceComponentSpec_To_v1beta1_JobServiceComponentSpec(&src.JobService, &dst.JobService)

	if src.ChartMuseum != nil {
		dst.ChartMuseum = &ChartMuseumComponentSpec{}
		Convert_v1alpha3_ChartMuseumComponentSpec_To_v1beta1_ChartMuseumComponentSpec(src.ChartMuseum, dst.ChartMuseum)
	}

	if src.Notary != nil {
		dst.Notary = &NotaryComponentSpec{}
		Convert_v1alpha3_NotaryComponentSpec_To_v1beta1_NotaryComponentSpec(src.Notary, dst.Notary)
	}

	if src.Trivy != nil {
		dst.Trivy = &TrivyComponentSpec{}
		Convert_v1alpha3_TrivyComponentSpec_To_v1beta1_TrivyComponentSpec(src.Trivy, dst.Trivy)
	}

	if src.Exporter != nil {
		dst.Exporter = &ExporterComponentSpec{}
		Convert_v1alpha3_ExporterComponentSpec_To_v1beta1_ExporterComponentSpec(src.Exporter, dst.Exporter)
	}
}

func Convert_v1alpha3_CoreComponentSpec_To_v1beta1_CoreComponentSpec(src *v1alpha3.CoreComponentSpec, dst *CoreComponentSpec) { // nolint
	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.TokenIssuer = src.TokenIssuer
}

func Convert_v1alpha3_RegistryComponentSpec_To_v1beta1_RegistryComponentSpec(src *v1alpha3.RegistryComponentSpec, dst *RegistryComponentSpec) { // nolint
	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
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

func Convert_v1alpha3_JobServiceComponentSpec_To_v1beta1_JobServiceComponentSpec(src *v1alpha3.JobServiceComponentSpec, dst *JobServiceComponentSpec) { // nolint
	dst.WorkerCount = src.WorkerCount
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
}

func Convert_v1alpha3_ChartMuseumComponentSpec_To_v1beta1_ChartMuseumComponentSpec(src *v1alpha3.ChartMuseumComponentSpec, dst *ChartMuseumComponentSpec) { // nolint
	dst.AbsoluteURL = src.AbsoluteURL
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
}

func Convert_v1alpha3_ExporterComponentSpec_To_v1beta1_ExporterComponentSpec(src *v1alpha3.ExporterComponentSpec, dst *ExporterComponentSpec) { // nolint
	dst.ComponentSpec = src.ComponentSpec
	dst.Port = src.Port
	dst.Path = src.Path

	Convert_v1alpha3_HarborExporterCacheSpec_To_v1beta1_HarborExporterCacheSpec(&src.Cache, &dst.Cache)
}

func Convert_v1alpha3_HarborExporterCacheSpec_To_v1beta1_HarborExporterCacheSpec(src *v1alpha3.HarborExporterCacheSpec, dst *HarborExporterCacheSpec) { // nolint
	dst.Duration = src.Duration
	dst.CleanInterval = src.CleanInterval
}

func Convert_v1alpha3_TrivyComponentSpec_To_v1beta1_TrivyComponentSpec(src *v1alpha3.TrivyComponentSpec, dst *TrivyComponentSpec) { // nolint
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
	dst.GithubTokenRef = src.GithubTokenRef
	dst.SkipUpdate = src.SkipUpdate

	Convert_v1alpha3_HarborStorageTrivyStorageSpec_To_v1beta1_HarborStorageTrivyStorageSpec(&src.Storage, &dst.Storage)
}

func Convert_v1alpha3_HarborStorageTrivyStorageSpec_To_v1beta1_HarborStorageTrivyStorageSpec(src *v1alpha3.HarborStorageTrivyStorageSpec, dst *HarborStorageTrivyStorageSpec) { // nolint
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

func Convert_v1alpha3_NotaryComponentSpec_To_v1beta1_NotaryComponentSpec(src *v1alpha3.NotaryComponentSpec, dst *NotaryComponentSpec) { // nolint
	dst.Server = src.Server
	dst.Signer = src.Signer
	dst.MigrationEnabled = src.MigrationEnabled
}

func Convert_v1alpha3_ImageSourceSpec_To_v1beta1_ImageSourceSpec(src *v1alpha3.ImageSourceSpec, dst *ImageSourceSpec) { // nolint
	dst.ImagePullPolicy = src.ImagePullPolicy
	dst.ImagePullSecrets = src.ImagePullSecrets
	dst.Repository = src.Repository
	dst.TagSuffix = src.TagSuffix
}

func Convert_v1alpha3_Cache_To_v1beta1_Cache(src *v1alpha3.Cache, dst *Cache) { // nolint
	if src.RedisSpec != nil {
		dst.Kind = "RedisFailover"
		dst.Spec = &CacheSpec{}
		Convert_v1alpha3_RedisSpec_To_v1beta1_CacheSpec(src.RedisSpec, dst.Spec)
	}
}

func Convert_v1alpha3_RedisSpec_To_v1beta1_CacheSpec(src *v1alpha3.RedisSpec, dst *CacheSpec) { // nolint
	dst.RedisFailover = &RedisFailoverSpec{}

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

	dst.RedisFailover.OperatorVersion = "1.0.0"
	dst.RedisFailover.ImageSpec = src.ImageSpec
}

func Convert_v1alpha3_HarborStorageImageChartStorageSpec_To_v1beta1_Storage(src *v1alpha3.HarborStorageImageChartStorageSpec, dst *Storage) { // nolint
	if src.FileSystem != nil {
		dst.Kind = "FileSystem"
		dst.Spec = StorageSpec{
			FileSystem: &FileSystemSpec{
				HarborStorageImageChartStorageFileSystemSpec: HarborStorageImageChartStorageFileSystemSpec{
					RegistryPersistentVolume: HarborStorageRegistryPersistentVolumeSpec{
						HarborStoragePersistentVolumeSpec: HarborStoragePersistentVolumeSpec{
							PersistentVolumeClaimVolumeSource: src.FileSystem.RegistryPersistentVolume.PersistentVolumeClaimVolumeSource,
							Prefix:                            src.FileSystem.RegistryPersistentVolume.Prefix,
						},
						MaxThreads: src.FileSystem.RegistryPersistentVolume.MaxThreads,
					},
				},
			},
		}

		if src.FileSystem.ChartPersistentVolume != nil {
			dst.Spec.FileSystem.ChartPersistentVolume = &HarborStoragePersistentVolumeSpec{
				PersistentVolumeClaimVolumeSource: src.FileSystem.ChartPersistentVolume.PersistentVolumeClaimVolumeSource,
				Prefix:                            src.FileSystem.ChartPersistentVolume.Prefix,
			}
		}
	}
}

func Convert_v1alpha3_Storage_To_v1beta1_Storage(src *v1alpha3.Storage, dst *Storage) { // nolint
	if src.MinIOSpec != nil {
		dst.Kind = "MinIO"
		dst.Spec.MinIO = &MinIOSpec{}
		Convert_v1alpha3_MinIOSpec_to_v1beta1_MinIOSpec(src.MinIOSpec, dst.Spec.MinIO)
	}
}

func Convert_v1alpha3_MinIOSpec_to_v1beta1_MinIOSpec(src *v1alpha3.MinIOSpec, dst *MinIOSpec) { // nolint
	dst.OperatorVersion = "4.0.6"

	dst.SecretRef = src.SecretRef
	dst.VolumeClaimTemplate = src.VolumeClaimTemplate
	dst.VolumesPerServer = src.VolumesPerServer

	dst.ImageSpec = src.ImageSpec

	dst.Replicas = src.Replicas
	dst.Resources = src.Resources

	dst.Redirect.Enable = src.Redirect.Enable
	if src.Redirect.Expose != nil {
		dst.Redirect.Expose = &HarborExposeComponentSpec{}
		Convert_v1alpha3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src.Redirect.Expose, dst.Redirect.Expose)
	}
}

func Convert_v1alpha3_HarborExposeSpec_To_v1beta1_HarborExposeSpec(src *v1alpha3.HarborExposeSpec, dst *HarborExposeSpec) { // nolint
	Convert_v1alpha3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(&src.Core, &dst.Core)

	if src.Notary != nil {
		dst.Notary = &HarborExposeComponentSpec{}
		Convert_v1alpha3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src.Notary, dst.Notary)
	}
}

func Convert_v1alpha3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src *v1alpha3.HarborExposeComponentSpec, dst *HarborExposeComponentSpec) { // nolint
	if src.Ingress != nil {
		dst.Ingress = &HarborExposeIngressSpec{}
		Convert_v1alpha3_HarborExposeIngressSpec_To_v1beta1_HarborExposeIngressSpec(src.Ingress, dst.Ingress)
	}

	if src.TLS != nil {
		dst.TLS = src.TLS
	}
}

func Convert_v1alpha3_HarborExposeIngressSpec_To_v1beta1_HarborExposeIngressSpec(src *v1alpha3.HarborExposeIngressSpec, dst *HarborExposeIngressSpec) { // nolint
	dst.Host = src.Host
	dst.Controller = src.Controller
	dst.Annotations = src.Annotations
}

func Convert_v1alpha3_HarborDatabaseSpec_To_v1beta1_Database(src *v1alpha3.HarborDatabaseSpec, dst *Database) { // nolint
	dst.Kind = "PostgreSQL"
	dst.Spec = DatabaseSpec{
		PostgreSQL: &PostgreSQLSpec{
			HarborDatabaseSpec{
				PostgresCredentials: src.PostgresCredentials,
				Hosts:               src.Hosts,
				SSLMode:             src.SSLMode,
				Prefix:              src.Prefix,
			},
		},
	}
}

func Convert_v1alpha3_Database_To_v1beta1_Database(src *v1alpha3.Database, dst *Database) { // nolint
	if src.PostgresSQLSpec != nil {
		dst.Kind = "Zlando/PostgreSQL"
		dst.Spec.ZlandoPostgreSQL = &ZlandoPostgreSQLSpec{}
		Convert_v1alpha3_PostgresSQLSpec_To_v1beta1_ZlandoPostgresSQLSpec(src.PostgresSQLSpec, dst.Spec.ZlandoPostgreSQL)
	}
}

func Convert_v1alpha3_PostgresSQLSpec_To_v1beta1_ZlandoPostgresSQLSpec(src *v1alpha3.PostgresSQLSpec, dst *ZlandoPostgreSQLSpec) { // nolint
	dst.OperatorVersion = "1.5.0"
	dst.Storage = src.Storage
	dst.Resources = src.Resources
	dst.Replicas = src.Replicas

	dst.ImageSpec = src.ImageSpec
	dst.StorageClassName = src.StorageClassName
}

func Convert_v1alpha3_HarborClusterStatus_To_v1beta1_HarborClusterStatus(src *v1alpha3.HarborClusterStatus, dst *HarborClusterStatus) { // nolint
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

func Convert_v1beta1_HarborClusterSpec_To_v1alpha3_HarborClusterSpec(src *HarborClusterSpec, dst *v1alpha3.HarborClusterSpec) { // nolint
	if src.Cache.Kind == "Redis" && src.Cache.Spec.Redis != nil {
		if dst.Redis == nil {
			dst.Redis = &v1alpha3.ExternalRedisSpec{}
		}

		Convert_v1beta1_ExternalRedisSpec_To_v1alpha3_ExternalRedisSpec(src.Cache.Spec.Redis, dst.Redis)
	} else if src.Cache.Kind == "RedisFailover" {
		if dst.InClusterCache == nil {
			dst.InClusterCache = &v1alpha3.Cache{}
		}

		Convert_v1beta1_Cache_To_v1alpha3_Cache(&src.Cache, dst.InClusterCache)
	}

	if src.Storage.Kind == "FileSystem" && src.Storage.Spec.FileSystem != nil {
		if dst.ImageChartStorage == nil {
			dst.ImageChartStorage = &v1alpha3.HarborStorageImageChartStorageSpec{}
		}

		Convert_v1beta1_FileSystemSpec_To_v1alpha3_HarborStorageImageChartStorage(src.Storage.Spec.FileSystem, dst.ImageChartStorage)
	} else if src.Storage.Kind == "MinIO" {
		if dst.InClusterStorage == nil {
			dst.InClusterStorage = &v1alpha3.Storage{}
		}

		Convert_v1beta1_Storage_To_v1alpha3_Storage(&src.Storage, dst.InClusterStorage)
	}

	if src.Database.Kind == "PostgreSQL" && src.Database.Spec.PostgreSQL != nil {
		if dst.Database == nil {
			dst.Database = &v1alpha3.HarborDatabaseSpec{}
		}

		Convert_v1beta1_PostgreSQLSpec_To_v1alpha3_HarborDatabaseSpec(src.Database.Spec.PostgreSQL, dst.Database)
	} else if src.Database.Kind == "Zlando/PostgreSQL" {
		if dst.InClusterDatabase == nil {
			dst.InClusterDatabase = &v1alpha3.Database{}
		}

		Convert_v1beta1_Database_To_v1alpha3_Database(&src.Database, dst.InClusterDatabase)
	}

	Convert_v1beta1_EmbeddedHarborSpec_To_v1alpha3_HarborSpec(&src.EmbeddedHarborSpec, &dst.HarborSpec)
}

func Convert_v1beta1_EmbeddedHarborSpec_To_v1alpha3_HarborSpec(src *EmbeddedHarborSpec, dst *v1alpha3.HarborSpec) { // nolint
	dst.ExternalURL = src.ExternalURL
	dst.InternalTLS = v1alpha3.HarborInternalTLSSpec{
		Enabled: src.InternalTLS.Enabled,
	}

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

	if src.ImageSource != nil {
		dst.ImageSource = &v1alpha3.ImageSourceSpec{}
		Convert_v1beta1_ImageSourceSpec_To_v1alpha3_ImageSourceSpec(src.ImageSource, dst.ImageSource)
	}

	Convert_v1beta1_HarborExposeSpec_To_v1alpha3_HarborExposeSpec(&src.Expose, &dst.Expose)

	Convert_v1beta1_HarborComponentSpec_To_v1alpha3_HarborComponentSpec(&src.HarborComponentsSpec, &dst.HarborComponentsSpec)
}

func Convert_v1beta1_HarborComponentSpec_To_v1alpha3_HarborComponentSpec(src *HarborComponentsSpec, dst *v1alpha3.HarborComponentsSpec) { // nolint
	Convert_v1beta1_CoreComponentSpec_To_v1alpha3_CoreComponentSpec(&src.Core, &dst.Core)

	Convert_v1beta1_RegistryComponentSpec_To_v1alpha3_RegistryComponentSpec(&src.Registry, &dst.Registry)

	Convert_v1beta1_JobServiceComponentSpec_To_v1alpha3_JobServiceComponentSpec(&src.JobService, &dst.JobService)

	if src.ChartMuseum != nil {
		dst.ChartMuseum = &v1alpha3.ChartMuseumComponentSpec{}
		Convert_v1beta1_ChartMuseumComponentSpec_To_v1alpha3_ChartMuseumComponentSpec(src.ChartMuseum, dst.ChartMuseum)
	}

	if src.Notary != nil {
		dst.Notary = &v1alpha3.NotaryComponentSpec{}
		Convert_v1beta1_NotaryComponentSpec_To_v1alpha3_NotaryComponentSpec(src.Notary, dst.Notary)
	}

	if src.Trivy != nil {
		dst.Trivy = &v1alpha3.TrivyComponentSpec{}
		Convert_v1beta1_TrivyComponentSpec_To_v1alpha3_TrivyComponentSpec(src.Trivy, dst.Trivy)
	}

	if src.Exporter != nil {
		dst.Exporter = &v1alpha3.ExporterComponentSpec{}
		Convert_v1beta1_ExporterComponentSpec_To_v1alpha3_ExporterComponentSpec(src.Exporter, dst.Exporter)
	}
}

func Convert_v1beta1_CoreComponentSpec_To_v1alpha3_CoreComponentSpec(src *CoreComponentSpec, dst *v1alpha3.CoreComponentSpec) { // nolint
	dst.CertificateInjection = v1alpha3.CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.TokenIssuer = src.TokenIssuer
}

func Convert_v1beta1_RegistryComponentSpec_To_v1alpha3_RegistryComponentSpec(src *RegistryComponentSpec, dst *v1alpha3.RegistryComponentSpec) { // nolint
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

func Convert_v1beta1_JobServiceComponentSpec_To_v1alpha3_JobServiceComponentSpec(src *JobServiceComponentSpec, dst *v1alpha3.JobServiceComponentSpec) { // nolint
	dst.WorkerCount = src.WorkerCount
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1alpha3.CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
}

func Convert_v1beta1_ChartMuseumComponentSpec_To_v1alpha3_ChartMuseumComponentSpec(src *ChartMuseumComponentSpec, dst *v1alpha3.ChartMuseumComponentSpec) { // nolint
	dst.AbsoluteURL = src.AbsoluteURL
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1alpha3.CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
}

func Convert_v1beta1_ExporterComponentSpec_To_v1alpha3_ExporterComponentSpec(src *ExporterComponentSpec, dst *v1alpha3.ExporterComponentSpec) { // nolint
	dst.ComponentSpec = src.ComponentSpec
	dst.Port = src.Port
	dst.Path = src.Path

	Convert_v1beta1_HarborExporterCacheSpec_To_v1alpha3_HarborExporterCacheSpec(&src.Cache, &dst.Cache)
}

func Convert_v1beta1_HarborExporterCacheSpec_To_v1alpha3_HarborExporterCacheSpec(src *HarborExporterCacheSpec, dst *v1alpha3.HarborExporterCacheSpec) { // nolint
	dst.Duration = src.Duration
	dst.CleanInterval = src.CleanInterval
}

func Convert_v1beta1_TrivyComponentSpec_To_v1alpha3_TrivyComponentSpec(src *TrivyComponentSpec, dst *v1alpha3.TrivyComponentSpec) { // nolint
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1alpha3.CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
	dst.GithubTokenRef = src.GithubTokenRef
	dst.SkipUpdate = src.SkipUpdate

	Convert_v1beta1_HarborStorageTrivyStorageSpec_To_v1alpha3_HarborStorageTrivyStorageSpec(&src.Storage, &dst.Storage)
}

func Convert_v1beta1_HarborStorageTrivyStorageSpec_To_v1alpha3_HarborStorageTrivyStorageSpec(src *HarborStorageTrivyStorageSpec, dst *v1alpha3.HarborStorageTrivyStorageSpec) { // nolint
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

func Convert_v1beta1_NotaryComponentSpec_To_v1alpha3_NotaryComponentSpec(src *NotaryComponentSpec, dst *v1alpha3.NotaryComponentSpec) { // nolint
	dst.Server = src.Server
	dst.Signer = src.Signer
	dst.MigrationEnabled = src.MigrationEnabled
}

func Convert_v1beta1_ImageSourceSpec_To_v1alpha3_ImageSourceSpec(src *ImageSourceSpec, dst *v1alpha3.ImageSourceSpec) { // nolint
	dst.ImagePullPolicy = src.ImagePullPolicy
	dst.ImagePullSecrets = src.ImagePullSecrets
	dst.Repository = src.Repository
	dst.TagSuffix = src.TagSuffix
}

func Convert_v1beta1_ExternalRedisSpec_To_v1alpha3_ExternalRedisSpec(src *ExternalRedisSpec, dst *v1alpha3.ExternalRedisSpec) { // nolint
	dst.RedisCredentials = src.RedisCredentials
	dst.RedisHostSpec = src.RedisHostSpec
}

func Convert_v1beta1_Cache_To_v1alpha3_Cache(src *Cache, dst *v1alpha3.Cache) { // nolint
	dst.Kind = "Redis"
	if src.Spec != nil {
		dst.RedisSpec = &v1alpha3.RedisSpec{}
		Convert_v1beta1_RedisSpec_To_v1alpha3_CacheSpec(src.Spec, dst.RedisSpec)
	}
}

func Convert_v1beta1_RedisSpec_To_v1alpha3_CacheSpec(src *CacheSpec, dst *v1alpha3.RedisSpec) { // nolint
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

func Convert_v1beta1_FileSystemSpec_To_v1alpha3_HarborStorageImageChartStorage(src *FileSystemSpec, dst *v1alpha3.HarborStorageImageChartStorageSpec) { // nolint
	dst.FileSystem = &v1alpha3.HarborStorageImageChartStorageFileSystemSpec{}

	if src.ChartPersistentVolume != nil {
		dst.FileSystem.ChartPersistentVolume = &v1alpha3.HarborStoragePersistentVolumeSpec{
			PersistentVolumeClaimVolumeSource: src.ChartPersistentVolume.PersistentVolumeClaimVolumeSource,
			Prefix:                            src.ChartPersistentVolume.Prefix,
		}
	}

	dst.FileSystem.RegistryPersistentVolume = v1alpha3.HarborStorageRegistryPersistentVolumeSpec{
		HarborStoragePersistentVolumeSpec: v1alpha3.HarborStoragePersistentVolumeSpec{
			PersistentVolumeClaimVolumeSource: src.RegistryPersistentVolume.PersistentVolumeClaimVolumeSource,
			Prefix:                            src.RegistryPersistentVolume.Prefix,
		},
		MaxThreads: src.RegistryPersistentVolume.MaxThreads,
	}
}

func Convert_v1beta1_Storage_To_v1alpha3_Storage(src *Storage, dst *v1alpha3.Storage) { // nolint
	dst.Kind = src.Kind

	if src.Spec.MinIO != nil {
		dst.MinIOSpec = &v1alpha3.MinIOSpec{}
		Convert_v1beta1_MinIOSpec_To_v1alpha3_MinIOSpec(src.Spec.MinIO, dst.MinIOSpec)
	}
}

func Convert_v1beta1_MinIOSpec_To_v1alpha3_MinIOSpec(src *MinIOSpec, dst *v1alpha3.MinIOSpec) { // nolint
	dst.SecretRef = src.SecretRef
	dst.VolumeClaimTemplate = src.VolumeClaimTemplate
	dst.VolumesPerServer = src.VolumesPerServer

	dst.ImageSpec = src.ImageSpec

	dst.Replicas = src.Replicas
	dst.Resources = src.Resources

	dst.Redirect.Enable = src.Redirect.Enable
	if src.Redirect.Expose != nil {
		dst.Redirect.Expose = &v1alpha3.HarborExposeComponentSpec{}
		Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src.Redirect.Expose, dst.Redirect.Expose)
	}
}

func Convert_v1beta1_HarborExposeSpec_To_v1alpha3_HarborExposeSpec(src *HarborExposeSpec, dst *v1alpha3.HarborExposeSpec) { // nolint
	Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(&src.Core, &dst.Core)

	if src.Notary != nil {
		dst.Notary = &v1alpha3.HarborExposeComponentSpec{}
		Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src.Notary, dst.Notary)
	}
}

func Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src *HarborExposeComponentSpec, dst *v1alpha3.HarborExposeComponentSpec) { // nolint
	if src.Ingress != nil {
		dst.Ingress = &v1alpha3.HarborExposeIngressSpec{}
		Convert_v1beta1_HarborExposeIngressSpec_To_v1alpha3_HarborExposeIngressSpec(src.Ingress, dst.Ingress)
	}

	if src.TLS != nil {
		dst.TLS = src.TLS
	}
}

func Convert_v1beta1_HarborExposeIngressSpec_To_v1alpha3_HarborExposeIngressSpec(src *HarborExposeIngressSpec, dst *v1alpha3.HarborExposeIngressSpec) { // nolint
	dst.Host = src.Host
	dst.Controller = src.Controller
	dst.Annotations = src.Annotations
}

func Convert_v1beta1_PostgreSQLSpec_To_v1alpha3_HarborDatabaseSpec(src *PostgreSQLSpec, dst *v1alpha3.HarborDatabaseSpec) { // nolint
	dst.PostgresCredentials = src.PostgresCredentials
	dst.Hosts = src.Hosts
	dst.Prefix = src.Prefix
	dst.SSLMode = src.SSLMode
}

func Convert_v1beta1_Database_To_v1alpha3_Database(src *Database, dst *v1alpha3.Database) { // nolint
	dst.Kind = "PostgresSQL"

	if src.Spec.ZlandoPostgreSQL != nil {
		dst.PostgresSQLSpec = &v1alpha3.PostgresSQLSpec{}
		Convert_v1beta1_ZlandoPostgreSQLSpec_To_v1alpha3_PostgresSQLSpec(src.Spec.ZlandoPostgreSQL, dst.PostgresSQLSpec)
	}
}

func Convert_v1beta1_ZlandoPostgreSQLSpec_To_v1alpha3_PostgresSQLSpec(src *ZlandoPostgreSQLSpec, dst *v1alpha3.PostgresSQLSpec) { // nolint
	dst.Storage = src.Storage
	dst.Resources = src.Resources
	dst.Replicas = src.Replicas

	dst.ImageSpec = src.ImageSpec
	dst.StorageClassName = src.StorageClassName
}

func Convert_v1beta1_HarborClusterStatus_To_v1alpha3_HarborClusterStatus(src *HarborClusterStatus, dst *v1alpha3.HarborClusterStatus) { // nolint
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
