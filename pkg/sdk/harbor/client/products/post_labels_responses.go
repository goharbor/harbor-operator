// Code generated by go-swagger; DO NOT EDIT.

package products

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// PostLabelsReader is a Reader for the PostLabels structure.
type PostLabelsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PostLabelsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewPostLabelsCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewPostLabelsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewPostLabelsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 409:
		result := NewPostLabelsConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 415:
		result := NewPostLabelsUnsupportedMediaType()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewPostLabelsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewPostLabelsCreated creates a PostLabelsCreated with default headers values
func NewPostLabelsCreated() *PostLabelsCreated {
	return &PostLabelsCreated{}
}

/*PostLabelsCreated handles this case with default header values.

Create successfully.
*/
type PostLabelsCreated struct {
	/*The URL of the created resource
	 */
	Location string
}

func (o *PostLabelsCreated) Error() string {
	return fmt.Sprintf("[POST /labels][%d] postLabelsCreated ", 201)
}

func (o *PostLabelsCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header Location
	o.Location = response.GetHeader("Location")

	return nil
}

// NewPostLabelsBadRequest creates a PostLabelsBadRequest with default headers values
func NewPostLabelsBadRequest() *PostLabelsBadRequest {
	return &PostLabelsBadRequest{}
}

/*PostLabelsBadRequest handles this case with default header values.

Invalid parameters.
*/
type PostLabelsBadRequest struct {
}

func (o *PostLabelsBadRequest) Error() string {
	return fmt.Sprintf("[POST /labels][%d] postLabelsBadRequest ", 400)
}

func (o *PostLabelsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPostLabelsUnauthorized creates a PostLabelsUnauthorized with default headers values
func NewPostLabelsUnauthorized() *PostLabelsUnauthorized {
	return &PostLabelsUnauthorized{}
}

/*PostLabelsUnauthorized handles this case with default header values.

User need to log in first.
*/
type PostLabelsUnauthorized struct {
}

func (o *PostLabelsUnauthorized) Error() string {
	return fmt.Sprintf("[POST /labels][%d] postLabelsUnauthorized ", 401)
}

func (o *PostLabelsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPostLabelsConflict creates a PostLabelsConflict with default headers values
func NewPostLabelsConflict() *PostLabelsConflict {
	return &PostLabelsConflict{}
}

/*PostLabelsConflict handles this case with default header values.

Label with the same name and same scope already exists.
*/
type PostLabelsConflict struct {
}

func (o *PostLabelsConflict) Error() string {
	return fmt.Sprintf("[POST /labels][%d] postLabelsConflict ", 409)
}

func (o *PostLabelsConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPostLabelsUnsupportedMediaType creates a PostLabelsUnsupportedMediaType with default headers values
func NewPostLabelsUnsupportedMediaType() *PostLabelsUnsupportedMediaType {
	return &PostLabelsUnsupportedMediaType{}
}

/*PostLabelsUnsupportedMediaType handles this case with default header values.

The Media Type of the request is not supported, it has to be "application/json"
*/
type PostLabelsUnsupportedMediaType struct {
}

func (o *PostLabelsUnsupportedMediaType) Error() string {
	return fmt.Sprintf("[POST /labels][%d] postLabelsUnsupportedMediaType ", 415)
}

func (o *PostLabelsUnsupportedMediaType) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPostLabelsInternalServerError creates a PostLabelsInternalServerError with default headers values
func NewPostLabelsInternalServerError() *PostLabelsInternalServerError {
	return &PostLabelsInternalServerError{}
}

/*PostLabelsInternalServerError handles this case with default header values.

Unexpected internal errors.
*/
type PostLabelsInternalServerError struct {
}

func (o *PostLabelsInternalServerError) Error() string {
	return fmt.Sprintf("[POST /labels][%d] postLabelsInternalServerError ", 500)
}

func (o *PostLabelsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
