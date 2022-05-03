package fox

import (
	"net/http"
	"reflect"

	"github.com/miclle/fox/render"
)

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
func (msg Error) Error() string {

	if msg.Err == nil {
		return http.StatusText(msg.Status)
	}

	return msg.Err.Error()
}

// Unwrap method
func (msg *Error) Unwrap() error { return msg.Err }

// Format creates a properly formatted message
func (msg *Error) Format() any {

	jsonData := map[string]any{}

	if msg.Message != nil {
		value := reflect.ValueOf(msg.Message)
		switch value.Kind() {
		case reflect.Struct:
			return msg.Message
		case reflect.Map:
			for _, key := range value.MapKeys() {
				jsonData[key.String()] = value.MapIndex(key).Interface()
			}
		default:
			jsonData["message"] = msg.Message
		}
	}

	if _, exists := jsonData["error"]; !exists {
		jsonData["error"] = msg.Error()
	}

	return jsonData
}

// MarshalJSON implements the json.Marshaller interface.
func (msg *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(msg.Format())
}

// Render (Error) writes data with request accept ContentType.
func (msg *Error) Render(w http.ResponseWriter) error {

	if msg.Status >= 100 && msg.Status <= 999 {
		w.WriteHeader(msg.Status)
	}

	header := w.Header()
	for k, v := range msg.Headers {
		header.Set(k, v)
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
