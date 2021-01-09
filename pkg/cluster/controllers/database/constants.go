package database

const (
	CheckDatabaseHealthError          = "Check database health error"
	ApplyDatabaseHealthError          = "Apply database CR error"
	CreateDatabaseCrError             = "Create database CR error"
	UpdateDatabaseCrError             = "Update database CR error"
	GenerateDatabaseCrError           = "Generate database CR error"
	GetDatabaseCrError                = "Get database CR error"
	SetOwnerReferenceError            = "Set owner reference error"
	DefaultUnstructuredConverterError = "Default unstructured converter error"
)

const (
	DownScalingDatabase     = "DatabaseDownScaling"
	UpScalingDatabase       = "DatabaseUpScaling"
	RollingUpgradesDatabase = "DatabaseRollingUpgrades"

	MessageDatabaseCreate = "Database  %s already created."

	MessageDatabaseUpdate = "Database  %s already update."

	MessageDatabaseDownScaling     = "Database downscale from %d to %d"
	MessageDatabaseUpScaling       = "Database upscale from %d to %d"
	MessageDatabaseRollingUpgrades = "Database resource from %s to %s"
)

const (
	InClusterDatabasePort              = "5432"
	InClusterDatabasePortInt32   int32 = 5432
	InClusterDatabaseUserName          = "postgres"
	InClusterDatabaseName              = "postgres"
	InClusterDatabasePasswordKey       = "password"
)

const (
	PostgresCRDResourceKind   = "postgresql"
	PostgresCRDResourcePlural = "postgresqls"

	// GroupName is the group name for the operator CRDs.
	GroupName = "acid.zalan.do"

	APIVersion = "v1"
)