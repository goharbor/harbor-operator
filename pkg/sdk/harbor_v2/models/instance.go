// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// Instance instance
//
// swagger:model Instance
type Instance struct {

	// The auth credential data if exists
	AuthInfo map[string]string `json:"auth_info,omitempty"`

	// The authentication way supported
	AuthMode string `json:"auth_mode,omitempty"`

	// Whether the instance is default or not
	Default bool `json:"default"`

	// Description of instance
	Description string `json:"description,omitempty"`

	// Whether the instance is activated or not
	Enabled bool `json:"enabled"`

	// The service endpoint of this instance
	Endpoint string `json:"endpoint,omitempty"`

	// Unique ID
	ID int64 `json:"id,omitempty"`

	// Whether the instance endpoint is insecure or not
	Insecure bool `json:"insecure"`

	// Instance name
	Name string `json:"name,omitempty"`

	// The timestamp of instance setting up
	SetupTimestamp int64 `json:"setup_timestamp,omitempty"`

	// The health status
	Status string `json:"status,omitempty"`

	// Based on which driver, identified by ID
	Vendor string `json:"vendor,omitempty"`
}

// Validate validates this instance
func (m *Instance) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Instance) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Instance) UnmarshalBinary(b []byte) error {
	var res Instance
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
