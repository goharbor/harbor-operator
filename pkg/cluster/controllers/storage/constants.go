package storage

// Define error message.
const (
	GetMinIOError            = "get minIO error"
	GenerateMinIOCrError     = "generate minIO cr error"
	UpdateMinIOError         = "update minIO error"
	CreateMinIOSecretError   = "create minIO secret error" // nolint:gosec
	CreateMinIOIngressError  = "create ingress of minIO error"
	CreateMinIOError         = "create minIO CR error"
	CreateDefaultBucketError = "create default bucket in minIO Error"
	GetMinIOProperties       = "get MinIO Properties error"
	UpdateIngressError       = "update minIO ingress error"
	GetMinIOIngressError     = "get minIO ingress error"
)
