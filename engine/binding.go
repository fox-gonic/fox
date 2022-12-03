package engine

import (
	"errors"
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
	binding.MIMEJSON:              binding.JSON,          // json
	binding.MIMEYAML:              binding.YAML,          // yaml
	binding.MIMEXML:               binding.XML,           // xml
	binding.MIMEXML2:              binding.XML,           // xml
	binding.MIMEMultipartPOSTForm: binding.FormMultipart, // form
	binding.MIMEPOSTForm:          binding.Form,          // form
	binding.MIMEPROTOBUF:          binding.ProtoBuf,      // protobuf
	binding.MIMETOML:              binding.TOML,          // toml
}

// Bind request arguments
func Bind(req *http.Request, obj interface{}, params ...map[string][]string) error {
	var binder = &binder{}
	return binder.Bind(req, obj, params...)
}

type binder struct{}

func (binder *binder) Bind(req *http.Request, obj interface{}, params ...map[string][]string) (err error) {

	vPtr := reflect.ValueOf(obj)

	if vPtr.Kind() != reflect.Ptr {
		return ErrBindNonPointerValue
	}

	// bind request body
	// --------------------------------------------------------------------------
	var contentType = filterFlags(req.Header.Get("Content-Type"))

	if binder, exists := binders[contentType]; exists {
		err = binder.Bind(req, obj)
	} else {
		if req.Method == http.MethodGet {
			err = binding.Form.Bind(req, obj)
		} else if DefaultBinder != nil {
			err = DefaultBinder.Bind(req, obj)
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

	if hasQueryField {
		err = Query.Bind(req, obj)
		if err != nil {
			return err
		}
	}

	if hasURIField && len(params) > 0 {
		err = binding.Uri.BindUri(params[0], obj)
		if err != nil {
			return err
		}
	}

	if hasHeaderField {
		err = binding.Header.Bind(req, obj)
		if err != nil {
			return err
		}
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
