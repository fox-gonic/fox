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
