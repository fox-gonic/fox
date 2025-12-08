package fox

import (
	"bytes"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

	t.Run("binding with nil validator", func(t *testing.T) {
		// Save original validator
		originalValidator := binding.Validator
		defer func() {
			binding.Validator = originalValidator
		}()

		// Set validator to nil
		binding.Validator = nil

		req, _ := http.NewRequest(http.MethodGet, "/?page=1&page_size=20", nil)

		var args QueryArgs
		qb := queryBinding{}
		err := qb.Bind(req, &args)

		require.NoError(t, err)
		assert.Equal(t, 1, args.Page)
		assert.Equal(t, 20, args.PageSize)
	})

	t.Run("binding with MapFormWithTag error", func(t *testing.T) {
		type InvalidArgs struct {
			Number int `query:"number"`
		}

		req, _ := http.NewRequest(http.MethodGet, "/?number=not-a-number", nil)

		var args InvalidArgs
		qb := queryBinding{}
		err := qb.Bind(req, &args)

		// MapFormWithTag will attempt to parse "not-a-number" as int, which should fail
		require.Error(t, err)
	})
}

// TestBind_DefaultBinder tests bind function with DefaultBinder
func TestBind_DefaultBinder(t *testing.T) {
	// Save original binders
	originalBinders := binders
	originalBodyBinders := bodyBinders
	originalDefaultBinder := DefaultBinder
	defer func() {
		binders = originalBinders
		bodyBinders = originalBodyBinders
		DefaultBinder = originalDefaultBinder
	}()

	t.Run("DefaultBinder as BindingBody", func(t *testing.T) {
		// Clear binders to force DefaultBinder usage
		binders = make(map[string]binding.Binding)
		bodyBinders = make(map[string]binding.BindingBody)
		DefaultBinder = binding.JSON

		type TestBody struct {
			Name string `json:"name"`
		}

		req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}`))
		req.Header.Set("Content-Type", "application/custom")

		ctx := &Context{
			Context: &gin.Context{
				Request: req,
			},
			Request: req,
		}

		var obj TestBody
		err := bind(ctx, &obj)
		require.NoError(t, err)
		assert.Equal(t, "test", obj.Name)
	})

	t.Run("DefaultBinder with empty body", func(t *testing.T) {
		// Clear binders to force DefaultBinder usage
		binders = make(map[string]binding.Binding)
		bodyBinders = make(map[string]binding.BindingBody)
		DefaultBinder = binding.JSON

		type TestBody struct {
			Name string `json:"name"`
		}

		req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(""))
		req.Header.Set("Content-Type", "application/custom")

		ctx := &Context{
			Context: &gin.Context{
				Request: req,
			},
			Request: req,
		}

		var obj TestBody
		err := bind(ctx, &obj)
		require.NoError(t, err)
		assert.Empty(t, obj.Name)
	})

	t.Run("DefaultBinder as regular Binding", func(t *testing.T) {
		// Clear binders to force DefaultBinder usage
		binders = make(map[string]binding.Binding)
		bodyBinders = make(map[string]binding.BindingBody)
		DefaultBinder = binding.Form

		type TestBody struct {
			Name string `form:"name"`
		}

		req, _ := http.NewRequest(http.MethodPost, "/?name=test", bytes.NewBufferString("name=formtest"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := &Context{
			Context: &gin.Context{
				Request: req,
			},
			Request: req,
		}

		var obj TestBody
		err := bind(ctx, &obj)
		require.NoError(t, err)
		assert.Equal(t, "formtest", obj.Name)
	})
}

// TestBind_BodyBinder tests bind function with bodyBinders
func TestBind_BodyBinder(t *testing.T) {
	// Save original binders
	originalBinders := binders
	originalBodyBinders := bodyBinders
	defer func() {
		binders = originalBinders
		bodyBinders = originalBodyBinders
	}()

	t.Run("BodyBinder with empty body", func(t *testing.T) {
		type TestBody struct {
			Name string `json:"name"`
		}

		req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(""))
		req.Header.Set("Content-Type", "application/json")

		ctx := &Context{
			Context: &gin.Context{
				Request: req,
			},
			Request: req,
		}

		var obj TestBody
		err := bind(ctx, &obj)
		require.NoError(t, err)
		assert.Empty(t, obj.Name)
	})
}

// TestBind_PointerToPointer tests bind with pointer to pointer
func TestBind_PointerToPointer(t *testing.T) {
	type Inner struct {
		Name string `json:"name"`
	}

	type TestBody struct {
		Data **Inner `json:"data"`
	}

	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"data":{"name":"test"}}`))
	req.Header.Set("Content-Type", "application/json")

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
		Request: req,
	}

	var obj TestBody
	err := bind(ctx, &obj)
	require.NoError(t, err)
	require.NotNil(t, obj.Data)
	require.NotNil(t, *obj.Data)
	assert.Equal(t, "test", (*obj.Data).Name)
}

// TestBind_NonStructTarget tests bind with non-struct target
func TestBind_NonStructTarget(t *testing.T) {
	type StringAlias string

	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`"test"`))
	req.Header.Set("Content-Type", "application/json")

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
		Request: req,
	}

	var obj StringAlias
	err := bind(ctx, &obj)
	require.NoError(t, err)
	assert.Equal(t, StringAlias("test"), obj)
}

// TestBind_RequestBodyError tests bind with request body read error
func TestBind_RequestBodyError(t *testing.T) {
	type TestBody struct {
		Name string `json:"name"`
	}

	// Create a request with a body that will cause an error when reading
	req, _ := http.NewRequest(http.MethodPost, "/", errReader(0))
	req.Header.Set("Content-Type", "application/json")

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
		Request: req,
	}

	var obj TestBody
	err := bind(ctx, &obj)
	require.Error(t, err)
}

// errReader is a reader that always returns an error
type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

// TestBind_ContextFieldConversion tests context field type conversion
func TestBind_ContextFieldConversion(t *testing.T) {
	type TestStruct struct {
		IntValue   int64 `context:"int_value"`
		FloatValue int   `context:"float_value"`
	}

	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
		Request: req,
	}

	// Set context values with compatible types
	ctx.Set("int_value", int64(123))
	ctx.Set("float_value", int(456))

	var obj TestStruct
	err := bind(ctx, &obj)
	require.NoError(t, err)
	assert.Equal(t, int64(123), obj.IntValue)
	assert.Equal(t, 456, obj.FloatValue)
}
