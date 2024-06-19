package fox

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin/binding"
)

// ErrBindNonPointerValue is required bind pointer
var ErrBindNonPointerValue = errors.New("can not bind to non-pointer value")

// DefaultBinder default binder
var DefaultBinder binding.Binding = binding.JSON

// Query binder
var Query = &queryBinding{}

var binders = map[string]binding.Binding{
	binding.MIMEMultipartPOSTForm: binding.FormMultipart, // form
	binding.MIMEPOSTForm:          binding.Form,          // form
}

var bodyBinders = map[string]binding.BindingBody{
	binding.MIMEJSON:     binding.JSON,     // json
	binding.MIMEYAML:     binding.YAML,     // yaml
	binding.MIMEXML:      binding.XML,      // xml
	binding.MIMEXML2:     binding.XML,      // xml
	binding.MIMEPROTOBUF: binding.ProtoBuf, // protobuf
	binding.MIMETOML:     binding.TOML,     // toml
}

// bind request arguments
func bind(ctx *Context, obj interface{}) (err error) {

	vPtr := reflect.ValueOf(obj)

	if vPtr.Kind() != reflect.Ptr {
		return ErrBindNonPointerValue
	}

	// bind request body
	// --------------------------------------------------------------------------
	var (
		contentType = filterFlags(ctx.Request.Header.Get("Content-Type"))
		body        []byte
	)

	if body, err = ctx.RequestBody(); err != nil {
		return err
	}

	defer func() {
		// copy the request body to the next handler
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}()

	if ctx.Request.Method == http.MethodGet {
		err = binding.Form.Bind(ctx.Request, obj)

	} else if binder, exists := binders[contentType]; exists {
		err = binder.Bind(ctx.Request, obj)

	} else if bodyBinder, exists := bodyBinders[contentType]; exists {
		if len(body) > 0 {
			err = bodyBinder.BindBody(body, obj)
		}

	} else if DefaultBinder != nil {
		if bodyBinder, ok := DefaultBinder.(binding.BindingBody); ok {
			if len(body) > 0 {
				err = bodyBinder.BindBody(body, obj)
			}
		} else {
			err = DefaultBinder.Bind(ctx.Request, obj)
		}
	}
	if err != nil {
		return err
	}

	// bind request query, header and uri
	// --------------------------------------------------------------------------
	vPtr = vPtr.Elem()

	for vPtr.Kind() == reflect.Ptr {
		if vPtr.IsNil() {
			vPtr.Set(reflect.New(vPtr.Type().Elem()))
		}
		vPtr = vPtr.Elem()
	}

	if vPtr.Kind() != reflect.Struct {
		return
	}

	var vType = vPtr.Type()
	var hasQueryField, hasURIField, hasHeaderField bool

	for i := 0; i < vPtr.NumField(); i++ {
		field := vType.Field(i)

		if tag := field.Tag.Get("query"); tag != "" && tag != "-" {
			hasQueryField = true
		}
		if tag := field.Tag.Get("uri"); tag != "" && tag != "-" {
			hasURIField = true
		}
		if tag := field.Tag.Get("header"); tag != "" && tag != "-" {
			hasHeaderField = true
		}
	}

	// bind query params
	if hasQueryField {
		if err = Query.Bind(ctx.Request, obj); err != nil {
			return err
		}
	}

	// bind uri path
	if hasURIField && len(ctx.Params) > 0 {
		m := make(map[string][]string)
		for _, v := range ctx.Params {
			m[v.Key] = []string{v.Value}
		}
		if err = binding.Uri.BindUri(m, obj); err != nil {
			return err
		}
	}

	// bind header fields
	if hasHeaderField {
		if err = binding.Header.Bind(ctx.Request, obj); err != nil {
			return err
		}
	}

	if valider, ok := obj.(IsValider); ok {
		return valider.IsValid()
	}

	return nil
}

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *http.Request, obj any) error {
	values := req.URL.Query()
	if err := binding.MapFormWithTag(obj, values, "query"); err != nil {
		return err
	}
	if binding.Validator == nil {
		return nil
	}
	return binding.Validator.ValidateStruct(obj)
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
