package fox

import (
	"net/http"

	"github.com/miclle/fox/render"
)

// Render interface is to be implemented by JSON, XML, HTML, YAML and so on.
type Render interface {
	// Render writes data with custom ContentType.
	Render(http.ResponseWriter) error
	// WriteContentType writes custom ContentType.
	WriteContentType(w http.ResponseWriter)
}

var (
	_ Render = render.JSON{}
	_ Render = render.IndentedJSON{}
	_ Render = render.JsonpJSON{}
	_ Render = render.XML{}
	_ Render = render.String{}
	_ Render = render.Redirect{}
	_ Render = render.Data{}
	_ Render = render.HTML{}
	_ Render = render.YAML{}
	_ Render = render.Reader{}
	_ Render = render.ASCIIJSON{}
	_ Render = render.ProtoBuf{}
)
