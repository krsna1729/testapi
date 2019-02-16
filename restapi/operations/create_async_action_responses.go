// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/mcastelino/testapi/models"
)

// CreateAsyncActionNoContentCode is the HTTP code returned for type CreateAsyncActionNoContent
const CreateAsyncActionNoContentCode int = 204

/*CreateAsyncActionNoContent The update was successful

swagger:response createAsyncActionNoContent
*/
type CreateAsyncActionNoContent struct {
}

// NewCreateAsyncActionNoContent creates CreateAsyncActionNoContent with default headers values
func NewCreateAsyncActionNoContent() *CreateAsyncActionNoContent {

	return &CreateAsyncActionNoContent{}
}

// WriteResponse to the client
func (o *CreateAsyncActionNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// CreateAsyncActionBadRequestCode is the HTTP code returned for type CreateAsyncActionBadRequest
const CreateAsyncActionBadRequestCode int = 400

/*CreateAsyncActionBadRequest The action cannot be executed due to bad input

swagger:response createAsyncActionBadRequest
*/
type CreateAsyncActionBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreateAsyncActionBadRequest creates CreateAsyncActionBadRequest with default headers values
func NewCreateAsyncActionBadRequest() *CreateAsyncActionBadRequest {

	return &CreateAsyncActionBadRequest{}
}

// WithPayload adds the payload to the create async action bad request response
func (o *CreateAsyncActionBadRequest) WithPayload(payload *models.Error) *CreateAsyncActionBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create async action bad request response
func (o *CreateAsyncActionBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAsyncActionBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*CreateAsyncActionDefault Internal Server Error

swagger:response createAsyncActionDefault
*/
type CreateAsyncActionDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreateAsyncActionDefault creates CreateAsyncActionDefault with default headers values
func NewCreateAsyncActionDefault(code int) *CreateAsyncActionDefault {
	if code <= 0 {
		code = 500
	}

	return &CreateAsyncActionDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the create async action default response
func (o *CreateAsyncActionDefault) WithStatusCode(code int) *CreateAsyncActionDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the create async action default response
func (o *CreateAsyncActionDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the create async action default response
func (o *CreateAsyncActionDefault) WithPayload(payload *models.Error) *CreateAsyncActionDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create async action default response
func (o *CreateAsyncActionDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAsyncActionDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}