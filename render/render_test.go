package render_test

import (
	"testing"

	foxrender "github.com/fox-gonic/fox/render"
	ginrender "github.com/gin-gonic/gin/render"
)

func TestRenderAliases(t *testing.T) {
	var _ foxrender.Render = ginrender.JSON{}

	acceptJSON := func(ginrender.JSON) {}
	acceptRedirect := func(ginrender.Redirect) {}

	acceptJSON(foxrender.JSON{Data: "x"})
	acceptRedirect(foxrender.Redirect{Code: 302, Location: "/"})
}
