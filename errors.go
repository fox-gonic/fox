package fox

import (
	"net/http"
	"reflect"

	"github.com/fox-gonic/fox/render"
)

var errorType = reflect.ValueOf(Error{}).Type()

// Error represents a error's specification.
type Error struct {
	Err         error
	Status      int
	Headers     map[string]string
	ContentType string
	Message     any
}

var _ error = &Error{}

// Error implements the error interface.
func (msg *Error) Error() string {

	if msg.Err == nil {
		return http.StatusText(msg.Status)
	}

	return msg.Err.Error()
}

// Unwrap method
func (msg *Error) Unwrap() error { return msg.Err }

// Format creates a properly formatted message
func (msg *Error) Format() any {

	data := map[string]any{}

	if msg.Message != nil {
		value := reflect.ValueOf(msg.Message)
		switch value.Kind() {
		case reflect.Struct:
			return msg.Message
		case reflect.Map:
			for _, key := range value.MapKeys() {
				data[key.String()] = value.MapIndex(key).Interface()
			}
		default:
			data["message"] = msg.Message
		}
	}

	if _, exists := data["error"]; !exists {
		data["error"] = msg.Error()
	}

	return data
}

// MarshalJSON implements the json.Marshaller interface.
func (msg *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(msg.Format())
}

// Render (Error) writes data with request accept ContentType.
func (msg Error) Render(w http.ResponseWriter, accepts ...string) error {

	if msg.Status >= 100 && msg.Status <= 999 {
		w.WriteHeader(msg.Status)
	}

	header := w.Header()
	for k, v := range msg.Headers {
		header.Set(k, v)
	}

	if msg.ContentType == "" {
		for _, v := range accepts {
			if len(v) > 0 {
				msg.ContentType = v
				break
			}
		}
	}

	var r Render
	switch msg.ContentType {
	case MIMEJSON:
		r = render.JSON{Data: msg.Format()}

	case MIMEXML, MIMEXML2:
		r = render.XML{Data: msg.Format()}

	case MIMEPROTOBUF:
		r = render.ProtoBuf{Data: msg.Format()}

	case MIMEYAML:
		r = render.YAML{Data: msg.Format()}

	default: // MIMEPlain
		r = render.String{Format: msg.Error()}
	}

	return r.Render(w)
}
