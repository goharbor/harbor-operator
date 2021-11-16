// Code generated by go-swagger; DO NOT EDIT.

package products

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// PostProjectsProjectIDWebhookPoliciesReader is a Reader for the PostProjectsProjectIDWebhookPolicies structure.
type PostProjectsProjectIDWebhookPoliciesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PostProjectsProjectIDWebhookPoliciesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewPostProjectsProjectIDWebhookPoliciesCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewPostProjectsProjectIDWebhookPoliciesBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewPostProjectsProjectIDWebhookPoliciesUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewPostProjectsProjectIDWebhookPoliciesForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewPostProjectsProjectIDWebhookPoliciesInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewPostProjectsProjectIDWebhookPoliciesCreated creates a PostProjectsProjectIDWebhookPoliciesCreated with default headers values
func NewPostProjectsProjectIDWebhookPoliciesCreated() *PostProjectsProjectIDWebhookPoliciesCreated {
	return &PostProjectsProjectIDWebhookPoliciesCreated{}
}

/*PostProjectsProjectIDWebhookPoliciesCreated handles this case with default header values.

Project webhook policy create successfully.
*/
type PostProjectsProjectIDWebhookPoliciesCreated struct {
	/*The URL of the created resource
	 */
	Location string
}

func (o *PostProjectsProjectIDWebhookPoliciesCreated) Error() string {
	return fmt.Sprintf("[POST /projects/{project_id}/webhook/policies][%d] postProjectsProjectIdWebhookPoliciesCreated ", 201)
}

func (o *PostProjectsProjectIDWebhookPoliciesCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header Location
	o.Location = response.GetHeader("Location")

	return nil
}

// NewPostProjectsProjectIDWebhookPoliciesBadRequest creates a PostProjectsProjectIDWebhookPoliciesBadRequest with default headers values
func NewPostProjectsProjectIDWebhookPoliciesBadRequest() *PostProjectsProjectIDWebhookPoliciesBadRequest {
	return &PostProjectsProjectIDWebhookPoliciesBadRequest{}
}

/*PostProjectsProjectIDWebhookPoliciesBadRequest handles this case with default header values.

Illegal format of provided ID value.
*/
type PostProjectsProjectIDWebhookPoliciesBadRequest struct {
}

func (o *PostProjectsProjectIDWebhookPoliciesBadRequest) Error() string {
	return fmt.Sprintf("[POST /projects/{project_id}/webhook/policies][%d] postProjectsProjectIdWebhookPoliciesBadRequest ", 400)
}

func (o *PostProjectsProjectIDWebhookPoliciesBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPostProjectsProjectIDWebhookPoliciesUnauthorized creates a PostProjectsProjectIDWebhookPoliciesUnauthorized with default headers values
func NewPostProjectsProjectIDWebhookPoliciesUnauthorized() *PostProjectsProjectIDWebhookPoliciesUnauthorized {
	return &PostProjectsProjectIDWebhookPoliciesUnauthorized{}
}

/*PostProjectsProjectIDWebhookPoliciesUnauthorized handles this case with default header values.

User need to log in first.
*/
type PostProjectsProjectIDWebhookPoliciesUnauthorized struct {
}

func (o *PostProjectsProjectIDWebhookPoliciesUnauthorized) Error() string {
	return fmt.Sprintf("[POST /projects/{project_id}/webhook/policies][%d] postProjectsProjectIdWebhookPoliciesUnauthorized ", 401)
}

func (o *PostProjectsProjectIDWebhookPoliciesUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPostProjectsProjectIDWebhookPoliciesForbidden creates a PostProjectsProjectIDWebhookPoliciesForbidden with default headers values
func NewPostProjectsProjectIDWebhookPoliciesForbidden() *PostProjectsProjectIDWebhookPoliciesForbidden {
	return &PostProjectsProjectIDWebhookPoliciesForbidden{}
}

/*PostProjectsProjectIDWebhookPoliciesForbidden handles this case with default header values.

User have no permission to create webhook policy of the project.
*/
type PostProjectsProjectIDWebhookPoliciesForbidden struct {
}

func (o *PostProjectsProjectIDWebhookPoliciesForbidden) Error() string {
	return fmt.Sprintf("[POST /projects/{project_id}/webhook/policies][%d] postProjectsProjectIdWebhookPoliciesForbidden ", 403)
}

func (o *PostProjectsProjectIDWebhookPoliciesForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPostProjectsProjectIDWebhookPoliciesInternalServerError creates a PostProjectsProjectIDWebhookPoliciesInternalServerError with default headers values
func NewPostProjectsProjectIDWebhookPoliciesInternalServerError() *PostProjectsProjectIDWebhookPoliciesInternalServerError {
	return &PostProjectsProjectIDWebhookPoliciesInternalServerError{}
}

/*PostProjectsProjectIDWebhookPoliciesInternalServerError handles this case with default header values.

Unexpected internal errors.
*/
type PostProjectsProjectIDWebhookPoliciesInternalServerError struct {
}

func (o *PostProjectsProjectIDWebhookPoliciesInternalServerError) Error() string {
	return fmt.Sprintf("[POST /projects/{project_id}/webhook/policies][%d] postProjectsProjectIdWebhookPoliciesInternalServerError ", 500)
}

func (o *PostProjectsProjectIDWebhookPoliciesInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
