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

// --------------------------------------------------------------------

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

// As is errors.As
func As(err error) (*Error, bool) {
	var t *Error
	ok := errors.As(err, &t)
	return t, ok
}

// --------------------------------------------------------------------

// StatusCoder is a interface for http status code
type StatusCoder interface {
	StatusCode() int
}

// --------------------------------------------------------------------

// ErrParams error params
type ErrParams map[string]interface{}

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
	return fmt.Sprintf("(%d): %s", e.HTTPCode, e.Err.Error())
}

// SetMeta sets the error's meta data.
func (e *Error) SetMeta(data any) *Error {
	e.Meta = data
	return e
}

// AddMessage adds message
func (e *Error) AddMessage(key string, value any) {
	if e.Message == nil {
		e.Message = ErrParams{}
	}
	e.Message[key] = value
}

// StatusCode return http status code
func (e *Error) StatusCode() int {
	return e.HTTPCode
}

// Unwrap method
func (e *Error) Unwrap() error { return e.Err }

// JSON creates a properly formatted JSON
func (e *Error) JSON() (any, error) {

	if len(e.Code) == 0 {
		e.Code = "UNKNOW_ERROR"
	}

	jsonData := map[string]any{}

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

	if _, exists := jsonData["error"]; !exists {
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

// --------------------------------------------------------------------

// GenerateUnknownError return a unknown error
func GenerateUnknownError(err error, httpCode ...int) *Error {
	var code = 500
	if len(httpCode) > 0 {
		code = httpCode[0]
	}

	return &Error{
		HTTPCode: code,
		Err:      err,
		Message:  ErrParams{"error": err.Error()},
	}
}

// --------------------------------------------------------------------

// Wrap booboo error wrapper helper
func Wrap(err error, httpCode ...int) (e *Error) {
	if err == nil {
		return nil
	}

	if errors.As(err, &e) {
		return e
	}

	return GenerateUnknownError(err, httpCode...)
}

// --------------------------------------------------------------------

// GetErrorWithMessage new error from exist error, but you need a data
func GetErrorWithMessage(err Error, datas ...ErrParams) *Error {
	if len(datas) != 0 {
		err.Message = datas[0]
	}

	return &err
}

// GetErrorWithKV new error from exist and modify message with key/value
func GetErrorWithKV(err Error, key string, value interface{}) *Error {
	errParams := ErrParams{key: value}
	err.Message = errParams

	return &err
}

// GetErrorWithKVPairs new error from exist error and new message with key/value pairs
func GetErrorWithKVPairs(err Error, keys []string, values []interface{}) *Error {
	if len(keys) != len(values) {
		panic("error len(keys) != len(values)")
	}

	errParams := ErrParams{}
	for i := range keys {
		errParams[keys[i]] = values[i]
	}

	err.Message = errParams

	return &err

}
