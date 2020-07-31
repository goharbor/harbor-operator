package v1alpha1

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal"}
// +kubebuilder:default="info"
// HarborLogLevel is the log level for Harbor.
type HarborLogLevel string

const (
	HarborDebug   HarborLogLevel = "debug"
	HarborInfo    HarborLogLevel = "info"
	HarborWarning HarborLogLevel = "warning"
	HarborError   HarborLogLevel = "error"
	HarborFatal   HarborLogLevel = "fatal"
)

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warn","error"}
// +kubebuilder:default="info"
// CoreLogLevel is the log level for Core.
type CoreLogLevel string

const (
	CoreDebug   CoreLogLevel = "debug"
	CoreInfo    CoreLogLevel = "info"
	CoreWarning CoreLogLevel = "warn"
	CoreError   CoreLogLevel = "error"
)

func (l HarborLogLevel) Core() CoreLogLevel {
	switch l {
	default:
		return CoreInfo
	case HarborDebug:
		return CoreDebug
	case HarborInfo:
		return CoreInfo
	case HarborWarning:
		return CoreWarning
	case HarborError, HarborFatal:
		return CoreError
	}
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warn","error"}
// +kubebuilder:default="info"
// CoreLogLevel is the log level for Registry.
type RegistryLogLevel string

const (
	RegistryDebug   RegistryLogLevel = "debug"
	RegistryInfo    RegistryLogLevel = "info"
	RegistryWarning RegistryLogLevel = "warning"
	RegistryError   RegistryLogLevel = "error"
)

func (l HarborLogLevel) Registry() RegistryLogLevel {
	switch l {
	default:
		return RegistryInfo
	case HarborDebug:
		return RegistryDebug
	case HarborInfo:
		return RegistryInfo
	case HarborWarning:
		return RegistryWarning
	case HarborError, HarborFatal:
		return RegistryError
	}
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal"}
// CoreLogLevel is the log level for RegistryController.
type RegistryCtlLogLevel string

const (
	RegistryCtlDebug   RegistryCtlLogLevel = "debug"
	RegistryCtlInfo    RegistryCtlLogLevel = "info"
	RegistryCtlWarning RegistryCtlLogLevel = "warning"
	RegistryCtlError   RegistryCtlLogLevel = "error"
	RegistryCtlFatal   RegistryCtlLogLevel = "fatal"
)

func (l HarborLogLevel) RegistryCtl() RegistryCtlLogLevel {
	switch l {
	default:
		return RegistryCtlInfo
	case HarborDebug:
		return RegistryCtlDebug
	case HarborInfo:
		return RegistryCtlInfo
	case HarborWarning:
		return RegistryCtlWarning
	case HarborError:
		return RegistryCtlError
	case HarborFatal:
		return RegistryCtlFatal
	}
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"DEBUG","INFO","WARNING","ERROR","FATAL"}
// +kubebuilder:default="INFO"
// CoreLogLevel is the log level for JobService.
type JobServiceLogLevel string

const (
	JobServiceDebug        JobServiceLogLevel = "DEBUG"
	JobServiceInfo         JobServiceLogLevel = "INFO"
	JobServiceWarning      JobServiceLogLevel = "WARNING"
	JobServiceError        JobServiceLogLevel = "ERROR"
	JobServiceFatal        JobServiceLogLevel = "FATAL"
	JobServiceDefaultLevel JobServiceLogLevel = JobServiceInfo
)

func (l HarborLogLevel) JobService() JobServiceLogLevel {
	switch l {
	default:
		return JobServiceDefaultLevel
	case HarborDebug:
		return JobServiceDebug
	case HarborInfo:
		return JobServiceInfo
	case HarborWarning:
		return JobServiceWarning
	case HarborError:
		return JobServiceError
	case HarborFatal:
		return JobServiceFatal
	}
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal","panic"}
// +kubebuilder:default="info"
// NotaryLogLevel is the log level for NotaryServer and NotarySigner.
type NotaryLogLevel string

const (
	NotaryDebug        NotaryLogLevel = "debug"
	NotaryInfo         NotaryLogLevel = "info"
	NotaryWarning      NotaryLogLevel = "warning"
	NotaryError        NotaryLogLevel = "error"
	NotaryFatal        NotaryLogLevel = "fatal"
	NotaryPanic        NotaryLogLevel = "panic"
	NotaryDefaultLevel NotaryLogLevel = NotaryInfo
)

func (l HarborLogLevel) Notary() NotaryLogLevel {
	switch l {
	default:
		return NotaryDefaultLevel
	case HarborDebug:
		return NotaryDebug
	case HarborInfo:
		return NotaryInfo
	case HarborWarning:
		return NotaryWarning
	case HarborError:
		return NotaryError
	case HarborFatal:
		return NotaryFatal
	}
}
