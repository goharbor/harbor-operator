package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Tenant is a specification for a MinIO resource
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Scheduler TenantScheduler `json:"scheduler,omitempty"`
	Spec      TenantSpec      `json:"spec"`
	// Status provides details of the state of the Tenant
	// +optional
	Status TenantStatus `json:"status"`
}

type TenantScheduler struct {
	// SchedulerName defines the name of scheduler to be used to schedule Tenant pods
	Name string `json:"name"`
}

// TenantSpec is the spec for a Tenant resource
type TenantSpec struct {
	// Definition for Cluster in given MinIO cluster
	Zones []Zone `json:"zones"`
	// Image defines the Tenant Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// ServiceName defines name of the Service that will be created for this instance, if none is specified,
	// it will default to the instance name
	// +optional
	ServiceName string `json:"serviceName,omitempty"`
	// ImagePullSecret defines the secret to be used for pull image from a private Docker image.
	// +optional
	ImagePullSecret corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`
	// Pod Management Policy for pod created by StatefulSet
	// +optional
	PodManagementPolicy appsv1.PodManagementPolicyType `json:"podManagementPolicy,omitempty"`
	// Metadata defines the object metadata passed to each pod that is a part of this Tenant
	Metadata *metav1.ObjectMeta `json:"metadata,omitempty"`
	// If provided, use this secret as the credentials for Tenant resource
	// Otherwise MinIO server creates dynamic credentials printed on MinIO server startup banner
	// +optional
	CredsSecret *corev1.LocalObjectReference `json:"credsSecret,omitempty"`
	// If provided, use these environment variables for Tenant resource
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// ExternalCertSecret allows a user to specify custom CA certificate, and private key. This is
	// used for enabling TLS support on MinIO Pods.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
	// ExternalClientCertSecret allows a user to specify custom CA client certificate, and private key. This is
	// used for adding client certificates on MinIO Pods --> used for KES authentication.
	// +optional
	ExternalClientCertSecret *LocalCertificateReference `json:"externalClientCertSecret,omitempty"`
	// Mount path for MinIO volume (PV). Defaults to /export
	// +optional
	Mountpath string `json:"mountPath,omitempty"`
	// Subpath inside mount path. This is the directory where MinIO stores data. Default to "" (empty)
	// +optional
	Subpath string `json:"subPath,omitempty"`
	// Liveness Probe for container liveness. Container will be restarted if the probe fails.
	// +optional
	Liveness *Liveness `json:"liveness,omitempty"`
	// RequestAutoCert allows user to enable Kubernetes based TLS cert generation and signing as explained here:
	// https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/
	// +optional
	RequestAutoCert bool `json:"requestAutoCert,omitempty"`
	// CertConfig allows users to set entries like CommonName, Organization, etc for the certificate
	// +optional
	CertConfig *CertificateConfig `json:"certConfig,omitempty"`
	// Security Context allows user to set entries like runAsUser, privilege escalation etc.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// ConsoleConfiguration is for setting up minio/console for graphical user interface
	//+optional
	Console *ConsoleConfiguration `json:"console,omitempty"`
	// KES is for setting up minio/kes as MinIO KMS
	//+optional
	KES *KESConfig `json:"kes,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run pods of all MinIO
	// Pods created as a part of this Tenant.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// PriorityClassName indicates the Pod priority and hence importance of a Pod relative to other Pods.
	// This is applied to MinIO pods only.
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// Image pull policy. One of Always, Never, IfNotPresent.
	// This is applied to MinIO pods only.
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// Zone defines the spec for a MinIO Zone
type Zone struct {
	// Name of the zone
	// +optional
	Name string `json:"name,omitempty"`
	// Number of Servers in the zone
	Servers int32 `json:"servers"`
	// Number of persistent volumes that will be attached per server
	VolumesPerServer int32 `json:"volumesPerServer"`
	// VolumeClaimTemplate allows a user to specify how volumes inside a Tenant
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// If specified, affinity will define the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations allows users to set entries like effect, key, operator, value.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// LocalCertificateReference defines the spec for a local certificate
type LocalCertificateReference struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// Liveness specifies the spec for liveness probe
type Liveness struct {
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	PeriodSeconds       int32 `json:"periodSeconds"`
	TimeoutSeconds      int32 `json:"timeoutSeconds"`
}

// CertificateConfig is a specification for certificate contents
type CertificateConfig struct {
	CommonName       string   `json:"commonName,omitempty"`
	OrganizationName []string `json:"organizationName,omitempty"`
	DNSNames         []string `json:"dnsNames,omitempty"`
}

// ConsoleConfiguration defines the specifications for Console Deployment
type ConsoleConfiguration struct {
	// Replicas defines number of pods for KES StatefulSet.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image defines the Tenant Console Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy. One of Always, Never, IfNotPresent.
	// This is applied to MinIO Console pods only.
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// This secret provides all environment variables for KES
	// This is a mandatory field
	ConsoleSecret *corev1.LocalObjectReference `json:"consoleSecret"`
	Metadata      *metav1.ObjectMeta           `json:"metadata,omitempty"`
	// If provided, use these environment variables for Console resource
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// ExternalCertSecret allows a user to specify custom CA certificate, and private key. This is
	// used for enabling TLS support on Console Pods.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
}

// KESConfig defines the specifications for KES StatefulSet
type KESConfig struct {
	// Replicas defines number of pods for KES StatefulSet.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image defines the Tenant KES Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy. One of Always, Never, IfNotPresent.
	// This is applied to KES pods only.
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// This kesSecret serves as the configuration for KES
	// This is a mandatory field
	Configuration *corev1.LocalObjectReference `json:"kesSecret"`
	Metadata      *metav1.ObjectMeta           `json:"metadata,omitempty"`
	// ExternalCertSecret allows a user to specify custom CA certificate, and private key for group replication SSL.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
	// ClientCertSecret allows a user to specify a custom root certificate, client certificate and client private key. This is
	// used for adding client certificates on KES --> used for KES authentication against Vault or other KMS that supports mTLS.
	// +optional
	ClientCertSecret *LocalCertificateReference `json:"clientCertSecret,omitempty"`
}

// TenantStatus is the status for a Tenant resource
type TenantStatus struct {
	CurrentState      string `json:"currentState"`
	AvailableReplicas int32  `json:"availableReplicas"`
}

// TenantList is a list of Tenant resources
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Tenant `json:"items"`
}
