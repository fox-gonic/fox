package httperrors

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// New returns a new http Error object
func New(httpCode int, text string, params ...ErrParams) *Error {

	if text == "" {
		text = http.StatusText(httpCode)
	}

	err := &Error{
		HTTPCode: httpCode,
		Err:      errors.New(text),
	}

	if len(params) > 0 {
		err.Message = params[0]
	}

	return err
}

// StatusCoder is a interface for http status code
type StatusCoder interface {
	StatusCode() int
}

// ErrParams error params
type ErrParams map[string]any

// Error custom error
type Error struct {
	HTTPCode int
	Err      error
	Code     string
	Meta     any
	Message  ErrParams
}

var _ error = (*Error)(nil)

func (e *Error) Error() string {
	if e.Err == nil {
		return ""
	}
	return fmt.Sprintf("(%d): %s", e.HTTPCode, e.Err.Error())
}

// SetHTTPCode sets the error's http code.
func (e *Error) SetHTTPCode(httpCode int) *Error {
	e.HTTPCode = httpCode
	return e
}

// SetCode sets the error's code.
func (e *Error) SetCode(code string) *Error {
	e.Code = code
	return e
}

// SetMeta sets the error's meta data.
func (e *Error) SetMeta(data any) *Error {
	e.Meta = data
	return e
}

// AddMessage adds message
func (e *Error) AddMessage(key string, value any) *Error {
	if e.Message == nil {
		e.Message = ErrParams{}
	}
	e.Message[key] = value
	return e
}

// StatusCode return http status code
func (e *Error) StatusCode() int {
	return e.HTTPCode
}

// Unwrap returns the wrapped error, to allow interoperability with errors.Is(), errors.As() and errors.Unwrap()
func (e *Error) Unwrap() error {
	return e.Err
}

// JSON creates a properly formatted JSON
func (e *Error) JSON() (any, error) {

	jsonData := ErrParams{}

	if e.Meta != nil {
		value := reflect.ValueOf(e.Meta)
		switch value.Kind() {
		case reflect.Struct:
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  &jsonData,
			})
			if err != nil {
				return nil, err
			}
			if err := decoder.Decode(e.Meta); err != nil {
				return nil, err
			}

		case reflect.Map:
			for _, key := range value.MapKeys() {
				jsonData[key.String()] = value.MapIndex(key).Interface()
			}

		default:
			jsonData["message"] = e.Meta
		}
	}

	if _, exists := jsonData["code"]; !exists {
		jsonData["code"] = e.Code
	}

	if _, exists := jsonData["error"]; !exists && e.Error() != "" {
		jsonData["error"] = e.Error()
	}

	if e.Message != nil {
		if _, exists := jsonData["message"]; exists {
			if value, ok := jsonData["message"].(map[string]any); ok {
				for k, v := range e.Message {
					value[k] = v
				}
			}
		} else {
			jsonData["message"] = e.Message
		}
	}

	return jsonData, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (e Error) MarshalJSON() ([]byte, error) {
	v, err := e.JSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// GenerateUnknownError return a unknown error
func GenerateUnknownError(err error, httpCode ...int) *Error {
	var code = 500
	if len(httpCode) > 0 {
		code = httpCode[0]
	}

	return &Error{
		HTTPCode: code,
		Err:      err,
	}
}

// As is errors.As
func As(err error) (t *Error, ok bool) {
	return t, errors.As(err, &t)
}

// Wrap httperrors wrapper helper
func Wrap(err error, httpCode ...int) (e *Error) {
	if err == nil {
		return nil
	}

	if errors.As(err, &e) {
		return e
	}

	return GenerateUnknownError(err, httpCode...)
}
