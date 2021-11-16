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
	"github.com/go-openapi/swag"

	"github.com/goharbor/harbor-operator/pkg/sdk/harbor/models"
)

// NewPostProjectsProjectIDWebhookPoliciesTestParams creates a new PostProjectsProjectIDWebhookPoliciesTestParams object
// with the default values initialized.
func NewPostProjectsProjectIDWebhookPoliciesTestParams() *PostProjectsProjectIDWebhookPoliciesTestParams {
	var ()
	return &PostProjectsProjectIDWebhookPoliciesTestParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewPostProjectsProjectIDWebhookPoliciesTestParamsWithTimeout creates a new PostProjectsProjectIDWebhookPoliciesTestParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewPostProjectsProjectIDWebhookPoliciesTestParamsWithTimeout(timeout time.Duration) *PostProjectsProjectIDWebhookPoliciesTestParams {
	var ()
	return &PostProjectsProjectIDWebhookPoliciesTestParams{

		timeout: timeout,
	}
}

// NewPostProjectsProjectIDWebhookPoliciesTestParamsWithContext creates a new PostProjectsProjectIDWebhookPoliciesTestParams object
// with the default values initialized, and the ability to set a context for a request
func NewPostProjectsProjectIDWebhookPoliciesTestParamsWithContext(ctx context.Context) *PostProjectsProjectIDWebhookPoliciesTestParams {
	var ()
	return &PostProjectsProjectIDWebhookPoliciesTestParams{

		Context: ctx,
	}
}

// NewPostProjectsProjectIDWebhookPoliciesTestParamsWithHTTPClient creates a new PostProjectsProjectIDWebhookPoliciesTestParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewPostProjectsProjectIDWebhookPoliciesTestParamsWithHTTPClient(client *http.Client) *PostProjectsProjectIDWebhookPoliciesTestParams {
	var ()
	return &PostProjectsProjectIDWebhookPoliciesTestParams{
		HTTPClient: client,
	}
}

/*PostProjectsProjectIDWebhookPoliciesTestParams contains all the parameters to send to the API endpoint
for the post projects project ID webhook policies test operation typically these are written to a http.Request
*/
type PostProjectsProjectIDWebhookPoliciesTestParams struct {

	/*Policy
	  Only property "targets" needed.

	*/
	Policy *models.WebhookPolicy
	/*ProjectID
	  Relevant project ID.

	*/
	ProjectID int64

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) WithTimeout(timeout time.Duration) *PostProjectsProjectIDWebhookPoliciesTestParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) WithContext(ctx context.Context) *PostProjectsProjectIDWebhookPoliciesTestParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) WithHTTPClient(client *http.Client) *PostProjectsProjectIDWebhookPoliciesTestParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithPolicy adds the policy to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) WithPolicy(policy *models.WebhookPolicy) *PostProjectsProjectIDWebhookPoliciesTestParams {
	o.SetPolicy(policy)
	return o
}

// SetPolicy adds the policy to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) SetPolicy(policy *models.WebhookPolicy) {
	o.Policy = policy
}

// WithProjectID adds the projectID to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) WithProjectID(projectID int64) *PostProjectsProjectIDWebhookPoliciesTestParams {
	o.SetProjectID(projectID)
	return o
}

// SetProjectID adds the projectId to the post projects project ID webhook policies test params
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) SetProjectID(projectID int64) {
	o.ProjectID = projectID
}

// WriteToRequest writes these params to a swagger request
func (o *PostProjectsProjectIDWebhookPoliciesTestParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Policy != nil {
		if err := r.SetBodyParam(o.Policy); err != nil {
			return err
		}
	}

	// path param project_id
	if err := r.SetPathParam("project_id", swag.FormatInt64(o.ProjectID)); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
