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
func New(httpCode int, text string) *Error {

	if text == "" {
		text = http.StatusText(httpCode)
	}

	return &Error{
		HTTPCode: httpCode,
		Err:      errors.New(text),
	}
}

// Error custom error
type Error struct {
	HTTPCode int
	Err      error
	Code     string

	Meta   any
	Fields map[string]any
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

// AddField adds field
func (e *Error) AddField(key string, value any) *Error {
	if e.Fields == nil {
		e.Fields = map[string]any{}
	}
	e.Fields[key] = value
	return e
}

// AddFields adds fields
func (e *Error) AddFields(fields map[string]any) *Error {
	if e.Fields == nil {
		e.Fields = map[string]any{}
	}
	for key, value := range fields {
		e.Fields[key] = value
	}
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

// MarshalJSON implements the json.Marshaler interface.
func (e Error) MarshalJSON() ([]byte, error) {
	jsonData := map[string]any{}

	if e.Meta == nil {
		e.Meta = e.Err
	}

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
			if _, ok := e.Meta.(error); !ok {
				jsonData["meta"] = e.Meta
			}
		}
	}

	if _, exists := jsonData["code"]; !exists {
		jsonData["code"] = e.Code
	}

	if _, exists := jsonData["error"]; !exists && e.Error() != "" {
		jsonData["error"] = e.Error()
	}

	if e.Fields != nil {
		for key, value := range e.Fields {
			jsonData[key] = value
		}
	}

	return json.Marshal(jsonData)
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

	var code int
	if len(httpCode) > 0 {
		code = httpCode[0]
	}

	return &Error{
		HTTPCode: code,
		Err:      err,
	}
}
