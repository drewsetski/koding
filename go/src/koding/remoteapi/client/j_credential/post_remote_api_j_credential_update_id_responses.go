package j_credential

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"

	strfmt "github.com/go-openapi/strfmt"

	"koding/remoteapi/models"
)

// PostRemoteAPIJCredentialUpdateIDReader is a Reader for the PostRemoteAPIJCredentialUpdateID structure.
type PostRemoteAPIJCredentialUpdateIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PostRemoteAPIJCredentialUpdateIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewPostRemoteAPIJCredentialUpdateIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewPostRemoteAPIJCredentialUpdateIDOK creates a PostRemoteAPIJCredentialUpdateIDOK with default headers values
func NewPostRemoteAPIJCredentialUpdateIDOK() *PostRemoteAPIJCredentialUpdateIDOK {
	return &PostRemoteAPIJCredentialUpdateIDOK{}
}

/*PostRemoteAPIJCredentialUpdateIDOK handles this case with default header values.

OK
*/
type PostRemoteAPIJCredentialUpdateIDOK struct {
	Payload PostRemoteAPIJCredentialUpdateIDOKBody
}

func (o *PostRemoteAPIJCredentialUpdateIDOK) Error() string {
	return fmt.Sprintf("[POST /remote.api/JCredential.update/{id}][%d] postRemoteApiJCredentialUpdateIdOK  %+v", 200, o.Payload)
}

func (o *PostRemoteAPIJCredentialUpdateIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*PostRemoteAPIJCredentialUpdateIDOKBody post remote API j credential update ID o k body
swagger:model PostRemoteAPIJCredentialUpdateIDOKBody
*/
type PostRemoteAPIJCredentialUpdateIDOKBody struct {
	models.JCredential

	models.DefaultResponse
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (o *PostRemoteAPIJCredentialUpdateIDOKBody) UnmarshalJSON(raw []byte) error {

	var postRemoteAPIJCredentialUpdateIDOKBodyAO0 models.JCredential
	if err := swag.ReadJSON(raw, &postRemoteAPIJCredentialUpdateIDOKBodyAO0); err != nil {
		return err
	}
	o.JCredential = postRemoteAPIJCredentialUpdateIDOKBodyAO0

	var postRemoteAPIJCredentialUpdateIDOKBodyAO1 models.DefaultResponse
	if err := swag.ReadJSON(raw, &postRemoteAPIJCredentialUpdateIDOKBodyAO1); err != nil {
		return err
	}
	o.DefaultResponse = postRemoteAPIJCredentialUpdateIDOKBodyAO1

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (o PostRemoteAPIJCredentialUpdateIDOKBody) MarshalJSON() ([]byte, error) {
	var _parts [][]byte

	postRemoteAPIJCredentialUpdateIDOKBodyAO0, err := swag.WriteJSON(o.JCredential)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, postRemoteAPIJCredentialUpdateIDOKBodyAO0)

	postRemoteAPIJCredentialUpdateIDOKBodyAO1, err := swag.WriteJSON(o.DefaultResponse)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, postRemoteAPIJCredentialUpdateIDOKBodyAO1)

	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this post remote API j credential update ID o k body
func (o *PostRemoteAPIJCredentialUpdateIDOKBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.JCredential.Validate(formats); err != nil {
		res = append(res, err)
	}

	if err := o.DefaultResponse.Validate(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
