// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// RetentionRuleParamMetadata rule param
//
// swagger:model RetentionRuleParamMetadata
type RetentionRuleParamMetadata struct {

	// required
	Required bool `json:"required,omitempty"`

	// type
	Type string `json:"type,omitempty"`

	// unit
	Unit string `json:"unit,omitempty"`
}

// Validate validates this retention rule param metadata
func (m *RetentionRuleParamMetadata) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *RetentionRuleParamMetadata) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RetentionRuleParamMetadata) UnmarshalBinary(b []byte) error {
	var res RetentionRuleParamMetadata
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
