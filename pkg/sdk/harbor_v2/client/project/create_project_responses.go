// Code generated by go-swagger; DO NOT EDIT.

package project

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor-operator/pkg/sdk/harbor_v2/models"
)

// CreateProjectReader is a Reader for the CreateProject structure.
type CreateProjectReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateProjectReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewCreateProjectCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewCreateProjectBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewCreateProjectUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 409:
		result := NewCreateProjectConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewCreateProjectInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewCreateProjectCreated creates a CreateProjectCreated with default headers values
func NewCreateProjectCreated() *CreateProjectCreated {
	return &CreateProjectCreated{}
}

/*CreateProjectCreated handles this case with default header values.

Created
*/
type CreateProjectCreated struct {
	/*The location of the resource
	 */
	Location string
	/*The ID of the corresponding request for the response
	 */
	XRequestID string
}

func (o *CreateProjectCreated) Error() string {
	return fmt.Sprintf("[POST /projects][%d] createProjectCreated ", 201)
}

func (o *CreateProjectCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header Location
	o.Location = response.GetHeader("Location")

	// response header X-Request-Id
	o.XRequestID = response.GetHeader("X-Request-Id")

	return nil
}

// NewCreateProjectBadRequest creates a CreateProjectBadRequest with default headers values
func NewCreateProjectBadRequest() *CreateProjectBadRequest {
	return &CreateProjectBadRequest{}
}

/*CreateProjectBadRequest handles this case with default header values.

Bad request
*/
type CreateProjectBadRequest struct {
	/*The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

func (o *CreateProjectBadRequest) Error() string {
	return fmt.Sprintf("[POST /projects][%d] createProjectBadRequest  %+v", 400, o.Payload)
}

func (o *CreateProjectBadRequest) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateProjectBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-Request-Id
	o.XRequestID = response.GetHeader("X-Request-Id")

	o.Payload = new(models.Errors)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateProjectUnauthorized creates a CreateProjectUnauthorized with default headers values
func NewCreateProjectUnauthorized() *CreateProjectUnauthorized {
	return &CreateProjectUnauthorized{}
}

/*CreateProjectUnauthorized handles this case with default header values.

Unauthorized
*/
type CreateProjectUnauthorized struct {
	/*The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

func (o *CreateProjectUnauthorized) Error() string {
	return fmt.Sprintf("[POST /projects][%d] createProjectUnauthorized  %+v", 401, o.Payload)
}

func (o *CreateProjectUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateProjectUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-Request-Id
	o.XRequestID = response.GetHeader("X-Request-Id")

	o.Payload = new(models.Errors)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateProjectConflict creates a CreateProjectConflict with default headers values
func NewCreateProjectConflict() *CreateProjectConflict {
	return &CreateProjectConflict{}
}

/*CreateProjectConflict handles this case with default header values.

Conflict
*/
type CreateProjectConflict struct {
	/*The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

func (o *CreateProjectConflict) Error() string {
	return fmt.Sprintf("[POST /projects][%d] createProjectConflict  %+v", 409, o.Payload)
}

func (o *CreateProjectConflict) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateProjectConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-Request-Id
	o.XRequestID = response.GetHeader("X-Request-Id")

	o.Payload = new(models.Errors)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateProjectInternalServerError creates a CreateProjectInternalServerError with default headers values
func NewCreateProjectInternalServerError() *CreateProjectInternalServerError {
	return &CreateProjectInternalServerError{}
}

/*CreateProjectInternalServerError handles this case with default header values.

Internal server error
*/
type CreateProjectInternalServerError struct {
	/*The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

func (o *CreateProjectInternalServerError) Error() string {
	return fmt.Sprintf("[POST /projects][%d] createProjectInternalServerError  %+v", 500, o.Payload)
}

func (o *CreateProjectInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateProjectInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header X-Request-Id
	o.XRequestID = response.GetHeader("X-Request-Id")

	o.Payload = new(models.Errors)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
