package render

import "github.com/gin-gonic/gin/render"

// render alias for gin/render.
type (
	Render       = render.Render
	JSON         = render.JSON
	IndentedJSON = render.IndentedJSON
	SecureJSON   = render.SecureJSON
	JsonpJSON    = render.JsonpJSON
	XML          = render.XML
	String       = render.String
	Redirect     = render.Redirect
	Data         = render.Data
	HTML         = render.HTML
	YAML         = render.YAML
	Reader       = render.Reader
	AsciiJSON    = render.AsciiJSON
	ProtoBuf     = render.ProtoBuf
	TOML         = render.TOML
)
