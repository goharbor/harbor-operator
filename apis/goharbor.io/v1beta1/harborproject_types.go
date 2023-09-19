package v1beta1

import (
	goyaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "sigs.k8s.io/yaml"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=harborproject
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor",shortName="hp"
// +kubebuilder:printcolumn:name="ProjectName",type=string,JSONPath=`.spec.projectName`,description="Project name in Harbor"
// +kubebuilder:printcolumn:name="HarborServerConfig",type=string,JSONPath=`.spec.harborServerConfig`,description="HarborServerConfiguration name"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`,description="HarborProject status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// HarborProject is the Schema for the harbors projects.
type HarborProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HarborProjectSpec `json:"spec,omitempty"`

	Status HarborProjectStatus `json:"status,omitempty"`
}

// HarborProjectSpec defines the spec of HarborProject.
type HarborProjectSpec struct {
	// The name of the harbor project. Has to match harbor's naming rules.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-z0-9]+(?:[._-][a-z0-9]+)*$"
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:MinLength=1
	ProjectName string `json:"projectName" yaml:"project_name"`
	// The CVE allowlist for the project.
	// +kubebuilder:validation:Optional
	CveAllowList []string `json:"cveAllowList" yaml:"cve_allow_list_items"`
	// The project's storage quota in human-readable format, like in Kubernetes memory requests/limits (Ti, Gi, Mi, Ki). The Harbor's default value is used if empty.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*(Ti|Gi|Mi|Ki)$"
	StorageQuota string `json:"storageQuota" yaml:"storage_quota"`
	// HarborProjectMetadata related configurations.
	// +kubebuilder:validation:Optional
	HarborProjectMetadata *HarborProjectMetadata `json:"metadata" yaml:"metadata"`
	// Group or user memberships of the project.
	// +kubebuilder:validation:Optional
	HarborProjectMemberships []*HarborProjectMember `json:"memberships" yaml:"memberships"`
	// HarborServerConfig contains the name of a HarborServerConfig resource describing the harbor instance to manage.
	// +kubebuilder:validation:Required
	HarborServerConfig string `json:"harborServerConfig"`
}

// ToJSON converts project spec to json payload.
func (h HarborProjectSpec) ToJSON() ([]byte, error) {
	data, err := goyaml.Marshal(h)
	if err != nil {
		return nil, err
	}

	// convert yaml to json
	return k8syaml.YAMLToJSON(data)
}

// HarborProjectMetadata defines the project related metadata.
type HarborProjectMetadata struct {
	// Whether content trust is enabled or not. If enabled, user can't pull unsigned images from this project.
	// +kubebuilder:validation:Optional
	EnableContentTrust *bool `json:"enableContentTrust,omitempty" yaml:"enable_content_trust,omitempty"`
	// Whether cosign content trust is enabled or not. Similar to enableContentTrust, but using cosign.
	// +kubebuilder:validation:Optional
	EnableContentTrustCosign *bool `json:"enableContentTrustCosign,omitempty" yaml:"enable_content_trust_cosign,omitempty"`
	// Whether to scan images automatically after pushing.
	// +kubebuilder:validation:Optional
	AutoScan *bool `json:"autoScan,omitempty" yaml:"auto_scan,omitempty"`
	// If an image's vulnerablilities are higher than the severity defined here, the image can't be pulled. Can be either `none`, `low`, `medium`, `high` or `critical`.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=none;low;medium;high;critical
	Severity string `json:"severity,omitempty" yaml:"severity,omitempty"`
	// Whether to prevent vulnerable images from running.
	// +kubebuilder:validation:Optional
	PreventVulnerable *bool `json:"preventVulnerable,omitempty" yaml:"prevent_vulnerable,omitempty"`
	// The flag to indicate whether the project should be public or not.
	// +kubebuilder:validation:Optional
	Public *bool `json:"public,omitempty" yaml:"public,omitempty"`
	// Whether this project reuses the system level CVE allowlist for itself. If this is set to `true`, the actual allowlist associated with this project will be ignored.
	// +kubebuilder:validation:Optional
	ReuseSysCveAllowlist *bool `json:"reuseSysCveAllowlist,omitempty" yaml:"reuse_sys_cve_allowlist,omitempty"`
}

// HarborProjectMember is a member of a HarborProject. Can be a user or group.
type HarborProjectMember struct {
	// Type of the member, group or user
	// +kubebuilder:validation:Enum="group";"user"
	Type string `json:"type" yaml:"type"`
	// Name of the member. Has to match with a existing user or group
	Name string `json:"name" yaml:"name"`
	// Role of the member in the Project. This controls the member's permissions on the project.
	// +kubebuilder:validation:Enum="projectAdmin";"developer";"guest";"maintainer"
	Role string `json:"role" yaml:"role"`
}

// HarborProjectStatusType defines the status type of project.
type HarborProjectStatusType string

const (
	// HarborProjectPhaseReady represents ready status.
	HarborProjectStatusReady HarborProjectStatusType = "Success"
	// HarborProjectPhaseFail represents fail status.
	HarborProjectStatusFail HarborProjectStatusType = "Fail"
	// HarborProjectPhaseError represents unknown status.
	HarborProjectStatusUnknown HarborProjectStatusType = "Unknown"
)

// HarborProjectStatus defines the status of HarborProject.
type HarborProjectStatus struct {
	// Status represents harbor project status.
	// +kubebuilder:validation:Optional
	Status HarborProjectStatusType `json:"status,omitempty"`
	// ProjectID represents ID of the managed project.
	// +kubebuilder:validation:Optional
	ProjectID int32 `json:"projectID,omitempty"`
	// QuotaID is the ID of the project's quota. Used to be able to update it.
	// +kubebuilder:validation:Optional
	QuotaID int64 `json:"quotaID,omitempty"`
	// MembershipHash provides a way to quickly notice changes in project membership.
	// +kubebuilder:validation:Optional
	MembershipHash string `json:"membershipHash,omitempty"`
	// Reason represents status reason.
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`
	// Message provides human-readable message.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
	// LastApplyTime represents the last apply configuration time.
	// +kubebuilder:validation:Optional
	LastApplyTime *metav1.Time `json:"lastApplyTime,omitempty"`
}

// +kubebuilder:object:root=true
// HarborProjectList contains a list of HarborProjects.
type HarborProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborProject `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&HarborProject{}, &HarborProjectList{})
}
