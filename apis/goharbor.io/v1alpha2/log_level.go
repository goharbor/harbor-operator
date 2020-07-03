package v1alpha2

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal"}
// +kubebuilder:default="info"
// HarborLogLevel is the log level for Harbor.
type HarborLogLevel string

const (
	HarborDebug        HarborLogLevel = "debug"
	HarborInfo                        = "info"
	HarborWarning                     = "warning"
	HarborError                       = "error"
	HarborFatal                       = "fatal"
	HarborDefaultLevel                = HarborInfo
)

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warn","error"}
// +kubebuilder:default="info"
// CoreLogLevel is the log level for Core.
type CoreLogLevel string

const (
	CoreDebug        CoreLogLevel = "debug"
	CoreInfo                      = "info"
	CoreWarning                   = "warn"
	CoreError                     = "error"
	CoreDefaultLevel              = CoreInfo
)

func (l HarborLogLevel) Core() CoreLogLevel {
	switch l {
	default:
		return CoreDefaultLevel
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
	RegistryDebug        RegistryLogLevel = "debug"
	RegistryInfo                          = "info"
	RegistryWarning                       = "warning"
	RegistryError                         = "error"
	RegistryDefaultLevel                  = RegistryInfo
)

func (l HarborLogLevel) Registry() RegistryLogLevel {
	switch l {
	default:
		return RegistryDefaultLevel
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
// CoreLogLevel is the log level for RegistryController.
type RegistryCtlLogLevel string

const (
	RegistryCtlDebug        RegistryCtlLogLevel = "debug"
	RegistryCtlInfo                             = "info"
	RegistryCtlWarning                          = "warning"
	RegistryCtlError                            = "error"
	RegistryCtlFatal                            = "fatal"
	RegistryCtlDefaultLevel                     = RegistryCtlInfo
)

func (l HarborLogLevel) RegistryCtl() RegistryCtlLogLevel {
	switch l {
	default:
		return RegistryCtlDefaultLevel
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
	JobServiceInfo                            = "INFO"
	JobServiceWarning                         = "WARNING"
	JobServiceError                           = "ERROR"
	JobServiceFatal                           = "FATAL"
	JobServiceDefaultLevel                    = JobServiceInfo
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
