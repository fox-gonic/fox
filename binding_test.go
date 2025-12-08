package fox

import (
	"bytes"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fox-gonic/fox/httperrors"
)

type AuthInfo struct {
	Username string
}

type Service struct {
	Name string
}

type MixStruct struct {
	Page       int        `query:"page"`
	PageSize   int        `query:"page_size"`
	IDs        []int      `query:"ids[]"`
	Start      *time.Time `query:"start"         time_format:"unix"`
	Referer    string     `header:"referer"`
	XRequestID string     `header:"X-Request-Id"`
	Vary       []string   `header:"vary"`
	Name       string     `json:"name"`
	Content    *string    `json:"content"`

	AuthInfo *AuthInfo `context:"auth_info"`
	service  *Service  `context:"service"`
	UserID   int64     `context:"user_id"`
}

func TestBinding(t *testing.T) {
	var (
		obj        MixStruct
		url        = "/?page=1&page_size=30&ids[]=1&ids[]=2&ids[]=3&ids[]=4&ids[]=5&start=1669732749"
		referer    = "http://domain.name/posts"
		varyHeader = []string{"X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame", "Accept-Encoding, Accept, X-Requested-With"}
		XRequestID = "l4dCIsjENo3QsCoX"
	)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Referer", referer)
	req.Header.Set("X-Request-Id", XRequestID)
	req.Header.Add("vary", varyHeader[0])
	req.Header.Add("vary", varyHeader[1])

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
		Request: req,
	}

	ctx.Set("auth_info", &AuthInfo{Username: "binder"})
	ctx.Set("service", &Service{Name: "service"})
	ctx.Set("user_id", int64(123))

	err := bind(ctx, obj)
	require.Equal(t, ErrBindNonPointerValue, err)

	err = bind(ctx, &obj)
	require.NoError(t, err)
	assert.Equal(t, 1, obj.Page)
	assert.Equal(t, 30, obj.PageSize)
	assert.Equal(t, referer, obj.Referer)
	assert.Equal(t, XRequestID, obj.XRequestID)
	assert.Equal(t, varyHeader, obj.Vary)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, obj.IDs)
	assert.NotZero(t, obj.Start)
	assert.Equal(t, &AuthInfo{Username: "binder"}, obj.AuthInfo)
	assert.Nil(t, obj.service)
	assert.Equal(t, int64(123), obj.UserID)
}

func TestBindingJSON(t *testing.T) {
	var (
		obj        MixStruct
		url        = "/?page=1&page_size=30&ids[]=1&ids[]=2&ids[]=3&ids[]=4&ids[]=5&start=1669732749"
		referer    = "http://domain.name/posts"
		varyHeader = []string{"X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame", "Accept-Encoding, Accept, X-Requested-With"}
		XRequestID = "l4dCIsjENo3QsCoX"
	)

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"name": "Binder"}`))
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Referer", referer)
	req.Header.Set("X-Request-Id", XRequestID)
	req.Header.Add("vary", varyHeader[0])
	req.Header.Add("vary", varyHeader[1])

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
		Request: req,
	}

	err := bind(ctx, &obj)
	require.NoError(t, err)
	assert.Equal(t, 1, obj.Page)
	assert.Equal(t, 30, obj.PageSize)
	assert.Equal(t, referer, obj.Referer)
	assert.Equal(t, XRequestID, obj.XRequestID)
	assert.Equal(t, varyHeader, obj.Vary)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, obj.IDs)
	assert.Equal(t, "Binder", obj.Name)
	assert.Nil(t, obj.Content)
	assert.NotZero(t, obj.Start)

	req, _ = http.NewRequest(http.MethodPost, url, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	ctx = &Context{
		Context: &gin.Context{
			Request: req,
		},
		Request: req,
	}

	err = bind(ctx, &obj)
	require.NoError(t, err)
}

var ErrPasswordTooShort = &httperrors.Error{
	HTTPCode: http.StatusBadRequest,
	Err:      errors.New("password too short"),
	Code:     "PASSWORD_TOO_SHORT",
}

type CreateUserArgs struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (args *CreateUserArgs) IsValid() error {
	if args.Username == "" && args.Email == "" {
		return httperrors.ErrInvalidArguments
	}
	if len(args.Password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}

func TestIsValider(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBufferString(`{"name": "Binder"}`))
	req.Header.Set("Content-Type", "application/json")

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
		Request: req,
	}

	err := bind(ctx, &CreateUserArgs{})
	require.Error(t, err)
	require.Equal(t, httperrors.ErrInvalidArguments, err)

	err = bind(ctx, &CreateUserArgs{
		Username: "binder",
	})
	require.Error(t, err)
	require.Equal(t, ErrPasswordTooShort, err)

	err = bind(ctx, &CreateUserArgs{
		Username: "binder",
		Password: "123456",
	})
	require.NoError(t, err)
}

// TestFilterFlags tests the filterFlags function
func TestFilterFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no flags",
			input:    "application/json",
			expected: "application/json",
		},
		{
			name:     "with space",
			input:    "application/json charset=utf-8",
			expected: "application/json",
		},
		{
			name:     "with semicolon",
			input:    "application/json; charset=utf-8",
			expected: "application/json",
		},
		{
			name:     "space at start",
			input:    " application/json",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterFlags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestQueryBinding_Name tests queryBinding Name method
func TestQueryBinding_Name(t *testing.T) {
	qb := queryBinding{}
	assert.Equal(t, "query", qb.Name())
}

// TestQueryBinding_Bind tests queryBinding Bind method
func TestQueryBinding_Bind(t *testing.T) {
	type QueryArgs struct {
		Page     int    `query:"page"`
		PageSize int    `query:"page_size"`
		Keyword  string `query:"keyword"`
	}

	t.Run("successful binding", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/?page=1&page_size=20&keyword=test", nil)

		var args QueryArgs
		qb := queryBinding{}
		err := qb.Bind(req, &args)

		require.NoError(t, err)
		assert.Equal(t, 1, args.Page)
		assert.Equal(t, 20, args.PageSize)
		assert.Equal(t, "test", args.Keyword)
	})

	t.Run("binding without keyword", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/?page=1&page_size=20", nil)

		var args QueryArgs
		qb := queryBinding{}
		err := qb.Bind(req, &args)

		require.NoError(t, err)
		assert.Equal(t, 1, args.Page)
		assert.Equal(t, 20, args.PageSize)
		assert.Empty(t, args.Keyword)
	})
}
