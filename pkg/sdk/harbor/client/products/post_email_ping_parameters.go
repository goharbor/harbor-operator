// Code generated by go-swagger; DO NOT EDIT.

package products

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor-operator/pkg/sdk/harbor/models"
)

// NewPostEmailPingParams creates a new PostEmailPingParams object
// with the default values initialized.
func NewPostEmailPingParams() *PostEmailPingParams {
	var ()
	return &PostEmailPingParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewPostEmailPingParamsWithTimeout creates a new PostEmailPingParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewPostEmailPingParamsWithTimeout(timeout time.Duration) *PostEmailPingParams {
	var ()
	return &PostEmailPingParams{

		timeout: timeout,
	}
}

// NewPostEmailPingParamsWithContext creates a new PostEmailPingParams object
// with the default values initialized, and the ability to set a context for a request
func NewPostEmailPingParamsWithContext(ctx context.Context) *PostEmailPingParams {
	var ()
	return &PostEmailPingParams{

		Context: ctx,
	}
}

// NewPostEmailPingParamsWithHTTPClient creates a new PostEmailPingParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewPostEmailPingParamsWithHTTPClient(client *http.Client) *PostEmailPingParams {
	var ()
	return &PostEmailPingParams{
		HTTPClient: client,
	}
}

/*PostEmailPingParams contains all the parameters to send to the API endpoint
for the post email ping operation typically these are written to a http.Request
*/
type PostEmailPingParams struct {

	/*Settings
	  Email server settings, if some of the settings are not assigned, they will be read from system configuration.

	*/
	Settings *models.EmailServerSetting

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the post email ping params
func (o *PostEmailPingParams) WithTimeout(timeout time.Duration) *PostEmailPingParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the post email ping params
func (o *PostEmailPingParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the post email ping params
func (o *PostEmailPingParams) WithContext(ctx context.Context) *PostEmailPingParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the post email ping params
func (o *PostEmailPingParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the post email ping params
func (o *PostEmailPingParams) WithHTTPClient(client *http.Client) *PostEmailPingParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the post email ping params
func (o *PostEmailPingParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithSettings adds the settings to the post email ping params
func (o *PostEmailPingParams) WithSettings(settings *models.EmailServerSetting) *PostEmailPingParams {
	o.SetSettings(settings)
	return o
}

// SetSettings adds the settings to the post email ping params
func (o *PostEmailPingParams) SetSettings(settings *models.EmailServerSetting) {
	o.Settings = settings
}

// WriteToRequest writes these params to a swagger request
func (o *PostEmailPingParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Settings != nil {
		if err := r.SetBodyParam(o.Settings); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
