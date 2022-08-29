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
// +resource:path=harborconfiguration
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="goharbor",shortName="hc"
// +kubebuilder:printcolumn:name="HarborCluster",type=string,JSONPath=`.spec.harborClusterRef`,description="HarborCluster name"
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
	// Configuration defines the harbor configuration types.
	Configuration HarborConfigurationModel `json:"configuration,omitempty"`
	// HarborClusterRef defines the reference of the harbor cluster name.
	HarborClusterRef string `json:"harborClusterRef,omitempty"`
}

// HarborConfigurationModel defines the spec of HarborConfiguration.
type HarborConfigurationModel struct {
	// The auth mode of current system, such as "db_auth", "ldap_auth", "oidc_auth".
	// +kubebuilder:validation:Optional
	AuthMode string `json:"authMode,omitempty" yaml:"auth_mode,omitempty"`
	// Email related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationEmail `json:",inline" yaml:",inline"`
	// LDAP related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationLdap `json:",inline" yaml:",inline"`
	// Indicate who can create projects, it could be ''adminonly'' or ''everyone''.
	// +kubebuilder:validation:Optional
	ProjectCreationRestriction string `json:"projectCreationRestriction,omitempty" yaml:"project_creation_restriction,omitempty"`
	// The flag to indicate whether Harbor is in readonly mode.
	// +kubebuilder:validation:Optional
	ReadOnly *bool `json:"readOnly,omitempty" yaml:"read_only,omitempty"`
	// Whether the Harbor instance supports self-registration.  If it''s set to false, admin need to add user to the instance.
	// +kubebuilder:validation:Optional
	SelfRegistration *bool `json:"selfRegistration,omitempty" yaml:"self_registration,omitempty"`
	// The expiration time of the token for internal Registry, in minutes.
	// +kubebuilder:validation:Optional
	TokenExpiration int `json:"tokenExpiration,omitempty" yaml:"token_expiration,omitempty"`
	// HttpAuthproxy related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationHTTPAuthProxy `json:",inline" yaml:",inline"`
	// Uaa related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationUaa `json:",inline" yaml:",inline"`
	// Oidc related configurations.
	// +kubebuilder:validation:Optional
	HarborConfigurationOidc `json:",inline" yaml:",inline"`
	// The robot account token duration in days.
	// +kubebuilder:validation:Optional
	RobotTokenDuration int `json:"robotTokenDuration,omitempty" yaml:"robot_token_duration,omitempty"`
	// The rebot account name prefix.
	// +kubebuilder:validation:Optional
	RobotNamePrefix string `json:"robotNamePrefix,omitempty" yaml:"robot_name_prefix,omitempty"`
	// Enable notification.
	// +kubebuilder:validation:Optional
	NotificationEnable *bool `json:"notificationEnable,omitempty" yaml:"notification_enable,omitempty"`
	// Enable quota per project.
	// +kubebuilder:validation:Optional
	QuotaPerProjectEnable *bool `json:"quotaPerProjectEnable,omitempty" yaml:"quota_per_project_enable,omitempty"`
	// The storage quota per project.
	// +kubebuilder:validation:Optional
	StoragePerProject int `json:"storagePerProject,omitempty" yaml:"storage_per_project,omitempty"`
}

// ToJSON converts configuration spec to json payload.
func (h HarborConfigurationModel) ToJSON() ([]byte, error) {
	data, err := goyaml.Marshal(h)
	if err != nil {
		return nil, err
	}

	// convert yaml to json
	return k8syaml.YAMLToJSON(data)
}

// HarborConfigurationEmail defines the email related spec.
type HarborConfigurationEmail struct {
	// The sender name for Email notification.
	// +kubebuilder:validation:Optional
	EmailFrom string `json:"emailFrom,omitempty" yaml:"email_from,omitempty"`
	// The hostname of SMTP server that sends Email notification.
	// +kubebuilder:validation:Optional
	EmailHost string `json:"emailHost,omitempty" yaml:"email_host,omitempty"`
	// By default it's empty so the email_username is picked
	// +kubebuilder:validation:Optional
	EmailIdentity string `json:"emailIdentity,omitempty" yaml:"email_identity,omitempty"`
	// Whether or not the certificate will be verified when Harbor tries to access the email server.
	// +kubebuilder:validation:Optional
	EmailInsecure *bool `json:"emailInsecure,omitempty" yaml:"email_insecure,omitempty"`
	// The username for authenticate against SMTP server.
	// +kubebuilder:validation:Optional
	EmailUsername string `json:"emailUsername,omitempty" yaml:"email_username,omitempty"`
	// Email password.
	// +kubebuilder:validation:Optional
	EmailPassword string `json:"emailPassword,omitempty" yaml:"email_password,omitempty"`
	// The port of SMTP server.
	// +kubebuilder:validation:Optional
	EmailPort int `json:"emailPort,omitempty" yaml:"emailPort,omitempty"`
	// When it''s set to true the system will access Email server via TLS by default.  If it''s set to false, it still will handle "STARTTLS" from server side.
	// +kubebuilder:validation:Optional
	EmailSSL *bool `json:"emailSsl,omitempty" yaml:"email_ssl,omitempty"`
}

// HarborConfigurationLDAP defines the ldap related spec.
type HarborConfigurationLdap struct {
	// The Base DN for LDAP binding.
	// +kubebuilder:validation:Optional
	LdapBaseDn string `json:"ldapBaseDn,omitempty" yaml:"ldap_base_dn,omitempty"`
	// The filter for LDAP search.
	// +kubebuilder:validation:Optional
	LdapFilter string `json:"ldapFilter,omitempty" yaml:"ldap_filter,omitempty"`
	// The base DN to search LDAP group.
	// +kubebuilder:validation:Optional
	LdapGroupBaseDn string `json:"ldapGroupBaseDn,omitempty" yaml:"ldap_group_base_dn,omitempty"`
	// Specify the ldap group which have the same privilege with Harbor admin.
	// +kubebuilder:validation:Optional
	LdapGroupAdminDn string `json:"ldapGroupAdminDn,omitempty" yaml:"ldap_group_admin_dn,omitempty"`
	// The attribute which is used as identity of the LDAP group, default is cn.
	// +kubebuilder:validation:Optional
	LdapGroupAttributeName string `json:"ldapGroupAttributeName,omitempty" yaml:"ldap_group_attribute_name,omitempty"`
	// The filter to search the ldap group.
	// +kubebuilder:validation:Optional
	LdapGroupSearchFilter string `json:"ldapGroupSearchFilter,omitempty" yaml:"ldap_group_search_filter,omitempty"`
	// The scope to search ldap group. ''0-LDAP_SCOPE_BASE, 1-LDAP_SCOPE_ONELEVEL, 2-LDAP_SCOPE_SUBTREE''.
	// +kubebuilder:validation:Optional
	LdapGroupSearchScope int `json:"ldapGroupSearchScope,omitempty" yaml:"ldap_group_search_scope,omitempty"`
	// The scope to search ldap users,'0-LDAP_SCOPE_BASE, 1-LDAP_SCOPE_ONELEVEL, 2-LDAP_SCOPE_SUBTREE'.
	// +kubebuilder:validation:Optional
	LdapScope int `json:"ldapScope,omitempty" yaml:"ldap_scope,omitempty"`
	// The DN of the user to do the search.
	// +kubebuilder:validation:Optional
	LdapSearchDn string `json:"ldapSearchDn,omitempty" yaml:"ldap_search_dn,omitempty"`
	// The password ref of the ldap search dn.
	// +kubebuilder:validation:Optional
	LdapSearchPassword string `json:"ldapSearchPassword,omitempty" yaml:"ldap_search_password,omitempty"`
	// Timeout in seconds for connection to LDAP server.
	// +kubebuilder:validation:Optional
	LdapTimeout int `json:"ldapTimeout,omitempty" yaml:"ldap_timeout,omitempty"`
	// The attribute which is used as identity for the LDAP binding, such as "CN" or "SAMAccountname".
	// +kubebuilder:validation:Optional
	LdapUID string `json:"ldapUid,omitempty" yaml:"ldap_uid,omitempty"`
	// The URL of LDAP server.
	// +kubebuilder:validation:Optional
	LdapURL string `json:"ldapUrl,omitempty" yaml:"ldap_url,omitempty"`
	// Whether verify your OIDC server certificate, disable it if your OIDC server is hosted via self-hosted certificate.
	// +kubebuilder:validation:Optional
	LdapVerifyCert *bool `json:"ldapVerifyCert,omitempty" yaml:"ldap_verify_cert,omitempty"`
	// The user attribute to identify the group membership.
	// +kubebuilder:validation:Optional
	LdapGroupMembershipAttribute string `json:"ldapGroupMembershipAttribute,omitempty" yaml:"ldap_group_membership_attribute,omitempty"`
}

// HarborConfigurationUaa defines the uaa related spec.
type HarborConfigurationUaa struct {
	// The client id of UAA.
	// +kubebuilder:validation:Optional
	UaaClientID string `json:"uaaClientId,omitempty" yaml:"uaa_client_id,omitempty"`
	// The client secret of the UAA.
	// +kubebuilder:validation:Optional
	UaaClientSecret string `json:"uaaClientSecret,omitempty" yaml:"uaa_client_secret,omitempty"`
	// The endpoint of the UAA.
	// +kubebuilder:validation:Optional
	UaaEndpoint string `json:"uaaEndpoint,omitempty" yaml:"uaa_endpoint,omitempty"`
	// Verify the certificate in UAA server.
	// +kubebuilder:validation:Optional
	UaaVerifyCert *bool `json:"uaaVerifyCert,omitempty" yaml:"uaa_verify_cert,omitempty"`
}

// HarborConfigurationHTTPAuthProxy defines the http_authproxy spec.
type HarborConfigurationHTTPAuthProxy struct {
	// The endpoint of the HTTP auth.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyEndpoint string `json:"httpAuthproxyEndpoint,omitempty" yaml:"http_authproxy_endpoint,omitempty"`
	// The token review endpoint.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyTokenreviewEndpoint string `json:"httpAuthproxyTokenreviewEndpoint,omitempty" yaml:"http_authproxy_tokenreview_endpoint,omitempty"`
	// The group which has the harbor admin privileges.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyAdminGroups string `json:"httpAuthproxyAdminGroups,omitempty" yaml:"http_authproxy_admin_groups,omitempty"`
	// The username which has the harbor admin privileges.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyAdminUsernames string `json:"httpAuthproxyAdminUsernames,omitempty" yaml:"http_authproxy_admin_usernames,omitempty"`
	// Verify the HTTP auth provider's certificate.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyVerifyCert *bool `json:"httpAuthproxyVerifyCert,omitempty" yaml:"http_authproxy_verify_cert,omitempty"`
	// Search user before onboard.
	// +kubebuilder:validation:Optional
	HTTPAuthproxySkipSearch *bool `json:"httpAuthproxySkipSearch,omitempty" yaml:"http_authproxy_skip_search,omitempty"`
	// The certificate of the HTTP auth provider.
	// +kubebuilder:validation:Optional
	HTTPAuthproxyServerCertificate *bool `json:"httpAuthproxyServerCertificate,omitempty" yaml:"http_authproxy_server_certificate,omitempty"`
}

// HarborConfigurationOidc defines the oidc spec.
type HarborConfigurationOidc struct {
	// The OIDC provider name.
	// +kubebuilder:validation:Optional
	OidcName string `json:"oidcName,omitempty" yaml:"oidc_name,omitempty"`
	// The endpoint of the OIDC provider.
	// +kubebuilder:validation:Optional
	OidcEndpoint string `json:"oidcEndpoint,omitempty" yaml:"oidc_endpoint,omitempty"`
	// The client ID of the OIDC provider.
	// +kubebuilder:validation:Optional
	OidcClientID string `json:"oidcClientId,omitempty" yaml:"oidc_client_id,omitempty"`
	// The OIDC provider secret.
	// +kubebuilder:validation:Optional
	OidcClientSecret string `json:"oidcClientSecret,omitempty" yaml:"oidc_client_secret,omitempty"`
	// The attribute claims the group name.
	// +kubebuilder:validation:Optional
	OidcGroupsClaim string `json:"oidcGroupsClaim,omitempty" yaml:"oidc_groups_claim,omitempty"`
	// The OIDC group which has the harbor admin privileges.
	// +kubebuilder:validation:Optional
	OidcAdminGroup string `json:"oidcAdminGroup,omitempty" yaml:"oidc_admin_group,omitempty"`
	// The scope of the OIDC provider.
	// +kubebuilder:validation:Optional
	OidcScope string `json:"oidcScope,omitempty" yaml:"oidc_scope,omitempty"`
	// The attribute claims the username.
	// +kubebuilder:validation:Optional
	OidcUserClaim string `json:"oidcUserClaim,omitempty" yaml:"oidc_user_claim,omitempty"`
	// Verify the OIDC provider's certificate'.
	// +kubebuilder:validation:Optional
	OidcVerifyCert *bool `json:"oidcVerifyCert,omitempty" yaml:"oidc_verify_cert,omitempty"`
	// Auto onboard the OIDC user.
	// +kubebuilder:validation:Optional
	OidcAutoOnboard *bool `json:"oidcAutoOnboard,omitempty" yaml:"oidc_auto_onboard,omitempty"`
	// Extra parameters to add when redirect request to OIDC provider.
	// +kubebuilder:validation:Optional
	OidcExtraRedirectParms string `json:"oidcExtraRedirectParms,omitempty" yaml:"oidc_extra_redirect_parms,omitempty"`
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

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&HarborConfiguration{}, &HarborConfigurationList{})
}
