package v1beta1

import (
	"encoding/json"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +resource:path=harborconfiguration
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor",shortName="hc"
// +kubebuilder:printcolumn:name="HarborName",type=string,JSONPath=`.metadata.labels['harbor-name']`,description="Harbor instance name"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`,description="HarborConfiguration status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// HarborConfiguration is the Schema for the harbors configuration.
type HarborConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HarborConfigurationSpec `json:"spec,omitempty"`

	Status HarborConfigurationStatus `json:"status,omitempty"`
}

// HarborConfigurationSpec defines the spec of HarborConfiguration.
type HarborConfigurationSpec struct {
	// AdditionalProperties provides workaround for those unknown configuration fields in the future.
	// +kubebuilder:validation:Optional
	Extension apiextensionsv1.JSON `json:"extension,omitempty"`
	// The auth mode of current system, such as "db_auth", "ldap_auth", "oidc_auth".
	// +kubebuilder:validation:Optional
	AuthMode string `json:"auth_mode,omitempty"`
	// Email related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationEmailSpec `json:",inline"`
	// LDAP related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationLdapSpec `json:",inline"`
	// Indicate who can create projects, it could be ''adminonly'' or ''everyone''.
	// +kubebuilder:validation:Optional
	ProjectCreationRestriction string `json:"project_creation_restriction,omitempty"`
	// The flag to indicate whether Harbor is in readonly mode.
	// +kubebuilder:validation:Optional
	ReadOnly *bool `json:"read_only,omitempty"`
	// Whether the Harbor instance supports self-registration.  If it''s set to false, admin need to add user to the instance.
	// +kubebuilder:validation:Optional
	SelfRegistration *bool `json:"self_registration,omitempty"`
	// The expiration time of the token for internal Registry, in minutes.
	// +kubebuilder:validation:Optional
	TokenExpiration int `json:"token_expiration,omitempty"`
	// HttpAuthproxy related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationHTTPAuthProxySpec `json:",inline"`
	// Uaa related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationUaaSpec `json:",inline"`
	// Oidc related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationOidcSpec `json:",inline"`
	// The robot account token duration in days.
	// +kubebuilder:validation:Optional
	RobotTokenDuration int `json:"robot_token_duration,omitempty"`
	// The rebot account name prefix.
	// +kubebuilder:validation:Optional
	RobotNamePrefix string `json:"robot_name_prefix,omitempty"`
	// Enable notification.
	// +kubebuilder:validation:Optional
	NotificationEnable *bool `json:"notification_enable,omitempty"`
	// Enable quota per project.
	// +kubebuilder:validation:Optional
	QuotaPerProjectEnable *bool `json:"quota_per_project_enable,omitempty"`
	// The storage quota per project.
	// +kubebuilder:validation:Optional
	StoragePerProject int `json:"storage_per_project,omitempty"`
}

// ToJSON converts configuration spec to json payload.
func (h HarborConfigurationSpec) ToJSON() ([]byte, error) {
	jsonData, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	if err = json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	if len(h.Extension.Raw) > 0 {
		extension := make(map[string]interface{})
		if err = json.Unmarshal(h.Extension.Raw, &extension); err != nil {
			return nil, err
		}

		delete(data, "extension")

		for k, v := range extension {
			data[k] = v
		}
	}

	return json.Marshal(data)
}

// HarborConfigurationEmailSpec defines the email related spec.
type HarborConfigurationEmailSpec struct {
	// The sender name for Email notification.
	// +kubebuilder:validation:Optional
	EmailFrom string `json:"email_from,omitempty"`
	// The hostname of SMTP server that sends Email notification.
	// +kubebuilder:validation:Optional
	EmailHost string `json:"email_host,omitempty"`
	// By default it's empty so the email_username is picked
	// +kubebuilder:validation:Optional
	EmailIdentity string `json:"email_identity,omitempty"`
	// Whether or not the certificate will be verified when Harbor tries to access the email server.
	// +kubebuilder:validation:Optional
	EmailInsecure *bool `json:"email_insecure,omitempty"`
	// The username for authenticate against SMTP server.
	// +kubebuilder:validation:Optional
	EmailUsername string `json:"email_username,omitempty"`
	// Email password.
	// +kubebuilder:validation:Optional
	EmailPassword string `json:"email_password,omitempty"`
	// The port of SMTP server.
	// +kubebuilder:validation:Optional
	EmailPort int `json:"email_port,omitempty"`
	// When it''s set to true the system will access Email server via TLS by default.  If it''s set to false, it still will handle "STARTTLS" from server side.
	// +kubebuilder:validation:Optional
	EmailSSL *bool `json:"email_ssl,omitempty"`
}

// HarborConfigurationLDAPSpec defines the ldap related spec.
type HarborConfigurationLdapSpec struct {
	// The Base DN for LDAP binding.
	// +kubebuilder:validation:Optional
	LdapBaseDn string `json:"ldap_base_dn,omitempty"`
	// The filter for LDAP search.
	// +kubebuilder:validation:Optional
	LdapFilter string `json:"ldap_filter,omitempty"`
	// The base DN to search LDAP group.
	// +kubebuilder:validation:Optional
	LdapGroupBaseDn string `json:"ldap_group_base_dn,omitempty"`
	// Specify the ldap group which have the same privilege with Harbor admin.
	// +kubebuilder:validation:Optional
	LdapGroupAdminDn string `json:"ldap_group_admin_dn,omitempty"`
	// The attribute which is used as identity of the LDAP group, default is cn.
	// +kubebuilder:validation:Optional
	LdapGroupAttributeName string `json:"ldap_group_attribute_name,omitempty"`
	// The filter to search the ldap group.
	// +kubebuilder:validation:Optional
	LdapGroupSearchFilter string `json:"ldap_group_search_filter,omitempty"`
	// The scope to search ldap group. ''0-LDAP_SCOPE_BASE, 1-LDAP_SCOPE_ONELEVEL, 2-LDAP_SCOPE_SUBTREE''.
	// +kubebuilder:validation:Optional
	LdapGroupSearchScope int `json:"ldap_group_search_scope,omitempty"`
	// The scope to search ldap users,'0-LDAP_SCOPE_BASE, 1-LDAP_SCOPE_ONELEVEL, 2-LDAP_SCOPE_SUBTREE'.
	// +kubebuilder:validation:Optional
	LdapScope int `json:"ldap_scope,omitempty"`
	// The DN of the user to do the search.
	// +kubebuilder:validation:Optional
	LdapSearchDn string `json:"ldap_search_dn,omitempty"`
	// The password ref of the ldap search dn.
	// +kubebuilder:validation:Optional
	LdapSearchPassword string `json:"ldap_search_password,omitempty"`
	// Timeout in seconds for connection to LDAP server.
	// +kubebuilder:validation:Optional
	LdapTimeout int `json:"ldap_timeout,omitempty"`
	// The attribute which is used as identity for the LDAP binding, such as "CN" or "SAMAccountname".
	// +kubebuilder:validation:Optional
	LdapUID string `json:"ldap_uid,omitempty"`
	// The URL of LDAP server.
	// +kubebuilder:validation:Optional
	LdapURL string `json:"ldap_url,omitempty"`
	// Whether verify your OIDC server certificate, disable it if your OIDC server is hosted via self-hosted certificate.
	// +kubebuilder:validation:Optional
	LdapVerifyCert *bool `json:"ldap_verify_cert,omitempty"`
	// The user attribute to identify the group membership.
	// +kubebuilder:validation:Optional
	LdapGroupMembershipAttribute string `json:"ldap_group_membership_attribute,omitempty"`
}

// HarborConfigurationUaaSpec defines the uaa related spec.
type HarborConfigurationUaaSpec struct {
	// The client id of UAA.
	// +kubebuilder:validation:Optional
	UaaClientID string `json:"uaa_client_id,omitempty"`
	// The client secret of the UAA.
	// +kubebuilder:validation:Optional
	UaaClientSecret string `json:"uaa_client_secret,omitempty"`
	// The endpoint of the UAA.
	// +kubebuilder:validation:Optional
	UaaEndpoint string `json:"uaa_endpoint,omitempty"`
	// Verify the certificate in UAA server.
	// +kubebuilder:validation:Optional
	UaaVerifyCert *bool `json:"uaa_verify_cert,omitempty"`
}

// HarborConfigurationHTTPAuthProxySpec defines the http_authproxy spec.
type HarborConfigurationHTTPAuthProxySpec struct {
	// The endpoint of the HTTP auth.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyEndpoint string `json:"http_authproxy_endpoint,omitempty"`
	// The token review endpoint.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyTokenreviewEndpoint string `json:"http_authproxy_tokenreview_endpoint,omitempty"`
	// The group which has the harbor admin privileges.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyAdminGroups string `json:"http_authproxy_admin_groups,omitempty"`
	// The username which has the harbor admin privileges.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyAdminUsernames string `json:"http_authproxy_admin_usernames,omitempty"`
	// Verify the HTTP auth provider's certificate.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyVerifyCert *bool `json:"http_authproxy_verify_cert,omitempty"`
	// Search user before onboard.
	// +kubebuilder:validation:Optional
	HTTPAuthproxySkipSearch *bool `json:"http_authproxy_skip_search,omitempty"`
	// The certificate of the HTTP auth provider.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyServerCertificate *bool `json:"http_authproxy_server_certificate,omitempty"`
}

// HarborConfigurationOidcSpec defines the oidc spec.
type HarborConfigurationOidcSpec struct {
	// The OIDC provider name.
	// +kubebuilder:validation:Optional
	OidcName string `json:"oidc_name,omitempty"`
	// The endpoint of the OIDC provider.
	// +kubebuilder:validation:Optional
	OidcEndpoint string `json:"oidc_endpoint,omitempty"`
	// The client ID of the OIDC provider.
	// +kubebuilder:validation:Optional
	OidcClientID string `json:"oidc_client_id,omitempty"`
	// The OIDC provider secret.
	// +kubebuilder:validation:Optional
	OidcClientSecret string `json:"oidc_client_secret,omitempty"`
	// The attribute claims the group name.
	// +kubebuilder:validation:Optional
	OidcGroupsClaim string `json:"oidc_groups_claim,omitempty"`
	// The OIDC group which has the harbor admin privileges.
	// +kubebuilder:validation:Optional
	OidcAdminGroup string `json:"oidc_admin_group,omitempty"`
	// The scope of the OIDC provider.
	// +kubebuilder:validation:Optional
	OidcScope string `json:"oidc_scope,omitempty"`
	// The attribute claims the username.
	// +kubebuilder:validation:Optional
	OidcUserClaim string `json:"oidc_user_claim,omitempty"`
	// Verify the OIDC provider's certificate'.
	// +kubebuilder:validation:Optional
	OidcVerifyCert *bool `json:"oidc_verify_cert,omitempty"`
	// Auto onboard the OIDC user.
	// +kubebuilder:validation:Optional
	OidcAutoOnboard *bool `json:"oidc_auto_onboard,omitempty"`
	// Extra parameters to add when redirect request to OIDC provider.
	// +kubebuilder:validation:Optional
	OidcExtraRedirectParms string `json:"oidc_extra_redirect_parms,omitempty"`
}

// HarborConfigurationStatusType defines the status type of configuration.
type HarborConfigurationStatusType string

const (
	// HarborConfigurationPhaseReady represents ready status.
	HarborConfigurationStatusReady HarborConfigurationStatusType = "Success"
	// HarborConfigurationPhaseFail represents fail status.
	HarborConfigurationStatusFail HarborConfigurationStatusType = "Fail"
	// HarborConfigurationPhaseError represents unknown status.
	HarborConfigurationStatusUnknown HarborConfigurationStatusType = "Unknown"
)

// HarborConfigurationStatus defines the status of HarborConfiguration.
type HarborConfigurationStatus struct {
	// Status represents harbor configuration status.
	// +kubebuilder:validation:Optional
	Status HarborConfigurationStatusType `json:"status,omitempty"`
	// Reason represents status reason.
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`
	// Message provides human-readable message.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
	// LastApplyTime represents the last apply configuration time.
	// +kubebuilder:validation:Optional
	LastApplyTime *metav1.Time `json:"lastApplyTime,omitempty"`
	// LastConfiguration represents the configuration of last time.
	// +kubebuilder:validation:Optional
	LastConfiguration *HarborConfigurationSpec `json:"lastConfiguration,omitempty"`
}

// +kubebuilder:object:root=true
// HarborConfigurationList contains a list of HarborConfiguration.
type HarborConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborConfiguration `json:"items"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&HarborConfiguration{}, &HarborConfigurationList{})
}
