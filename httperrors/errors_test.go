package httperrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert := assert.New(t)

	err := New(400, "invalid arguments")
	assert.Equal(400, err.HTTPCode)
	assert.Equal(400, err.StatusCode())

	assert.Equal("(400): invalid arguments", err.Error())
	assert.Equal(errors.New("invalid arguments"), err.Unwrap())

	{
		data, e := json.Marshal(err)
		assert.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		assert.NoError(e)
		assert.Equal("(400): invalid arguments", obj["error"])
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

		assert.True(err.HasField("x-request-id"))
		assert.False(err.HasField("x-access-key"))

		data, e := json.Marshal(err)
		assert.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		assert.NoError(e)
		assert.Equal("invalid arguments", obj["error"])
		assert.Equal("F4CD:20C1B9:2894CD0:3468624:6692A040", obj["reqid"])
		assert.Equal([]any{
			"title field is required",
			"content field is required",
		}, obj["details"])
		assert.Equal("400", fmt.Sprintf("%v", obj["code"]))
		assert.Equal("invalid_arguments", obj["error_code"])
		assert.Empty(obj["omit_empty_field"])
		assert.Empty(obj["ignore_field"])
		assert.Equal("hvnmjnCVyvQ3aOIX", obj["x-request-id"])
		assert.Equal("1.799868", fmt.Sprintf("%v", obj["latency"]))
		assert.Equal("HANDLER", obj["ftype"])
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
		assert.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		assert.NoError(e)
		assert.Equal("invalid arguments", obj["error"])
		assert.Equal("F4CD:20C1B9:2894CD0:3468624:6692A040", obj["reqid"])
		assert.Equal([]any{
			"title field is required",
			"content field is required",
		}, obj["details"])
		assert.Equal("INVALID_ARGUMENTS", fmt.Sprintf("%v", obj["code"]))
	}

	{
		e := New(400, "")
		assert.Equal(400, e.HTTPCode)
		assert.Equal(400, e.StatusCode())
		assert.Equal("(400): Bad Request", e.Error())
	}

	{
		err2 := &Error{}
		assert.Equal("", err2.Error())
		_ = err2.SetHTTPCode(500)
		_ = err2.SetCode("internal_error")

		assert.Equal(500, err2.HTTPCode)
		assert.Equal("internal_error", err2.Code)

		assert.Nil(err2.Fields)
		assert.False(err2.HasField("foo"))

		_ = err2.AddFields(map[string]any{"foo": "bar"})
		assert.NotNil(err2.Fields)
		assert.Equal(map[string]any{"foo": "bar"}, err2.Fields)

		_ = err2.SetMeta(E("database internal error"))
		data, e := json.Marshal(err2)
		assert.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		assert.NoError(e)
		assert.Equal("database internal error", obj["meta"])

		_ = err2.SetMeta(123)
		data, e = json.Marshal(err2)
		assert.NoError(e)

		e = json.Unmarshal(data, &obj)
		assert.NoError(e)
		assert.Equal(float64(123), obj["meta"])
	}

	{
		e := err.Clone()
		assert.Equal(400, e.HTTPCode)
		assert.Equal(400, e.StatusCode())
		assert.Equal(err.Error(), e.Error())
		assert.Equal(errors.New("invalid arguments"), e.Unwrap())
		assert.Equal(err, e)
		assert.True(errors.Is(err, e))
		assert.True(e.Is(err))
		assert.False(e.Is(errors.New("invalid arguments")))
	}

	{
		e := fmt.Errorf("error: %w", err)
		assert.True(errors.Is(e, err))
	}

	{
		e, ok := As(err)
		assert.True(ok)
		assert.Equal(err, e)

		e, ok = As(errors.New("error"))
		assert.False(ok)
		assert.NotEqual(err, e)
	}
}
