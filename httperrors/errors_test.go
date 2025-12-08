package httperrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type E string

func (e E) Error() string {
	return string(e)
}

type ErrorInfo struct {
	Err            string   `json:"error"`
	Reqid          string   `json:"reqid"`
	Details        []string `json:"details"`
	Code           int      `json:"code"`
	ErrCode        string   `json:"error_code,omitempty"`
	OmitEmptyField string   `json:"omit_empty_field,omitempty"`
	IgnoreField    string   `json:"-"`
}

func TestError(t *testing.T) {
	r := require.New(t)
	err := New(400, "invalid arguments")
	r.Equal(400, err.HTTPCode)
	r.Equal(400, err.StatusCode())

	r.Equal("(400): invalid arguments", err.Error())
	r.Equal(errors.New("invalid arguments"), err.Unwrap())

	{
		data, e := json.Marshal(err)
		r.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		r.NoError(e)
		r.Equal("(400): invalid arguments", obj["error"])
	}

	{
		_ = err.SetMeta(ErrorInfo{
			Err:   "invalid arguments",
			Reqid: "F4CD:20C1B9:2894CD0:3468624:6692A040",
			Details: []string{
				"title field is required",
				"content field is required",
			},
			Code:           400,
			ErrCode:        "invalid_arguments",
			OmitEmptyField: "",
			IgnoreField:    "IgnoreField",
		})

		_ = err.AddField("x-request-id", "hvnmjnCVyvQ3aOIX")
		_ = err.AddFields(map[string]any{
			"latency": 1.799868,
			"ftype":   "HANDLER",
		})

		r.True(err.HasField("x-request-id"))
		r.False(err.HasField("x-access-key"))

		data, e := json.Marshal(err)
		r.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		r.NoError(e)
		r.Equal("invalid arguments", obj["error"])
		r.Equal("F4CD:20C1B9:2894CD0:3468624:6692A040", obj["reqid"])
		r.Equal([]any{
			"title field is required",
			"content field is required",
		}, obj["details"])
		r.Equal("400", fmt.Sprintf("%v", obj["code"]))
		r.Equal("invalid_arguments", obj["error_code"])
		r.Empty(obj["omit_empty_field"])
		r.Empty(obj["ignore_field"])
		r.Equal("hvnmjnCVyvQ3aOIX", obj["x-request-id"])
		r.Equal("1.799868", fmt.Sprintf("%v", obj["latency"]))
		r.Equal("HANDLER", obj["ftype"])
	}

	{
		err2 := err.Clone()
		_ = err2.SetMeta(map[string]any{
			"error": "invalid arguments",
			"reqid": "F4CD:20C1B9:2894CD0:3468624:6692A040",
			"details": []string{
				"title field is required",
				"content field is required",
			},
		})

		_ = err2.SetCode("INVALID_ARGUMENTS")

		data, e := json.Marshal(err2)
		r.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		r.NoError(e)
		r.Equal("invalid arguments", obj["error"])
		r.Equal("F4CD:20C1B9:2894CD0:3468624:6692A040", obj["reqid"])
		r.Equal([]any{
			"title field is required",
			"content field is required",
		}, obj["details"])
		r.Equal("INVALID_ARGUMENTS", fmt.Sprintf("%v", obj["code"]))
	}

	{
		e := New(400, "")
		r.Equal(400, e.HTTPCode)
		r.Equal(400, e.StatusCode())
		r.Equal("(400): Bad Request", e.Error())
	}

	{
		err2 := &Error{}
		r.Empty(err2.Error())
		_ = err2.SetHTTPCode(500)
		_ = err2.SetCode("internal_error")

		r.Equal(500, err2.HTTPCode)
		r.Equal("internal_error", err2.Code)

		r.Nil(err2.Fields)
		r.False(err2.HasField("foo"))

		_ = err2.AddFields(map[string]any{"foo": "bar"})
		r.NotNil(err2.Fields)
		r.Equal(map[string]any{"foo": "bar"}, err2.Fields)

		_ = err2.SetMeta(E("database internal error"))
		data, e := json.Marshal(err2)
		r.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		r.NoError(e)
		r.Equal("database internal error", obj["meta"])

		_ = err2.SetMeta(123)
		data, e = json.Marshal(err2)
		r.NoError(e)

		e = json.Unmarshal(data, &obj)
		r.NoError(e)
		r.InEpsilon(123, obj["meta"].(float64), 0.0001)
	}

	{
		e := err.Clone()
		r.Equal(400, e.HTTPCode)
		r.Equal(400, e.StatusCode())
		r.Equal(err.Error(), e.Error())
		r.Equal(errors.New("invalid arguments"), e.Unwrap())
		r.Equal(err, e)
		r.ErrorIs(err, e)
		r.ErrorIs(e, err)
		r.True(e.Is(err))
		r.False(e.Is(errors.New("invalid arguments")))
		r.NotErrorIs(e, errors.New("invalid arguments"))
	}

	{
		e := fmt.Errorf("error: %w", err)
		r.ErrorIs(e, err)
	}

	{
		e, ok := As(err)
		r.True(ok)
		r.Equal(err, e)

		e, ok = As(errors.New("error"))
		r.False(ok)
		r.NotEqual(err, e)
	}
}

// TestMarshalJSON_MetaNil tests MarshalJSON with nil Meta
func TestMarshalJSON_MetaNil(t *testing.T) {
	r := require.New(t)

	err := &Error{
		HTTPCode: 400,
		Err:      errors.New("test error"),
		Code:     "TEST_ERROR",
		Meta:     nil, // Explicitly set to nil
	}

	data, e := json.Marshal(err)
	r.NoError(e)

	var obj map[string]any
	e = json.Unmarshal(data, &obj)
	r.NoError(e)
	r.Equal("TEST_ERROR", obj["code"])
	r.Equal("(400): test error", obj["error"])
	r.Equal("test error", obj["meta"])
}

// TestMarshalJSON_CodeEmpty tests MarshalJSON with empty Code
func TestMarshalJSON_CodeEmpty(t *testing.T) {
	r := require.New(t)

	err := &Error{
		HTTPCode: 404,
		Err:      errors.New("not found"),
		Code:     "", // Empty code
	}

	data, e := json.Marshal(err)
	r.NoError(e)

	var obj map[string]any
	e = json.Unmarshal(data, &obj)
	r.NoError(e)
	r.Equal("404", obj["code"]) // Should use HTTPCode as code
}

// TestMarshalJSON_ErrorEmpty tests MarshalJSON with empty error message
func TestMarshalJSON_ErrorEmpty(t *testing.T) {
	r := require.New(t)

	err := &Error{
		HTTPCode: 500,
		Err:      nil,
		Code:     "INTERNAL_ERROR",
	}

	data, e := json.Marshal(err)
	r.NoError(e)

	var obj map[string]any
	e = json.Unmarshal(data, &obj)
	r.NoError(e)
	r.Equal("INTERNAL_ERROR", obj["code"])
	// error field should not exist or be empty
	_, hasError := obj["error"]
	r.False(hasError)
}

// TestMarshalJSON_MetaCodeInStruct tests MarshalJSON with code already in Meta struct
func TestMarshalJSON_MetaCodeInStruct(t *testing.T) {
	r := require.New(t)

	type MetaWithCode struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	err := &Error{
		HTTPCode: 400,
		Err:      errors.New("test error"),
		Code:     "DEFAULT_CODE",
		Meta: MetaWithCode{
			Code:    "META_CODE",
			Message: "test message",
		},
	}

	data, e := json.Marshal(err)
	r.NoError(e)

	var obj map[string]any
	e = json.Unmarshal(data, &obj)
	r.NoError(e)
	// Should use code from Meta struct, not from Error.Code
	r.Equal("META_CODE", obj["code"])
	r.Equal("test message", obj["message"])
}

// TestMarshalJSON_MetaErrorInStruct tests MarshalJSON with error already in Meta struct
func TestMarshalJSON_MetaErrorInStruct(t *testing.T) {
	r := require.New(t)

	type MetaWithError struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}

	err := &Error{
		HTTPCode: 400,
		Err:      errors.New("original error"),
		Code:     "TEST_CODE",
		Meta: MetaWithError{
			Error:   "meta error message",
			Details: "details",
		},
	}

	data, e := json.Marshal(err)
	r.NoError(e)

	var obj map[string]any
	e = json.Unmarshal(data, &obj)
	r.NoError(e)
	// Should use error from Meta struct
	r.Equal("meta error message", obj["error"])
	r.Equal("details", obj["details"])
}

// TestMarshalJSON_MetaWithMap tests MarshalJSON with map as Meta
func TestMarshalJSON_MetaWithMap(t *testing.T) {
	r := require.New(t)

	err := &Error{
		HTTPCode: 400,
		Err:      errors.New("test error"),
		Code:     "TEST_CODE",
		Meta: map[string]any{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		},
	}

	data, e := json.Marshal(err)
	r.NoError(e)

	var obj map[string]any
	e = json.Unmarshal(data, &obj)
	r.NoError(e)
	r.Equal("value1", obj["key1"])
	r.InEpsilon(123, obj["key2"].(float64), 0.001)
	r.True(obj["key3"].(bool))
	r.Equal("TEST_CODE", obj["code"])
}

// TestMarshalJSON_MetaWithPrimitiveType tests MarshalJSON with primitive type as Meta
func TestMarshalJSON_MetaWithPrimitiveType(t *testing.T) {
	r := require.New(t)

	// Test with string
	err1 := &Error{
		HTTPCode: 400,
		Err:      errors.New("test error"),
		Meta:     "simple string meta",
	}

	data, e := json.Marshal(err1)
	r.NoError(e)

	var obj map[string]any
	e = json.Unmarshal(data, &obj)
	r.NoError(e)
	r.Equal("simple string meta", obj["meta"])

	// Test with number
	err2 := &Error{
		HTTPCode: 400,
		Err:      errors.New("test error"),
		Meta:     42,
	}

	data, e = json.Marshal(err2)
	r.NoError(e)

	e = json.Unmarshal(data, &obj)
	r.NoError(e)
	r.InEpsilon(42, obj["meta"].(float64), 0.001)
}
