package v1alpha3

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &HarborCluster{}

func (hc *HarborCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.HarborCluster)

	dst.ObjectMeta = hc.ObjectMeta

	Convert_v1alpha3_HarborClusterSpec_To_v1beta1_HarborClusterSpec(&hc.Spec, &dst.Spec)

	Convert_v1alpha3_HarborClusterStatus_To_v1beta1_HarborClusterStatus(&hc.Status, &dst.Status)

	return nil
}

func (hc *HarborCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.HarborCluster)

	hc.ObjectMeta = src.ObjectMeta

	Convert_v1beta1_HarborClusterSpec_To_v1alpha3_HarborClusterSpec(&src.Spec, &hc.Spec)

	Convert_v1beta1_HarborClusterStatus_To_v1alpha3_HarborClusterStatus(&src.Status, &hc.Status)

	return nil
}

func Convert_v1alpha3_HarborClusterSpec_To_v1beta1_HarborClusterSpec(src *HarborClusterSpec, dst *v1beta1.HarborClusterSpec) { //nolint
	if src.InClusterCache != nil {
		Convert_v1alpha3_Cache_To_v1beta1_Cache(src.InClusterCache, &dst.Cache)
	} else if src.Redis != nil {
		dst.Cache = v1beta1.Cache{
			Kind: v1beta1.KindCacheRedis,
			Spec: &v1beta1.CacheSpec{
				Redis: &v1beta1.ExternalRedisSpec{
					RedisHostSpec:    src.Redis.RedisHostSpec,
					RedisCredentials: src.Redis.RedisCredentials,
				},
			},
		}
	}

	if src.InClusterStorage != nil {
		dst.Storage = v1beta1.Storage{}
		Convert_v1alpha3_Storage_To_v1beta1_Storage(src.InClusterStorage, &dst.Storage)
	} else if src.ImageChartStorage != nil {
		dst.Storage = v1beta1.Storage{}
		Convert_v1alpha3_HarborStorageImageChartStorageSpec_To_v1beta1_Storage(src.ImageChartStorage, &dst.Storage)
	}

	if src.InClusterDatabase != nil {
		dst.Database = v1beta1.Database{}
		Convert_v1alpha3_Database_To_v1beta1_Database(src.InClusterDatabase, &dst.Database)
	} else if src.Database != nil {
		dst.Database = v1beta1.Database{}
		Convert_v1alpha3_HarborDatabaseSpec_To_v1beta1_Database(src.Database, &dst.Database)
	}

	Convert_v1alpha3_HarborSpec_To_v1beta1_HarborSpec(&src.HarborSpec, &dst.EmbeddedHarborSpec)
}

func Convert_v1alpha3_HarborSpec_To_v1beta1_HarborSpec(src *HarborSpec, dst *v1beta1.EmbeddedHarborSpec) { //nolint
	dst.ExternalURL = src.ExternalURL
	dst.InternalTLS = v1beta1.HarborInternalTLSSpec{
		Enabled: src.InternalTLS.Enabled,
	}
	// dst.ImageChartStorage
	dst.LogLevel = src.LogLevel
	dst.HarborAdminPasswordRef = src.HarborAdminPasswordRef
	dst.UpdateStrategyType = src.UpdateStrategyType
	dst.Version = src.Version
	dst.ImageSource = src.ImageSource.DeepCopy()

	if src.Proxy != nil {
		dst.Proxy = &v1beta1.HarborProxySpec{
			ProxySpec:  src.Proxy.ProxySpec,
			Components: src.Proxy.Components,
		}
	}

	Convert_v1alpha3_HarborExposeSpec_To_v1beta1_HarborExposeSpec(&src.Expose, &dst.Expose)

	Convert_v1alpha3_HarborComponentSpec_To_v1beta1_EmbeddedHarborComponentsSpec(&src.HarborComponentsSpec, &dst.EmbeddedHarborComponentsSpec)
}

func Convert_v1alpha3_HarborComponentSpec_To_v1beta1_EmbeddedHarborComponentsSpec(src *HarborComponentsSpec, dst *v1beta1.EmbeddedHarborComponentsSpec) { //nolint
	Convert_v1alpha3_CoreComponentSpec_To_v1beta1_CoreComponentSpec(&src.Core, &dst.Core)

	Convert_v1alpha3_RegistryComponentSpec_To_v1beta1_RegistryComponentSpec(&src.Registry, &dst.Registry)

	Convert_v1alpha3_JobServiceComponentSpec_To_v1beta1_JobServiceComponentSpec(&src.JobService, &dst.JobService)

	if src.ChartMuseum != nil {
		dst.ChartMuseum = &v1beta1.ChartMuseumComponentSpec{}
		Convert_v1alpha3_ChartMuseumComponentSpec_To_v1beta1_ChartMuseumComponentSpec(src.ChartMuseum, dst.ChartMuseum)
	}

	if src.Notary != nil {
		dst.Notary = &v1beta1.NotaryComponentSpec{}
		Convert_v1alpha3_NotaryComponentSpec_To_v1beta1_NotaryComponentSpec(src.Notary, dst.Notary)
	}

	if src.Trivy != nil {
		dst.Trivy = &v1beta1.TrivyComponentSpec{}
		Convert_v1alpha3_TrivyComponentSpec_To_v1beta1_TrivyComponentSpec(src.Trivy, dst.Trivy)
	}

	if src.Exporter != nil {
		dst.Exporter = &v1beta1.ExporterComponentSpec{}
		Convert_v1alpha3_ExporterComponentSpec_To_v1beta1_ExporterComponentSpec(src.Exporter, dst.Exporter)
	}
}

func Convert_v1alpha3_CoreComponentSpec_To_v1beta1_CoreComponentSpec(src *CoreComponentSpec, dst *v1beta1.CoreComponentSpec) { //nolint
	dst.CertificateInjection = v1beta1.CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.TokenIssuer = src.TokenIssuer
}

func Convert_v1alpha3_RegistryComponentSpec_To_v1beta1_RegistryComponentSpec(src *RegistryComponentSpec, dst *v1beta1.RegistryComponentSpec) { //nolint
	dst.CertificateInjection = v1beta1.CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.RelativeURLs = src.RelativeURLs

	if src.StorageMiddlewares != nil {
		dst.StorageMiddlewares = make([]v1beta1.RegistryMiddlewareSpec, 0, len(src.StorageMiddlewares))
		for _, value := range src.StorageMiddlewares {
			dst.StorageMiddlewares = append(dst.StorageMiddlewares, v1beta1.RegistryMiddlewareSpec{
				Name:       value.Name,
				OptionsRef: value.OptionsRef,
			})
		}
	}
}

func Convert_v1alpha3_JobServiceComponentSpec_To_v1beta1_JobServiceComponentSpec(src *JobServiceComponentSpec, dst *v1beta1.JobServiceComponentSpec) { //nolint
	dst.WorkerCount = src.WorkerCount
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1beta1.CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
}

func Convert_v1alpha3_ChartMuseumComponentSpec_To_v1beta1_ChartMuseumComponentSpec(src *ChartMuseumComponentSpec, dst *v1beta1.ChartMuseumComponentSpec) { //nolint
	dst.AbsoluteURL = src.AbsoluteURL
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1beta1.CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
}

func Convert_v1alpha3_ExporterComponentSpec_To_v1beta1_ExporterComponentSpec(src *ExporterComponentSpec, dst *v1beta1.ExporterComponentSpec) { //nolint
	dst.ComponentSpec = src.ComponentSpec
	dst.Port = src.Port
	dst.Path = src.Path

	Convert_v1alpha3_HarborExporterCacheSpec_To_v1beta1_HarborExporterCacheSpec(&src.Cache, &dst.Cache)
}

func Convert_v1alpha3_HarborExporterCacheSpec_To_v1beta1_HarborExporterCacheSpec(src *HarborExporterCacheSpec, dst *v1beta1.HarborExporterCacheSpec) { //nolint
	dst.Duration = src.Duration
	dst.CleanInterval = src.CleanInterval
}

func Convert_v1alpha3_TrivyComponentSpec_To_v1beta1_TrivyComponentSpec(src *TrivyComponentSpec, dst *v1beta1.TrivyComponentSpec) { //nolint
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = v1beta1.CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
	dst.GithubTokenRef = src.GithubTokenRef
	dst.SkipUpdate = src.SkipUpdate

	Convert_v1alpha3_HarborStorageTrivyStorageSpec_To_v1beta1_HarborStorageTrivyStorageSpec(&src.Storage, &dst.Storage)
}

func Convert_v1alpha3_HarborStorageTrivyStorageSpec_To_v1beta1_HarborStorageTrivyStorageSpec(src *HarborStorageTrivyStorageSpec, dst *v1beta1.HarborStorageTrivyStorageSpec) { //nolint
	if src.CachePersistentVolume != nil {
		dst.CachePersistentVolume = &v1beta1.HarborStoragePersistentVolumeSpec{}
		dst.CachePersistentVolume.Prefix = src.CachePersistentVolume.Prefix
		dst.CachePersistentVolume.PersistentVolumeClaimVolumeSource = src.CachePersistentVolume.PersistentVolumeClaimVolumeSource
	}

	if dst.ReportsPersistentVolume != nil {
		dst.ReportsPersistentVolume = &v1beta1.HarborStoragePersistentVolumeSpec{}
		dst.CachePersistentVolume.Prefix = src.CachePersistentVolume.Prefix
		dst.CachePersistentVolume.PersistentVolumeClaimVolumeSource = src.CachePersistentVolume.PersistentVolumeClaimVolumeSource
	}
}

func Convert_v1alpha3_NotaryComponentSpec_To_v1beta1_NotaryComponentSpec(src *NotaryComponentSpec, dst *v1beta1.NotaryComponentSpec) { //nolint
	dst.Server = src.Server
	dst.Signer = src.Signer
	dst.MigrationEnabled = src.MigrationEnabled
}

func Convert_v1alpha3_Cache_To_v1beta1_Cache(src *Cache, dst *v1beta1.Cache) { //nolint
	if src.RedisSpec != nil {
		dst.Kind = v1beta1.KindCacheRedisFailover
		dst.Spec = &v1beta1.CacheSpec{}
		Convert_v1alpha3_RedisSpec_To_v1beta1_CacheSpec(src.RedisSpec, dst.Spec)
	}
}

func Convert_v1alpha3_RedisSpec_To_v1beta1_CacheSpec(src *RedisSpec, dst *v1beta1.CacheSpec) { //nolint
	dst.RedisFailover = &v1beta1.RedisFailoverSpec{}

	if src.Server != nil {
		dst.RedisFailover.Server = &v1beta1.RedisServer{
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
		dst.RedisFailover.Sentinel = &v1beta1.RedisSentinel{
			Replicas: src.Sentinel.Replicas,
		}
	}

	dst.RedisFailover.OperatorVersion = "1.0.0"
	dst.RedisFailover.ImageSpec = src.ImageSpec
}

func Convert_v1alpha3_HarborStorageImageChartStorageSpec_To_v1beta1_Storage(src *HarborStorageImageChartStorageSpec, dst *v1beta1.Storage) { //nolint
	if src.FileSystem != nil {
		dst.Kind = v1beta1.KindStorageFileSystem
		dst.Spec.FileSystem = &v1beta1.FileSystemSpec{
			HarborStorageImageChartStorageFileSystemSpec: v1beta1.HarborStorageImageChartStorageFileSystemSpec{
				RegistryPersistentVolume: v1beta1.HarborStorageRegistryPersistentVolumeSpec{
					HarborStoragePersistentVolumeSpec: v1beta1.HarborStoragePersistentVolumeSpec{
						PersistentVolumeClaimVolumeSource: src.FileSystem.RegistryPersistentVolume.PersistentVolumeClaimVolumeSource,
						Prefix:                            src.FileSystem.RegistryPersistentVolume.Prefix,
					},
					MaxThreads: src.FileSystem.RegistryPersistentVolume.MaxThreads,
				},
			},
		}

		if src.FileSystem.ChartPersistentVolume != nil {
			dst.Spec.FileSystem.ChartPersistentVolume = &v1beta1.HarborStoragePersistentVolumeSpec{
				PersistentVolumeClaimVolumeSource: src.FileSystem.ChartPersistentVolume.PersistentVolumeClaimVolumeSource,
				Prefix:                            src.FileSystem.ChartPersistentVolume.Prefix,
			}
		}
	}

	if src.S3 != nil {
		dst.Kind = v1beta1.KindStorageS3
		dst.Spec.S3 = &v1beta1.S3Spec{
			HarborStorageImageChartStorageS3Spec: v1beta1.HarborStorageImageChartStorageS3Spec{
				RegistryStorageDriverS3Spec: v1beta1.RegistryStorageDriverS3Spec{
					AccessKey:      src.S3.AccessKey,
					SecretKeyRef:   src.S3.SecretKeyRef,
					Region:         src.S3.Region,
					RegionEndpoint: src.S3.RegionEndpoint,
					Bucket:         src.S3.Bucket,
					RootDirectory:  src.S3.RootDirectory,
					StorageClass:   src.S3.StorageClass,
					KeyID:          src.S3.KeyID,
					Encrypt:        src.S3.Encrypt,
					SkipVerify:     src.S3.SkipVerify,
					CertificateRef: src.S3.CertificateRef,
					Secure:         src.S3.Secure,
					V4Auth:         src.S3.V4Auth,
					ChunkSize:      src.S3.ChunkSize,
				},
			},
		}
	}

	if src.Swift != nil {
		dst.Kind = v1beta1.KindStorageSwift
		dst.Spec.Swift = &v1beta1.SwiftSpec{
			HarborStorageImageChartStorageSwiftSpec: v1beta1.HarborStorageImageChartStorageSwiftSpec{
				RegistryStorageDriverSwiftSpec: v1beta1.RegistryStorageDriverSwiftSpec{
					AuthURL:            src.Swift.AuthURL,
					Username:           src.Swift.Username,
					PasswordRef:        src.Swift.PasswordRef,
					Region:             src.Swift.Region,
					Container:          src.Swift.Container,
					Tenant:             src.Swift.Tenant,
					TenantID:           src.Swift.TenantID,
					Domain:             src.Swift.Domain,
					DomainID:           src.Swift.DomainID,
					TrustID:            src.Swift.TrustID,
					InsecureSkipVerify: src.Swift.InsecureSkipVerify,
					ChunkSize:          src.Swift.ChunkSize,
					Prefix:             src.Swift.Prefix,
					SecretKeyRef:       src.Swift.SecretKeyRef,
					AccessKey:          src.Swift.AccessKey,
					AuthVersion:        src.Swift.AuthVersion,
					EndpointType:       src.Swift.EndpointType,
				},
			},
		}
	}
}

func Convert_v1alpha3_Storage_To_v1beta1_Storage(src *Storage, dst *v1beta1.Storage) { //nolint
	if src.MinIOSpec != nil {
		dst.Kind = v1beta1.KindStorageMinIO
		dst.Spec.MinIO = &v1beta1.MinIOSpec{}
		Convert_v1alpha3_MinIOSpec_to_v1beta1_MinIOSpec(src.MinIOSpec, dst.Spec.MinIO)
	}
}

func Convert_v1alpha3_MinIOSpec_to_v1beta1_MinIOSpec(src *MinIOSpec, dst *v1beta1.MinIOSpec) { //nolint
	dst.OperatorVersion = "4.0.6"

	dst.SecretRef = src.SecretRef
	dst.VolumeClaimTemplate = src.VolumeClaimTemplate
	dst.VolumesPerServer = src.VolumesPerServer

	dst.ImageSpec = src.ImageSpec

	dst.Replicas = src.Replicas
	dst.Resources = src.Resources

	dst.Redirect.Enable = src.Redirect.Enable
	if src.Redirect.Expose != nil {
		dst.Redirect.Expose = &v1beta1.HarborExposeComponentSpec{}
		Convert_v1alpha3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src.Redirect.Expose, dst.Redirect.Expose)
	}

	if src.MinIOClientSpec != nil {
		dst.MinIOClientSpec = &v1beta1.MinIOClientSpec{ImageSpec: src.MinIOClientSpec.ImageSpec}
	}
}

func Convert_v1alpha3_HarborExposeSpec_To_v1beta1_HarborExposeSpec(src *HarborExposeSpec, dst *v1beta1.HarborExposeSpec) { //nolint
	Convert_v1alpha3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(&src.Core, &dst.Core)

	if src.Notary != nil {
		dst.Notary = &v1beta1.HarborExposeComponentSpec{}
		Convert_v1alpha3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src.Notary, dst.Notary)
	}
}

func Convert_v1alpha3_HarborExposeComponentSpec_To_v1beta1_HarborExposeComponentSpec(src *HarborExposeComponentSpec, dst *v1beta1.HarborExposeComponentSpec) { //nolint
	if src.Ingress != nil {
		dst.Ingress = &v1beta1.HarborExposeIngressSpec{}
		Convert_v1alpha3_HarborExposeIngressSpec_To_v1beta1_HarborExposeIngressSpec(src.Ingress, dst.Ingress)
	}

	if src.TLS != nil {
		dst.TLS = src.TLS
	}
}

func Convert_v1alpha3_HarborExposeIngressSpec_To_v1beta1_HarborExposeIngressSpec(src *HarborExposeIngressSpec, dst *v1beta1.HarborExposeIngressSpec) { //nolint
	dst.Host = src.Host
	dst.Controller = src.Controller
	dst.Annotations = src.Annotations
}

func Convert_v1alpha3_HarborDatabaseSpec_To_v1beta1_Database(src *HarborDatabaseSpec, dst *v1beta1.Database) { //nolint
	dst.Kind = v1beta1.KindDatabasePostgreSQL
	dst.Spec = v1beta1.DatabaseSpec{
		PostgreSQL: &v1beta1.PostgreSQLSpec{
			HarborDatabaseSpec: v1beta1.HarborDatabaseSpec{
				PostgresCredentials: src.PostgresCredentials,
				Hosts:               src.Hosts,
				SSLMode:             src.SSLMode,
				Prefix:              src.Prefix,
			},
		},
	}
}

func Convert_v1alpha3_Database_To_v1beta1_Database(src *Database, dst *v1beta1.Database) { //nolint
	if src.PostgresSQLSpec != nil {
		dst.Kind = v1beta1.KindDatabaseZlandoPostgreSQL
		dst.Spec.ZlandoPostgreSQL = &v1beta1.ZlandoPostgreSQLSpec{}
		Convert_v1alpha3_PostgresSQLSpec_To_v1beta1_ZlandoPostgresSQLSpec(src.PostgresSQLSpec, dst.Spec.ZlandoPostgreSQL)
	}
}

func Convert_v1alpha3_PostgresSQLSpec_To_v1beta1_ZlandoPostgresSQLSpec(src *PostgresSQLSpec, dst *v1beta1.ZlandoPostgreSQLSpec) { //nolint
	dst.OperatorVersion = "1.5.0"
	dst.Storage = src.Storage
	dst.Resources = src.Resources
	dst.Replicas = src.Replicas

	dst.ImageSpec = src.ImageSpec
	dst.StorageClassName = src.StorageClassName
}

func Convert_v1alpha3_HarborClusterStatus_To_v1beta1_HarborClusterStatus(src *HarborClusterStatus, dst *v1beta1.HarborClusterStatus) { //nolint
	dst.Operator = src.Operator
	dst.Status = v1beta1.ClusterStatus(src.Status)
	dst.ObservedGeneration = src.ObservedGeneration
	dst.Revision = src.Revision

	dst.Conditions = func() []v1beta1.HarborClusterCondition {
		conditions := make([]v1beta1.HarborClusterCondition, 0, len(src.Conditions))
		for _, cond := range src.Conditions {
			conditions = append(conditions, v1beta1.HarborClusterCondition{
				Type:               v1beta1.HarborClusterConditionType(cond.Type),
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

func Convert_v1beta1_HarborClusterSpec_To_v1alpha3_HarborClusterSpec(src *v1beta1.HarborClusterSpec, dst *HarborClusterSpec) { //nolint
	if src.Cache.Kind == v1beta1.KindCacheRedis && src.Cache.Spec.Redis != nil {
		if dst.Redis == nil {
			dst.Redis = &ExternalRedisSpec{}
		}

		Convert_v1beta1_ExternalRedisSpec_To_v1alpha3_ExternalRedisSpec(src.Cache.Spec.Redis, dst.Redis)
	} else if src.Cache.Kind == v1beta1.KindCacheRedisFailover {
		if dst.InClusterCache == nil {
			dst.InClusterCache = &Cache{}
		}

		Convert_v1beta1_Cache_To_v1alpha3_Cache(&src.Cache, dst.InClusterCache)
	}

	if dst.ImageChartStorage == nil {
		dst.ImageChartStorage = &HarborStorageImageChartStorageSpec{}
	}

	switch src.Storage.Kind {
	case v1beta1.KindStorageFileSystem:
		if src.Storage.Spec.FileSystem != nil {
			Convert_v1beta1_FileSystemSpec_To_v1alpha3_HarborStorageImageChartStorage(src.Storage.Spec.FileSystem, dst.ImageChartStorage)
		}
	case v1beta1.KindStorageS3:
		if src.Storage.Spec.S3 != nil {
			Convert_v1beta1_S3Spec_To_v1alpha3_HarborStorageImageChartStorage(src.Storage.Spec.S3, dst.ImageChartStorage)
		}
	case v1beta1.KindStorageSwift:
		if src.Storage.Spec.Swift != nil {
			Convert_v1beta1_SwiftSpec_To_v1alpha3_HarborStorageImageChartStorage(src.Storage.Spec.Swift, dst.ImageChartStorage)
		}
	case v1beta1.KindStorageMinIO:
		if dst.InClusterStorage == nil {
			dst.InClusterStorage = &Storage{}
		}

		Convert_v1beta1_Storage_To_v1alpha3_Storage(&src.Storage, dst.InClusterStorage)
	}

	if src.Database.Kind == v1beta1.KindDatabasePostgreSQL && src.Database.Spec.PostgreSQL != nil {
		if dst.Database == nil {
			dst.Database = &HarborDatabaseSpec{}
		}

		Convert_v1beta1_PostgreSQLSpec_To_v1alpha3_HarborDatabaseSpec(src.Database.Spec.PostgreSQL, dst.Database)
	} else if src.Database.Kind == v1beta1.KindDatabaseZlandoPostgreSQL {
		if dst.InClusterDatabase == nil {
			dst.InClusterDatabase = &Database{}
		}

		Convert_v1beta1_Database_To_v1alpha3_Database(&src.Database, dst.InClusterDatabase)
	}

	Convert_v1beta1_EmbeddedHarborSpec_To_v1alpha3_HarborSpec(&src.EmbeddedHarborSpec, &dst.HarborSpec)
}

func Convert_v1beta1_S3Spec_To_v1alpha3_HarborStorageImageChartStorage(src *v1beta1.S3Spec, dst *HarborStorageImageChartStorageSpec) { //nolint
	dst.S3 = &HarborStorageImageChartStorageS3Spec{
		RegistryStorageDriverS3Spec: RegistryStorageDriverS3Spec{
			AccessKey:      src.AccessKey,
			SecretKeyRef:   src.SecretKeyRef,
			Region:         src.Region,
			RegionEndpoint: src.RegionEndpoint,
			Bucket:         src.Bucket,
			RootDirectory:  src.RootDirectory,
			StorageClass:   src.StorageClass,
			KeyID:          src.KeyID,
			Encrypt:        src.Encrypt,
			SkipVerify:     src.SkipVerify,
			CertificateRef: src.CertificateRef,
			Secure:         src.Secure,
			V4Auth:         src.V4Auth,
			ChunkSize:      src.ChunkSize,
		},
	}
}

func Convert_v1beta1_SwiftSpec_To_v1alpha3_HarborStorageImageChartStorage(src *v1beta1.SwiftSpec, dst *HarborStorageImageChartStorageSpec) { //nolint
	dst.Swift = &HarborStorageImageChartStorageSwiftSpec{
		RegistryStorageDriverSwiftSpec: RegistryStorageDriverSwiftSpec{
			AuthURL:            src.AuthURL,
			Username:           src.Username,
			PasswordRef:        src.PasswordRef,
			Region:             src.Region,
			Container:          src.Container,
			Tenant:             src.Tenant,
			TenantID:           src.TenantID,
			Domain:             src.Domain,
			DomainID:           src.DomainID,
			TrustID:            src.TrustID,
			InsecureSkipVerify: src.InsecureSkipVerify,
			ChunkSize:          src.ChunkSize,
			Prefix:             src.Prefix,
			SecretKeyRef:       src.SecretKeyRef,
			AccessKey:          src.AccessKey,
			AuthVersion:        src.AuthVersion,
			EndpointType:       src.EndpointType,
		},
	}
}

func Convert_v1beta1_EmbeddedHarborSpec_To_v1alpha3_HarborSpec(src *v1beta1.EmbeddedHarborSpec, dst *HarborSpec) { //nolint
	dst.ExternalURL = src.ExternalURL
	dst.InternalTLS = HarborInternalTLSSpec{
		Enabled: src.InternalTLS.Enabled,
	}

	dst.LogLevel = src.LogLevel
	dst.HarborAdminPasswordRef = src.HarborAdminPasswordRef
	dst.UpdateStrategyType = src.UpdateStrategyType
	dst.Version = src.Version
	dst.ImageSource = src.ImageSource.DeepCopy()

	if src.Proxy != nil {
		dst.Proxy = &HarborProxySpec{
			ProxySpec:  src.Proxy.ProxySpec,
			Components: src.Proxy.Components,
		}
	}

	Convert_v1beta1_HarborExposeSpec_To_v1alpha3_HarborExposeSpec(&src.Expose, &dst.Expose)

	Convert_v1beta1_EmbeddedHarborComponentsSpec_To_v1alpha3_HarborComponentSpec(&src.EmbeddedHarborComponentsSpec, &dst.HarborComponentsSpec)
}

func Convert_v1beta1_EmbeddedHarborComponentsSpec_To_v1alpha3_HarborComponentSpec(src *v1beta1.EmbeddedHarborComponentsSpec, dst *HarborComponentsSpec) { //nolint
	Convert_v1beta1_CoreComponentSpec_To_v1alpha3_CoreComponentSpec(&src.Core, &dst.Core)

	Convert_v1beta1_RegistryComponentSpec_To_v1alpha3_RegistryComponentSpec(&src.Registry, &dst.Registry)

	Convert_v1beta1_JobServiceComponentSpec_To_v1alpha3_JobServiceComponentSpec(&src.JobService, &dst.JobService)

	if src.ChartMuseum != nil {
		dst.ChartMuseum = &ChartMuseumComponentSpec{}
		Convert_v1beta1_ChartMuseumComponentSpec_To_v1alpha3_ChartMuseumComponentSpec(src.ChartMuseum, dst.ChartMuseum)
	}

	if src.Notary != nil {
		dst.Notary = &NotaryComponentSpec{}
		Convert_v1beta1_NotaryComponentSpec_To_v1alpha3_NotaryComponentSpec(src.Notary, dst.Notary)
	}

	if src.Trivy != nil {
		dst.Trivy = &TrivyComponentSpec{}
		Convert_v1beta1_TrivyComponentSpec_To_v1alpha3_TrivyComponentSpec(src.Trivy, dst.Trivy)
	}

	if src.Exporter != nil {
		dst.Exporter = &ExporterComponentSpec{}
		Convert_v1beta1_ExporterComponentSpec_To_v1alpha3_ExporterComponentSpec(src.Exporter, dst.Exporter)
	}
}

func Convert_v1beta1_CoreComponentSpec_To_v1alpha3_CoreComponentSpec(src *v1beta1.CoreComponentSpec, dst *CoreComponentSpec) { //nolint
	dst.CertificateInjection = CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
	dst.Metrics = src.Metrics
	dst.ComponentSpec = src.ComponentSpec
	dst.TokenIssuer = src.TokenIssuer
}

func Convert_v1beta1_RegistryComponentSpec_To_v1alpha3_RegistryComponentSpec(src *v1beta1.RegistryComponentSpec, dst *RegistryComponentSpec) { //nolint
	dst.CertificateInjection = CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
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

func Convert_v1beta1_JobServiceComponentSpec_To_v1alpha3_JobServiceComponentSpec(src *v1beta1.JobServiceComponentSpec, dst *JobServiceComponentSpec) { //nolint
	dst.WorkerCount = src.WorkerCount
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
}

func Convert_v1beta1_ChartMuseumComponentSpec_To_v1alpha3_ChartMuseumComponentSpec(src *v1beta1.ChartMuseumComponentSpec, dst *ChartMuseumComponentSpec) { //nolint
	dst.AbsoluteURL = src.AbsoluteURL
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{CertificateRefs: src.CertificateInjection.CertificateRefs}
}

func Convert_v1beta1_ExporterComponentSpec_To_v1alpha3_ExporterComponentSpec(src *v1beta1.ExporterComponentSpec, dst *ExporterComponentSpec) { //nolint
	dst.ComponentSpec = src.ComponentSpec
	dst.Port = src.Port
	dst.Path = src.Path

	Convert_v1beta1_HarborExporterCacheSpec_To_v1alpha3_HarborExporterCacheSpec(&src.Cache, &dst.Cache)
}

func Convert_v1beta1_HarborExporterCacheSpec_To_v1alpha3_HarborExporterCacheSpec(src *v1beta1.HarborExporterCacheSpec, dst *HarborExporterCacheSpec) { //nolint
	dst.Duration = src.Duration
	dst.CleanInterval = src.CleanInterval
}

func Convert_v1beta1_TrivyComponentSpec_To_v1alpha3_TrivyComponentSpec(src *v1beta1.TrivyComponentSpec, dst *TrivyComponentSpec) { //nolint
	dst.ComponentSpec = src.ComponentSpec
	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateInjection.CertificateRefs,
	}
	dst.GithubTokenRef = src.GithubTokenRef
	dst.SkipUpdate = src.SkipUpdate

	Convert_v1beta1_HarborStorageTrivyStorageSpec_To_v1alpha3_HarborStorageTrivyStorageSpec(&src.Storage, &dst.Storage)
}

func Convert_v1beta1_HarborStorageTrivyStorageSpec_To_v1alpha3_HarborStorageTrivyStorageSpec(src *v1beta1.HarborStorageTrivyStorageSpec, dst *HarborStorageTrivyStorageSpec) { //nolint
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

func Convert_v1beta1_NotaryComponentSpec_To_v1alpha3_NotaryComponentSpec(src *v1beta1.NotaryComponentSpec, dst *NotaryComponentSpec) { //nolint
	dst.Server = src.Server
	dst.Signer = src.Signer
	dst.MigrationEnabled = src.MigrationEnabled
}

func Convert_v1beta1_ExternalRedisSpec_To_v1alpha3_ExternalRedisSpec(src *v1beta1.ExternalRedisSpec, dst *ExternalRedisSpec) { //nolint
	dst.RedisCredentials = src.RedisCredentials
	dst.RedisHostSpec = src.RedisHostSpec
}

func Convert_v1beta1_Cache_To_v1alpha3_Cache(src *v1beta1.Cache, dst *Cache) { //nolint
	dst.Kind = v1beta1.KindCacheRedis
	if src.Spec != nil {
		dst.RedisSpec = &RedisSpec{}
		Convert_v1beta1_RedisSpec_To_v1alpha3_CacheSpec(src.Spec, dst.RedisSpec)
	}
}

func Convert_v1beta1_RedisSpec_To_v1alpha3_CacheSpec(src *v1beta1.CacheSpec, dst *RedisSpec) { //nolint
	if src.RedisFailover != nil {
		if src.RedisFailover.Server != nil {
			dst.Server = &RedisServer{
				Replicas:         src.RedisFailover.Server.Replicas,
				Resources:        src.RedisFailover.Server.Resources,
				StorageClassName: src.RedisFailover.Server.StorageClassName,
				Storage:          src.RedisFailover.Server.Storage,
			}
		}

		if src.RedisFailover.Sentinel != nil {
			dst.Sentinel = &RedisSentinel{
				Replicas: src.RedisFailover.Sentinel.Replicas,
			}
		}

		dst.ImageSpec = src.RedisFailover.ImageSpec
	}
}

func Convert_v1beta1_FileSystemSpec_To_v1alpha3_HarborStorageImageChartStorage(src *v1beta1.FileSystemSpec, dst *HarborStorageImageChartStorageSpec) { //nolint
	dst.FileSystem = &HarborStorageImageChartStorageFileSystemSpec{}

	if src.ChartPersistentVolume != nil {
		dst.FileSystem.ChartPersistentVolume = &HarborStoragePersistentVolumeSpec{
			PersistentVolumeClaimVolumeSource: src.ChartPersistentVolume.PersistentVolumeClaimVolumeSource,
			Prefix:                            src.ChartPersistentVolume.Prefix,
		}
	}

	dst.FileSystem.RegistryPersistentVolume = HarborStorageRegistryPersistentVolumeSpec{
		HarborStoragePersistentVolumeSpec: HarborStoragePersistentVolumeSpec{
			PersistentVolumeClaimVolumeSource: src.RegistryPersistentVolume.PersistentVolumeClaimVolumeSource,
			Prefix:                            src.RegistryPersistentVolume.Prefix,
		},
		MaxThreads: src.RegistryPersistentVolume.MaxThreads,
	}
}

func Convert_v1beta1_Storage_To_v1alpha3_Storage(src *v1beta1.Storage, dst *Storage) { //nolint
	dst.Kind = src.Kind

	if src.Spec.MinIO != nil {
		dst.MinIOSpec = &MinIOSpec{}
		Convert_v1beta1_MinIOSpec_To_v1alpha3_MinIOSpec(src.Spec.MinIO, dst.MinIOSpec)
	}
}

func Convert_v1beta1_MinIOSpec_To_v1alpha3_MinIOSpec(src *v1beta1.MinIOSpec, dst *MinIOSpec) { //nolint
	dst.SecretRef = src.SecretRef
	dst.VolumeClaimTemplate = src.VolumeClaimTemplate
	dst.VolumesPerServer = src.VolumesPerServer

	dst.ImageSpec = src.ImageSpec

	dst.Replicas = src.Replicas
	dst.Resources = src.Resources

	if src.Redirect != nil {
		dst.Redirect.Enable = src.Redirect.Enable
		if src.Redirect.Expose != nil {
			dst.Redirect.Expose = &HarborExposeComponentSpec{}
			Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src.Redirect.Expose, dst.Redirect.Expose)
		}
	}

	if src.MinIOClientSpec != nil {
		dst.MinIOClientSpec = &MinIOClientSpec{ImageSpec: src.MinIOClientSpec.ImageSpec}
	}
}

func Convert_v1beta1_HarborExposeSpec_To_v1alpha3_HarborExposeSpec(src *v1beta1.HarborExposeSpec, dst *HarborExposeSpec) { //nolint
	Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(&src.Core, &dst.Core)

	if src.Notary != nil {
		dst.Notary = &HarborExposeComponentSpec{}
		Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src.Notary, dst.Notary)
	}
}

func Convert_v1beta1_HarborExposeComponentSpec_To_v1alpha3_HarborExposeComponentSpec(src *v1beta1.HarborExposeComponentSpec, dst *HarborExposeComponentSpec) { //nolint
	if src.Ingress != nil {
		dst.Ingress = &HarborExposeIngressSpec{}
		Convert_v1beta1_HarborExposeIngressSpec_To_v1alpha3_HarborExposeIngressSpec(src.Ingress, dst.Ingress)
	}

	if src.TLS != nil {
		dst.TLS = src.TLS
	}
}

func Convert_v1beta1_HarborExposeIngressSpec_To_v1alpha3_HarborExposeIngressSpec(src *v1beta1.HarborExposeIngressSpec, dst *HarborExposeIngressSpec) { //nolint
	dst.Host = src.Host
	dst.Controller = src.Controller
	dst.Annotations = src.Annotations
}

func Convert_v1beta1_PostgreSQLSpec_To_v1alpha3_HarborDatabaseSpec(src *v1beta1.PostgreSQLSpec, dst *HarborDatabaseSpec) { //nolint
	dst.PostgresCredentials = src.PostgresCredentials
	dst.Hosts = src.Hosts
	dst.Prefix = src.Prefix
	dst.SSLMode = src.SSLMode
}

func Convert_v1beta1_Database_To_v1alpha3_Database(src *v1beta1.Database, dst *Database) { //nolint
	dst.Kind = v1beta1.KindDatabasePostgreSQL

	if src.Spec.ZlandoPostgreSQL != nil {
		dst.PostgresSQLSpec = &PostgresSQLSpec{}
		Convert_v1beta1_ZlandoPostgreSQLSpec_To_v1alpha3_PostgresSQLSpec(src.Spec.ZlandoPostgreSQL, dst.PostgresSQLSpec)
	}
}

func Convert_v1beta1_ZlandoPostgreSQLSpec_To_v1alpha3_PostgresSQLSpec(src *v1beta1.ZlandoPostgreSQLSpec, dst *PostgresSQLSpec) { //nolint
	dst.Storage = src.Storage
	dst.Resources = src.Resources
	dst.Replicas = src.Replicas

	dst.ImageSpec = src.ImageSpec
	dst.StorageClassName = src.StorageClassName
}

func Convert_v1beta1_HarborClusterStatus_To_v1alpha3_HarborClusterStatus(src *v1beta1.HarborClusterStatus, dst *HarborClusterStatus) { //nolint
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
