package tmplengine

import (
	"html/template"
	"reflect"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/unrolled/render"
)

type Render struct {
	*render.Render
	options render.Options
}

var CntRender *Render

func New() {
	funcs := []template.FuncMap{
		template.FuncMap{
			"hasField": hasField,
		},
	}

	htmlOptions := render.Options{
		Directory:  core.GetTemplateFolder(),
		Extensions: []string{".html"},
		Layout:     "layout",
		Funcs:      funcs,
	}

	htmlRender := render.New(htmlOptions)

	CntRender = &Render{htmlRender, htmlOptions}
}

func hasField(v interface{}, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	return rv.FieldByName(name).IsValid()
}
