package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Registry) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Registry)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_RegistrySpec_To_v1alpha3_RegistrySpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *Registry) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Registry)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_RegistrySpec_To_v1beta1_RegistrySpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_RegistrySpec_To_v1alpha3_RegistrySpec(src *RegistrySpec, dst *v1alpha3.RegistrySpec) {
	dst.ComponentSpec = src.ComponentSpec
	dst.Proxy = src.Proxy

	dst.CertificateInjection = v1alpha3.CertificateInjection{
		CertificateRefs: src.CertificateRefs,
	}

	Convert_v1beta1_RegistryConfig01_To_v1alpha3_RegistryConfig01(&src.RegistryConfig01, &dst.RegistryConfig01)
}

func Convert_v1beta1_RegistryConfig01_To_v1alpha3_RegistryConfig01(src *RegistryConfig01, dst *v1alpha3.RegistryConfig01) {
	Convert_v1beta1_RegistryLogSpec_To_v1alpha3_RegistryLogSpec(&src.Log, &dst.Log)

	Convert_v1beta1_RegistryHTTPSpec_To_v1alpha3_RegistryHTTPSpec(&src.HTTP, &dst.HTTP)

	Convert_v1beta1_RegistryHealthSpec_To_v1alpha3_RegistryHealthSpec(&src.Health, &dst.Health)

	Convert_v1beta1_RegistryNotificationsSpec_To_v1alpha3_RegistryNotificationsSpec(&src.Notifications, &dst.Notifications)

	Convert_v1beta1_RegistryAuthenticationSpec_To_v1alpha3_RegistryAuthenticationSpec(&src.Authentication, &dst.Authentication)

	Convert_v1beta1_RegistryValidationSpec_To_v1alpha3_RegistryValidationSpec(&src.Validation, &dst.Validation)

	Convert_v1beta1_RegistryCompatibilitySpec_To_v1alpha3_RegistryCompatibilitySpec(&src.Compatibility, &dst.Compatibility)

	Convert_v1beta1_RegistryStorageSpec_To_v1alpha3_RegistryStorageSpec(&src.Storage, &dst.Storage)

	Convert_v1beta1_RegistryMiddlewaresSpec_To_v1alpha3_RegistryMiddlewaresSpec(&src.Middlewares, &dst.Middlewares)

	if src.Redis != nil {
		dst.Redis = &v1alpha3.RegistryRedisSpec{}
		Convert_v1beta1_RegistryRedisSpec_To_v1alpha3_RegistryRedisSpec(src.Redis, dst.Redis)
	}

	dst.Reporting = map[string]string{}
	for key, value := range src.Reporting {
		dst.Reporting[key] = value
	}
}

func Convert_v1beta1_RegistryLogSpec_To_v1alpha3_RegistryLogSpec(src *RegistryLogSpec, dst *v1alpha3.RegistryLogSpec) {
	dst.Level = src.Level
	dst.Formatter = src.Formatter

	dst.AccessLog = v1alpha3.RegistryAccessLogSpec{
		Disabled: src.AccessLog.Disabled,
	}

	dst.Fields = map[string]string{}
	for key, value := range src.Fields {
		dst.Fields[key] = value
	}

	dst.Hooks = make([]v1alpha3.RegistryLogHookSpec, len(src.Hooks))
	for _, hook := range src.Hooks {
		dst.Hooks = append(dst.Hooks, v1alpha3.RegistryLogHookSpec{
			Type:       hook.Type,
			Levels:     hook.Levels,
			OptionsRef: hook.OptionsRef,
		})
	}
}

func Convert_v1beta1_RegistryHTTPSpec_To_v1alpha3_RegistryHTTPSpec(src *RegistryHTTPSpec, dst *v1alpha3.RegistryHTTPSpec) {
	dst.SecretRef = src.SecretRef
	dst.Host = src.Host
	dst.Net = src.Net
	dst.Prefix = src.Prefix
	dst.DrainTimeout = src.DrainTimeout
	dst.RelativeURLs = src.RelativeURLs
	dst.TLS = src.TLS

	dst.HTTP2 = v1alpha3.RegistryHTTPHTTP2Spec{
		Disabled: src.HTTP2.Disabled,
	}

	if src.Debug != nil {
		dst.Debug = &v1alpha3.RegistryHTTPDebugSpec{}
		Convert_v1beta1_RegistryHTTPDebugSpec_To_v1alpha3_RegistryHTTPDebugSpec(src.Debug, dst.Debug)
	}
}

func Convert_v1beta1_RegistryHTTPDebugSpec_To_v1alpha3_RegistryHTTPDebugSpec(src *RegistryHTTPDebugSpec, dst *v1alpha3.RegistryHTTPDebugSpec) {
	dst.Port = src.Port
	dst.Prometheus = v1alpha3.RegistryHTTPDebugPrometheusSpec{
		Enabled: src.Prometheus.Enabled,
		Path:    src.Prometheus.Path,
	}
}

func Convert_v1beta1_RegistryHealthSpec_To_v1alpha3_RegistryHealthSpec(src *RegistryHealthSpec, dst *v1alpha3.RegistryHealthSpec) {
	dst.StorageDriver = v1alpha3.RegistryHealthStorageDriverSpec{
		Enabled:   src.StorageDriver.Enabled,
		Threshold: src.StorageDriver.Threshold,
		Interval:  src.StorageDriver.Interval,
	}

	if len(src.HTTP) > 0 {
		dst.HTTP = make([]v1alpha3.RegistryHealthHTTPSpec, len(src.HTTP))
		for _, http := range src.HTTP {
			spec := v1alpha3.RegistryHealthHTTPSpec{}
			Convert_v1beta1_RegistryHealthHTTPSpec_To_v1alpha3_RegistryHealthHTTPSpec(&http, &spec)
			dst.HTTP = append(dst.HTTP, spec)
		}
	}

	if len(src.File) > 0 {
		dst.File = make([]v1alpha3.RegistryHealthFileSpec, len(src.File))
		for _, file := range src.File {
			spec := v1alpha3.RegistryHealthFileSpec{}
			Convert_v1beta1_RegistryHealthFileSpec_To_v1alpha3_RegistryHealthFileSpec(&file, &spec)
			dst.File = append(dst.File, spec)
		}
	}

	if len(src.TCP) > 0 {
		dst.TCP = make([]v1alpha3.RegistryHealthTCPSpec, len(src.TCP))
		for _, TCP := range src.TCP {
			spec := v1alpha3.RegistryHealthTCPSpec{}
			Convert_v1beta1_RegistryHealthTCPSpec_To_v1alpha3_RegistryHealthTCPSpec(&TCP, &spec)
			dst.TCP = append(dst.TCP, spec)
		}
	}
}

func Convert_v1beta1_RegistryHealthHTTPSpec_To_v1alpha3_RegistryHealthHTTPSpec(src *RegistryHealthHTTPSpec, dst *v1alpha3.RegistryHealthHTTPSpec) {
	dst.URI = src.URI
	dst.Interval = src.Interval
	dst.Threshold = src.Threshold
	dst.StatusCode = src.StatusCode
	dst.Timeout = src.Timeout

	dst.Headers = src.Headers
}

func Convert_v1beta1_RegistryHealthFileSpec_To_v1alpha3_RegistryHealthFileSpec(src *RegistryHealthFileSpec, dst *v1alpha3.RegistryHealthFileSpec) {
	dst.File = src.File
	dst.Interval = src.Interval
}

func Convert_v1beta1_RegistryHealthTCPSpec_To_v1alpha3_RegistryHealthTCPSpec(src *RegistryHealthTCPSpec, dst *v1alpha3.RegistryHealthTCPSpec) {
	dst.Interval = src.Interval
	dst.Timeout = src.Timeout
	dst.Threshold = src.Threshold
	dst.Address = src.Address
}

func Convert_v1beta1_RegistryNotificationsSpec_To_v1alpha3_RegistryNotificationsSpec(src *RegistryNotificationsSpec, dst *v1alpha3.RegistryNotificationsSpec) {
	dst.Endpoints = make([]v1alpha3.RegistryNotificationEndpointSpec, len(src.Endpoints))
	for _, ep := range src.Endpoints {
		dst.Endpoints = append(dst.Endpoints, v1alpha3.RegistryNotificationEndpointSpec{
			Name:              ep.Name,
			URL:               ep.URL,
			Disabled:          ep.Disabled,
			Threshold:         ep.Threshold,
			Timeout:           ep.Timeout,
			Backoff:           ep.Backoff,
			Headers:           ep.Headers,
			IgnoredMediaTypes: ep.IgnoredMediaTypes,
			Ignore: v1alpha3.RegistryNotificationEndpointIgnoreSpec{
				MediaTypes: ep.Ignore.MediaTypes,
				Actions:    ep.Ignore.Actions,
			},
		})
	}

	dst.Events = v1alpha3.RegistryNotificationEventsSpec{
		IncludeReferences: src.Events.IncludeReferences,
	}
}

func Convert_v1beta1_RegistryAuthenticationSpec_To_v1alpha3_RegistryAuthenticationSpec(src *RegistryAuthenticationSpec, dst *v1alpha3.RegistryAuthenticationSpec) {
	if src.Silly != nil {
		dst.Silly = &v1alpha3.RegistryAuthenticationSillySpec{
			Realm:   src.Silly.Realm,
			Service: src.Silly.Service,
		}
	}

	if src.Token != nil {
		dst.Token = &v1alpha3.RegistryAuthenticationTokenSpec{
			Realm:          src.Token.Realm,
			Service:        src.Token.Service,
			Issuer:         src.Token.Issuer,
			CertificateRef: src.Token.CertificateRef,
			AutoRedirect:   src.Token.AutoRedirect,
		}
	}

	if src.HTPasswd != nil {
		dst.HTPasswd = &v1alpha3.RegistryAuthenticationHTPasswdSpec{
			Realm:     src.HTPasswd.Realm,
			SecretRef: src.HTPasswd.SecretRef,
		}
	}
}

func Convert_v1beta1_RegistryValidationSpec_To_v1alpha3_RegistryValidationSpec(src *RegistryValidationSpec, dst *v1alpha3.RegistryValidationSpec) {
	dst.Disabled = src.Disabled
	dst.Manifests = v1alpha3.RegistryValidationManifestSpec{
		URLs: v1alpha3.RegistryValidationManifestURLsSpec{
			Allow: src.Manifests.URLs.Allow,
			Deny:  src.Manifests.URLs.Deny,
		},
	}
}

func Convert_v1beta1_RegistryCompatibilitySpec_To_v1alpha3_RegistryCompatibilitySpec(src *RegistryCompatibilitySpec, dst *v1alpha3.RegistryCompatibilitySpec) {
	dst.Schema1 = v1alpha3.RegistryCompatibilitySchemaSpec{
		Enabled:        src.Schema1.Enabled,
		CertificateRef: src.Schema1.CertificateRef,
	}
}

func Convert_v1beta1_RegistryStorageSpec_To_v1alpha3_RegistryStorageSpec(src *RegistryStorageSpec, dst *v1alpha3.RegistryStorageSpec) {
	dst.Cache = v1alpha3.RegistryStorageCacheSpec{
		Blobdescriptor: src.Cache.Blobdescriptor,
	}

	dst.Maintenance = v1alpha3.RegistryStorageMaintenanceSpec{
		UploadPurging: v1alpha3.RegistryStorageMaintenanceUploadPurgingSpec{
			Enabled:  src.Maintenance.UploadPurging.Enabled,
			DryRun:   src.Maintenance.UploadPurging.DryRun,
			Age:      src.Maintenance.UploadPurging.Age,
			Interval: src.Maintenance.UploadPurging.Interval,
		},
		ReadOnly: v1alpha3.RegistryStorageMaintenanceReadOnlySpec{
			Enabled: src.Maintenance.ReadOnly.Enabled,
		},
	}

	dst.Delete = v1alpha3.RegistryStorageDeleteSpec{
		Enabled: src.Delete.Enabled,
	}

	dst.Redirect = v1alpha3.RegistryStorageRedirectSpec{
		Disable: src.Redirect.Disable,
	}

	Convert_v1beta1_RegistryStorageDriverSpec_To_v1alpha3_RegistryStorageDriverSpec(&src.Driver, &dst.Driver)
}

func Convert_v1beta1_RegistryStorageDriverSpec_To_v1alpha3_RegistryStorageDriverSpec(src *RegistryStorageDriverSpec, dst *v1alpha3.RegistryStorageDriverSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.RegistryStorageDriverSpec{}
	}

	if src.InMemory != nil {
		dst.InMemory = &v1alpha3.RegistryStorageDriverInmemorySpec{}
	}

	if src.FileSystem != nil {
		dst.FileSystem = &v1alpha3.RegistryStorageDriverFilesystemSpec{
			VolumeSource: src.FileSystem.VolumeSource,
			MaxThreads:   src.FileSystem.MaxThreads,
			Prefix:       src.FileSystem.Prefix,
		}
	}

	if src.S3 != nil {
		dst.S3 = &v1alpha3.RegistryStorageDriverS3Spec{
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
		}
	}

	if src.Swift != nil {
		dst.Swift = &v1alpha3.RegistryStorageDriverSwiftSpec{
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
		}
	}
}

func Convert_v1beta1_RegistryMiddlewaresSpec_To_v1alpha3_RegistryMiddlewaresSpec(src *RegistryMiddlewaresSpec, dst *v1alpha3.RegistryMiddlewaresSpec) {
	if len(src.Storage) > 0 {
		dst.Storage = make([]v1alpha3.RegistryMiddlewareSpec, len(src.Storage))
		for _, storage := range src.Storage {
			spec := v1alpha3.RegistryMiddlewareSpec{}
			Convert_v1beta1_RegistryMiddlewareSpec_To_v1alpha3_RegistryMiddlewareSpec(&storage, &spec)
			dst.Storage = append(dst.Storage, spec)
		}
	}

	if len(src.Registry) > 0 {
		dst.Registry = make([]v1alpha3.RegistryMiddlewareSpec, len(src.Registry))
		for _, registry := range src.Registry {
			spec := v1alpha3.RegistryMiddlewareSpec{}
			Convert_v1beta1_RegistryMiddlewareSpec_To_v1alpha3_RegistryMiddlewareSpec(&registry, &spec)
			dst.Registry = append(dst.Registry, spec)
		}
	}

	if len(src.Repository) > 0 {
		dst.Repository = make([]v1alpha3.RegistryMiddlewareSpec, len(src.Repository))
		for _, repository := range src.Repository {
			spec := v1alpha3.RegistryMiddlewareSpec{}
			Convert_v1beta1_RegistryMiddlewareSpec_To_v1alpha3_RegistryMiddlewareSpec(&repository, &spec)
			dst.Repository = append(dst.Repository, spec)
		}
	}
}

func Convert_v1beta1_RegistryMiddlewareSpec_To_v1alpha3_RegistryMiddlewareSpec(src *RegistryMiddlewareSpec, dst *v1alpha3.RegistryMiddlewareSpec) {
	dst.Name = src.Name
	dst.OptionsRef = src.OptionsRef
}

func Convert_v1beta1_RegistryRedisSpec_To_v1alpha3_RegistryRedisSpec(src *RegistryRedisSpec, dst *v1alpha3.RegistryRedisSpec) {
	dst.RedisConnection = src.RedisConnection
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
	dst.DialTimeout = src.DialTimeout

	dst.Pool = v1alpha3.RegistryRedisPoolSpec{
		MaxIdle:     src.Pool.MaxIdle,
		MaxActive:   src.Pool.MaxActive,
		IdleTimeout: src.Pool.IdleTimeout,
	}
}

func Convert_v1alpha3_RegistrySpec_To_v1beta1_RegistrySpec(src *v1alpha3.RegistrySpec, dst *RegistrySpec) {
	dst.ComponentSpec = src.ComponentSpec
	dst.Proxy = src.Proxy

	dst.CertificateInjection = CertificateInjection{
		CertificateRefs: src.CertificateRefs,
	}

	Convert_v1alpha3_RegistryConfig01_To_v1beta1_RegistryConfig01(&src.RegistryConfig01, &dst.RegistryConfig01)
}

func Convert_v1alpha3_RegistryConfig01_To_v1beta1_RegistryConfig01(src *v1alpha3.RegistryConfig01, dst *RegistryConfig01) {
	Convert_v1alpha3_RegistryLogSpec_To_v1beta1_RegistryLogSpec(&src.Log, &dst.Log)

	Convert_v1alpha3_RegistryHTTPSpec_To_v1beta1_RegistryHTTPSpec(&src.HTTP, &dst.HTTP)

	Convert_v1alpha3_RegistryHealthSpec_To_v1beta1_RegistryHealthSpec(&src.Health, &dst.Health)

	Convert_v1alpha3_RegistryNotificationsSpec_To_v1beta1_RegistryNotificationsSpec(&src.Notifications, &dst.Notifications)

	Convert_v1alpha3_RegistryAuthenticationSpec_To_v1beta1_RegistryAuthenticationSpec(&src.Authentication, &dst.Authentication)

	Convert_v1alpha3_RegistryValidationSpec_To_v1beta1_RegistryValidationSpec(&src.Validation, &dst.Validation)

	Convert_v1alpha3_RegistryCompatibilitySpec_To_v1beta1_RegistryCompatibilitySpec(&src.Compatibility, &dst.Compatibility)

	Convert_v1alpha3_RegistryStorageSpec_To_v1beta1_RegistryStorageSpec(&src.Storage, &dst.Storage)

	Convert_v1alpha3_RegistryMiddlewaresSpec_To_v1beta1_RegistryMiddlewaresSpec(&src.Middlewares, &dst.Middlewares)

	if src.Redis != nil {
		dst.Redis = &RegistryRedisSpec{}
		Convert_v1alpha3_RegistryRedisSpec_To_v1beta1_RegistryRedisSpec(src.Redis, dst.Redis)
	}

	dst.Reporting = map[string]string{}
	for key, value := range src.Reporting {
		dst.Reporting[key] = value
	}
}

func Convert_v1alpha3_RegistryLogSpec_To_v1beta1_RegistryLogSpec(src *v1alpha3.RegistryLogSpec, dst *RegistryLogSpec) {
	dst.Level = src.Level
	dst.Formatter = src.Formatter

	dst.AccessLog = RegistryAccessLogSpec{
		Disabled: src.AccessLog.Disabled,
	}

	dst.Fields = map[string]string{}
	for key, value := range src.Fields {
		dst.Fields[key] = value
	}

	dst.Hooks = make([]RegistryLogHookSpec, len(src.Hooks))
	for _, hook := range src.Hooks {
		dst.Hooks = append(dst.Hooks, RegistryLogHookSpec{
			Type:       hook.Type,
			Levels:     hook.Levels,
			OptionsRef: hook.OptionsRef,
		})
	}
}

func Convert_v1alpha3_RegistryHTTPSpec_To_v1beta1_RegistryHTTPSpec(src *v1alpha3.RegistryHTTPSpec, dst *RegistryHTTPSpec) {
	dst.SecretRef = src.SecretRef
	dst.Host = src.Host
	dst.Net = src.Net
	dst.Prefix = src.Prefix
	dst.DrainTimeout = src.DrainTimeout
	dst.RelativeURLs = src.RelativeURLs
	dst.TLS = src.TLS

	dst.HTTP2 = RegistryHTTPHTTP2Spec{
		Disabled: src.HTTP2.Disabled,
	}

	if src.Debug != nil {
		dst.Debug = &RegistryHTTPDebugSpec{}
		Convert_v1alpha3_RegistryHTTPDebugSpec_To_v1beta1_RegistryHTTPDebugSpec(src.Debug, dst.Debug)
	}
}

func Convert_v1alpha3_RegistryHTTPDebugSpec_To_v1beta1_RegistryHTTPDebugSpec(src *v1alpha3.RegistryHTTPDebugSpec, dst *RegistryHTTPDebugSpec) {
	dst.Port = src.Port
	dst.Prometheus = RegistryHTTPDebugPrometheusSpec{
		Enabled: src.Prometheus.Enabled,
		Path:    src.Prometheus.Path,
	}
}

func Convert_v1alpha3_RegistryHealthSpec_To_v1beta1_RegistryHealthSpec(src *v1alpha3.RegistryHealthSpec, dst *RegistryHealthSpec) {

	dst.StorageDriver = RegistryHealthStorageDriverSpec{
		Enabled:   src.StorageDriver.Enabled,
		Threshold: src.StorageDriver.Threshold,
		Interval:  src.StorageDriver.Interval,
	}

	if len(src.HTTP) > 0 {
		dst.HTTP = make([]RegistryHealthHTTPSpec, len(src.HTTP))
		for _, http := range src.HTTP {
			spec := RegistryHealthHTTPSpec{}
			Convert_v1alpha3_RegistryHealthHTTPSpec_To_v1beta1_RegistryHealthHTTPSpec(&http, &spec)
			dst.HTTP = append(dst.HTTP, spec)
		}
	}

	if len(src.File) > 0 {
		dst.File = make([]RegistryHealthFileSpec, len(src.File))
		for _, file := range src.File {
			spec := RegistryHealthFileSpec{}
			Convert_v1alpha3_RegistryHealthFileSpec_To_v1beta1_RegistryHealthFileSpec(&file, &spec)
			dst.File = append(dst.File, spec)
		}
	}

	if len(src.TCP) > 0 {
		dst.TCP = make([]RegistryHealthTCPSpec, len(src.TCP))
		for _, TCP := range src.TCP {
			spec := RegistryHealthTCPSpec{}
			Convert_v1alpha3_RegistryHealthTCPSpec_To_v1beta1_RegistryHealthTCPSpec(&TCP, &spec)
			dst.TCP = append(dst.TCP, spec)
		}
	}
}

func Convert_v1alpha3_RegistryHealthHTTPSpec_To_v1beta1_RegistryHealthHTTPSpec(src *v1alpha3.RegistryHealthHTTPSpec, dst *RegistryHealthHTTPSpec) {

	dst.URI = src.URI
	dst.Interval = src.Interval
	dst.Threshold = src.Threshold
	dst.StatusCode = src.StatusCode
	dst.Timeout = src.Timeout

	dst.Headers = src.Headers
}

func Convert_v1alpha3_RegistryHealthFileSpec_To_v1beta1_RegistryHealthFileSpec(src *v1alpha3.RegistryHealthFileSpec, dst *RegistryHealthFileSpec) {

	dst.File = src.File
	dst.Interval = src.Interval
}

func Convert_v1alpha3_RegistryHealthTCPSpec_To_v1beta1_RegistryHealthTCPSpec(src *v1alpha3.RegistryHealthTCPSpec, dst *RegistryHealthTCPSpec) {

	dst.Interval = src.Interval
	dst.Timeout = src.Timeout
	dst.Threshold = src.Threshold
	dst.Address = src.Address
}

func Convert_v1alpha3_RegistryNotificationsSpec_To_v1beta1_RegistryNotificationsSpec(src *v1alpha3.RegistryNotificationsSpec, dst *RegistryNotificationsSpec) {

	dst.Endpoints = make([]RegistryNotificationEndpointSpec, len(src.Endpoints))
	for _, ep := range src.Endpoints {
		dst.Endpoints = append(dst.Endpoints, RegistryNotificationEndpointSpec{
			Name:              ep.Name,
			URL:               ep.URL,
			Disabled:          ep.Disabled,
			Threshold:         ep.Threshold,
			Timeout:           ep.Timeout,
			Backoff:           ep.Backoff,
			Headers:           ep.Headers,
			IgnoredMediaTypes: ep.IgnoredMediaTypes,
			Ignore: RegistryNotificationEndpointIgnoreSpec{
				MediaTypes: ep.Ignore.MediaTypes,
				Actions:    ep.Ignore.Actions,
			},
		})
	}

	dst.Events = RegistryNotificationEventsSpec{
		IncludeReferences: src.Events.IncludeReferences,
	}
}

func Convert_v1alpha3_RegistryAuthenticationSpec_To_v1beta1_RegistryAuthenticationSpec(src *v1alpha3.RegistryAuthenticationSpec, dst *RegistryAuthenticationSpec) {

	if src.Silly != nil {
		dst.Silly = &RegistryAuthenticationSillySpec{
			Realm:   src.Silly.Realm,
			Service: src.Silly.Service,
		}
	}

	if src.Token != nil {
		dst.Token = &RegistryAuthenticationTokenSpec{
			Realm:          src.Token.Realm,
			Service:        src.Token.Service,
			Issuer:         src.Token.Issuer,
			CertificateRef: src.Token.CertificateRef,
			AutoRedirect:   src.Token.AutoRedirect,
		}
	}

	if src.HTPasswd != nil {
		dst.HTPasswd = &RegistryAuthenticationHTPasswdSpec{
			Realm:     src.HTPasswd.Realm,
			SecretRef: src.HTPasswd.SecretRef,
		}
	}
}

func Convert_v1alpha3_RegistryValidationSpec_To_v1beta1_RegistryValidationSpec(src *v1alpha3.RegistryValidationSpec, dst *RegistryValidationSpec) {

	dst.Disabled = src.Disabled
	dst.Manifests = RegistryValidationManifestSpec{
		URLs: RegistryValidationManifestURLsSpec{
			Allow: src.Manifests.URLs.Allow,
			Deny:  src.Manifests.URLs.Deny,
		},
	}
}

func Convert_v1alpha3_RegistryCompatibilitySpec_To_v1beta1_RegistryCompatibilitySpec(src *v1alpha3.RegistryCompatibilitySpec, dst *RegistryCompatibilitySpec) {

	dst.Schema1 = RegistryCompatibilitySchemaSpec{
		Enabled:        src.Schema1.Enabled,
		CertificateRef: src.Schema1.CertificateRef,
	}
}

func Convert_v1alpha3_RegistryStorageSpec_To_v1beta1_RegistryStorageSpec(src *v1alpha3.RegistryStorageSpec, dst *RegistryStorageSpec) {

	dst.Cache = RegistryStorageCacheSpec{
		Blobdescriptor: src.Cache.Blobdescriptor,
	}

	dst.Maintenance = RegistryStorageMaintenanceSpec{
		UploadPurging: RegistryStorageMaintenanceUploadPurgingSpec{
			Enabled:  src.Maintenance.UploadPurging.Enabled,
			DryRun:   src.Maintenance.UploadPurging.DryRun,
			Age:      src.Maintenance.UploadPurging.Age,
			Interval: src.Maintenance.UploadPurging.Interval,
		},
		ReadOnly: RegistryStorageMaintenanceReadOnlySpec{
			Enabled: src.Maintenance.ReadOnly.Enabled,
		},
	}

	dst.Delete = RegistryStorageDeleteSpec{
		Enabled: src.Delete.Enabled,
	}

	dst.Redirect = RegistryStorageRedirectSpec{
		Disable: src.Redirect.Disable,
	}

	Convert_v1alpha3_RegistryStorageDriverSpec_To_v1beta1_RegistryStorageDriverSpec(&src.Driver, &dst.Driver)
}

func Convert_v1alpha3_RegistryStorageDriverSpec_To_v1beta1_RegistryStorageDriverSpec(src *v1alpha3.RegistryStorageDriverSpec, dst *RegistryStorageDriverSpec) {

	if src.InMemory != nil {
		dst.InMemory = &RegistryStorageDriverInmemorySpec{}
	}

	if src.FileSystem != nil {
		dst.FileSystem = &RegistryStorageDriverFilesystemSpec{
			VolumeSource: src.FileSystem.VolumeSource,
			MaxThreads:   src.FileSystem.MaxThreads,
			Prefix:       src.FileSystem.Prefix,
		}
	}

	if src.S3 != nil {
		dst.S3 = &RegistryStorageDriverS3Spec{
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
		}
	}

	if src.Swift != nil {
		dst.Swift = &RegistryStorageDriverSwiftSpec{
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
		}
	}
}

func Convert_v1alpha3_RegistryMiddlewaresSpec_To_v1beta1_RegistryMiddlewaresSpec(src *v1alpha3.RegistryMiddlewaresSpec, dst *RegistryMiddlewaresSpec) {

	if len(src.Storage) > 0 {
		dst.Storage = make([]RegistryMiddlewareSpec, len(src.Storage))
		for _, storage := range src.Storage {
			spec := RegistryMiddlewareSpec{}
			Convert_v1alpha3_RegistryMiddlewareSpec_To_v1beta1_RegistryMiddlewareSpec(&storage, &spec)
			dst.Storage = append(dst.Storage, spec)
		}
	}

	if len(src.Registry) > 0 {
		dst.Registry = make([]RegistryMiddlewareSpec, len(src.Registry))
		for _, registry := range src.Registry {
			spec := RegistryMiddlewareSpec{}
			Convert_v1alpha3_RegistryMiddlewareSpec_To_v1beta1_RegistryMiddlewareSpec(&registry, &spec)
			dst.Registry = append(dst.Registry, spec)
		}
	}

	if len(src.Repository) > 0 {
		dst.Repository = make([]RegistryMiddlewareSpec, len(src.Repository))
		for _, repository := range src.Repository {
			spec := RegistryMiddlewareSpec{}
			Convert_v1alpha3_RegistryMiddlewareSpec_To_v1beta1_RegistryMiddlewareSpec(&repository, &spec)
			dst.Repository = append(dst.Repository, spec)
		}
	}
}

func Convert_v1alpha3_RegistryMiddlewareSpec_To_v1beta1_RegistryMiddlewareSpec(src *v1alpha3.RegistryMiddlewareSpec, dst *RegistryMiddlewareSpec) {
	dst.Name = src.Name
	dst.OptionsRef = src.OptionsRef
}

func Convert_v1alpha3_RegistryRedisSpec_To_v1beta1_RegistryRedisSpec(src *v1alpha3.RegistryRedisSpec, dst *RegistryRedisSpec) {
	dst.RedisConnection = src.RedisConnection
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
	dst.DialTimeout = src.DialTimeout

	dst.Pool = RegistryRedisPoolSpec{
		MaxIdle:     src.Pool.MaxIdle,
		MaxActive:   src.Pool.MaxActive,
		IdleTimeout: src.Pool.IdleTimeout,
	}
}
