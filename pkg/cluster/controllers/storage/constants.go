package storage

// Define error message.
const (
	GetMinIOError            = "Get minIO error"
	GenerateMinIOCrError     = "Generate minIO cr error"
	UpdateMinIOError         = "Update minIO error"
	CreateMinIOSecretError   = "Create minIO secret error" // nolint:gosec
	CreateMinIOServiceError  = "Create service of minIO error"
	CreateMinIOIngressError  = "Create ingress of minIO error"
	CreateMinIOError         = "Create minIO CR error"
	ScaleMinIOError          = "Scale minIO error"
	CreateDefaultBucketError = "Create default bucket in minIO Error"
	getMinIOProperties       = "Get MinIO Properties error"
	updateIngressError       = "update minIO ingress error"
)
