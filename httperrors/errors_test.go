package httperrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	assert.Equal("(400): invalid arguments", err.Error())

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

		_ = err.AddMessage("stack", "test")

		data, e := json.Marshal(err)
		assert.NoError(e)

		var obj map[string]any
		e = json.Unmarshal(data, &obj)
		assert.NoError(e)
		assert.Equal("invalid arguments", obj["error"])
		assert.Equal("F4CD:20C1B9:2894CD0:3468624:6692A040", obj["reqid"])
		assert.Equal([]any{"title field is required", "content field is required"}, obj["details"])
		assert.Equal("400", fmt.Sprintf("%v", obj["code"]))
		assert.Equal("invalid_arguments", obj["error_code"])
		assert.Empty(obj["omit_empty_field"])
		assert.Empty(obj["ignore_field"])
		assert.NotEmpty(obj["message"])
	}
}
