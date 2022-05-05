package v1alpha1

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal"}
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
// CoreLogLevel is the log level for Core.
type CoreLogLevel string

const (
	CoreDebug   CoreLogLevel = "debug"
	CoreInfo    CoreLogLevel = "info"
	CoreWarning CoreLogLevel = "warn"
	CoreError   CoreLogLevel = "error"
)

// Core get the log level for Core component.
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
// +kubebuilder:validation:Enum={"debug","info","warning","error","panic"}
// ExporterLogLevel is the log level for Exporter.
type ExporterLogLevel string

func (l ExporterLogLevel) String() string {
	return string(l)
}

const (
	ExporterDebug   ExporterLogLevel = "debug"
	ExporterInfo    ExporterLogLevel = "info"
	ExporterWarning ExporterLogLevel = "warning"
	ExporterError   ExporterLogLevel = "error"
	ExporterFatal   ExporterLogLevel = "panic"
)

func (l HarborLogLevel) Exporter() ExporterLogLevel {
	switch l {
	default:
		return ExporterInfo
	case HarborDebug:
		return ExporterDebug
	case HarborInfo:
		return ExporterInfo
	case HarborWarning:
		return ExporterWarning
	case HarborError:
		return ExporterError
	case HarborFatal:
		return ExporterFatal
	}
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warning","error"}
// RegistryLogLevel is the log level for Registry.
type RegistryLogLevel string

const (
	RegistryDebug   RegistryLogLevel = "debug"
	RegistryInfo    RegistryLogLevel = "info"
	RegistryWarning RegistryLogLevel = "warning"
	RegistryError   RegistryLogLevel = "error"
)

// Registry get the log level for Registry component.
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
// +kubebuilder:validation:Enum={"text","json","logstash"}
// RegistryLogFormatter is the log formatter for Registry.
type RegistryLogFormatter string

const (
	RegistryTextFormatter     RegistryLogFormatter = "text"
	RegistryJSONFormatter     RegistryLogFormatter = "json"
	RegistryLogstashFormatter RegistryLogFormatter = "logstash"
)

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal"}
// RegistryCtlLogLevel is the log level for RegistryController.
type RegistryCtlLogLevel string

const (
	RegistryCtlDebug   RegistryCtlLogLevel = "debug"
	RegistryCtlInfo    RegistryCtlLogLevel = "info"
	RegistryCtlWarning RegistryCtlLogLevel = "warning"
	RegistryCtlError   RegistryCtlLogLevel = "error"
	RegistryCtlFatal   RegistryCtlLogLevel = "fatal"
)

// RegistryCtl get the log level for RegistryController component.
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
// JobServiceLogLevel is the log level for JobService.
type JobServiceLogLevel string

const (
	JobServiceDebug   JobServiceLogLevel = "DEBUG"
	JobServiceInfo    JobServiceLogLevel = "INFO"
	JobServiceWarning JobServiceLogLevel = "WARNING"
	JobServiceError   JobServiceLogLevel = "ERROR"
	JobServiceFatal   JobServiceLogLevel = "FATAL"
)

// JobService get the log level for JobService component.
func (l HarborLogLevel) JobService() JobServiceLogLevel {
	switch l {
	default:
		return JobServiceInfo
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
// NotaryLogLevel is the log level for NotaryServer and NotarySigner.
type NotaryLogLevel string

const (
	NotaryDebug   NotaryLogLevel = "debug"
	NotaryInfo    NotaryLogLevel = "info"
	NotaryWarning NotaryLogLevel = "warning"
	NotaryError   NotaryLogLevel = "error"
	NotaryFatal   NotaryLogLevel = "fatal"
	NotaryPanic   NotaryLogLevel = "panic"
)

// Notary get the log level for Notary component.
func (l HarborLogLevel) Notary() NotaryLogLevel {
	switch l {
	default:
		return NotaryInfo
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

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"debug","info","warning","error","fatal","panic"}
// TrivyLogLevel is the log level for Trivy.
type TrivyLogLevel string

const (
	TrivyDebug   TrivyLogLevel = "debug"
	TrivyInfo    TrivyLogLevel = "info"
	TrivyWarning TrivyLogLevel = "warning"
	TrivyError   TrivyLogLevel = "error"
	TrivyFatal   TrivyLogLevel = "fatal"
	TrivyPanic   TrivyLogLevel = "panic"
)

// Trivy get the log level for Trivy component.
func (l HarborLogLevel) Trivy() TrivyLogLevel {
	switch l {
	default:
		return TrivyInfo
	case HarborDebug:
		return TrivyDebug
	case HarborInfo:
		return TrivyInfo
	case HarborWarning:
		return TrivyWarning
	case HarborError:
		return TrivyError
	case HarborFatal:
		return TrivyFatal
	}
}
