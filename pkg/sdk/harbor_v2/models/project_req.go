// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ProjectReq project req
//
// swagger:model ProjectReq
type ProjectReq struct {

	// The CVE allowlist of the project.
	CveAllowlist *CVEAllowlist `json:"cve_allowlist,omitempty"`

	// The metadata of the project.
	Metadata *ProjectMetadata `json:"metadata,omitempty"`

	// The name of the project.
	ProjectName string `json:"project_name,omitempty"`

	// deprecated, reserved for project creation in replication
	Public *bool `json:"public,omitempty"`

	// The ID of referenced registry when creating the proxy cache project
	RegistryID *int64 `json:"registry_id,omitempty"`

	// The storage quota of the project.
	StorageLimit *int64 `json:"storage_limit,omitempty"`
}

// Validate validates this project req
func (m *ProjectReq) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCveAllowlist(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMetadata(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ProjectReq) validateCveAllowlist(formats strfmt.Registry) error {

	if swag.IsZero(m.CveAllowlist) { // not required
		return nil
	}

	if m.CveAllowlist != nil {
		if err := m.CveAllowlist.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cve_allowlist")
			}
			return err
		}
	}

	return nil
}

func (m *ProjectReq) validateMetadata(formats strfmt.Registry) error {

	if swag.IsZero(m.Metadata) { // not required
		return nil
	}

	if m.Metadata != nil {
		if err := m.Metadata.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("metadata")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *ProjectReq) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProjectReq) UnmarshalBinary(b []byte) error {
	var res ProjectReq
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
