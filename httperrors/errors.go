package httperrors

import (
	"errors"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
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
	Message  ErrParams
}

// JSON custom error json struct
type JSON struct {
	HTTPCode int       `json:"-"`
	Err      error     `json:"-"`
	Code     string    `json:"code"`
	Message  ErrParams `json:"message"`
}

func (e *Error) Error() string {
	if len(e.Message) == 0 {
		return fmt.Sprintf("(%d): %s", e.HTTPCode, e.Err.Error())
	}

	message, _ := json.Marshal(e.Message)
	return fmt.Sprintf("(%d): %s %s", e.HTTPCode, e.Err.Error(), string(message))
}

// StatusCode return http status code
func (e *Error) StatusCode() int {
	return e.HTTPCode
}

// Unwrap method
func (e *Error) Unwrap() error { return e.Err }

// MarshalJSON implements the json.Marshaler interface.
func (e Error) MarshalJSON() ([]byte, error) {
	// TODO: exchange e.Code
	if len(e.Code) == 0 {
		e.Code = "UNKNOW_ERROR"
	}

	if len(e.Message) == 0 {
		e.Message = ErrParams{"error": e.Err.Error()}
	}

	// return json.Marshal(JSON{
	// 	HTTPCode: e.HTTPCode,
	// 	Err:      e.Err,
	// 	Code:     e.Code,
	// 	Message:  e.Message,
	// })

	return json.Marshal(map[string]interface{}{
		"code":    e.Code,
		"message": e.Message,
	})
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
