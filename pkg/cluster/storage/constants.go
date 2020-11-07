package storage

// Define error message
const (
	GetMinIOError           = "Get minIO error"
	UpdateMinIOError        = "Update minIO error"
	CreateMinIOSecretError  = "Create minIO secret error"
	CreateMinIOServiceError = "Create service of minIO error"
	CreateMinIOIngressError = "Create ingress of minIO error"
	GetMinIOSecretError     = "Get minIO secret error"
	CreateMinIOError        = "Create minIO CR error"
	ScaleMinIOError         = "Scale minIO error"

	CreateExternalSecretError = "Create external storage secret error"
	GetExternalSecretError    = "Get external storage secret error"
	UpdateExternalSecretError = "Update external storage secret error"
	NotSupportType            = "The type of storage are not supported"
	CreateDefaultBucketError  = "Create default bucket in minIO Error"

	CreateChartMuseumStorageSecretError   = "Create chart museum storage secret err"
	GenerateChartMuseumStorageSecretError = "Generate chart museum storage secret err"
)
