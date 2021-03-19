package database

const (
	CheckDatabaseHealthError          = "Check database health error"
	CreateDatabaseCrError             = "Create database CR error"
	UpdateDatabaseCrError             = "Update database CR error"
	GenerateDatabaseCrError           = "Generate database CR error"
	GetDatabaseCrError                = "Get database CR error"
	SetOwnerReferenceError            = "Set owner reference error"
	DefaultUnstructuredConverterError = "Default unstructured converter error"
)

const (
	InClusterDatabasePort              = "5432"
	InClusterDatabasePortInt32   int32 = 5432
	InClusterDatabasePasswordKey       = "password"
)

const (
	PostgresCRDResourcePlural = "postgresqls"
	// GroupName is the group name for the operator CRDs.
	GroupName  = "acid.zalan.do"
	APIVersion = "v1"
)
