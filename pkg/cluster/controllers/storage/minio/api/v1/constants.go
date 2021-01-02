package v1

// MinIOVolumeMountPath specifies the default mount path for MinIO volumes.
const MinIOVolumeMountPath = "/export"

// MinIOCRDResourceKind is the Kind of a Cluster.
const MinIOCRDResourceKind = "Tenant"

// Standard Status messages for Tenant
const (
	StatusInitialized = "Initialized"
	StatusReady       = "Ready"
)
