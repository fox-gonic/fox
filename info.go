package fox

import (
	_ "embed" // embed tmpl
	"html/template"
	"log"

	"github.com/fox-gonic/fox/render"
)

//go:embed info.tmpl
var infoTmpl string

// Info for Engine
func (engine *Engine) Info(c *Context) (*render.HTML, error) {

	fns := template.FuncMap{
		"nameOfFunction": nameOfFunction,
	}

	tmpl, err := template.New("info").Funcs(fns).Parse(infoTmpl)
	if err != nil {
		log.Panicf("parse info template error: %v", err)
		return nil, err
	}

	var sessionOptions *SessionOptions
	engine.Configurations.UnmarshalKey("session", &sessionOptions)

	data := map[string]interface{}{
		"engine":         engine,
		"handlers":       engine.Handlers,
		"not_found":      engine.noRoute,
		"no_method":      engine.noMethod,
		"sessionOptions": sessionOptions,
		"routes":         engine.Routes(),
	}

	return &render.HTML{
		Template: tmpl,
		Data:     data,
	}, nil
}
