package core

import (
	"html/template"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/kataras/iris"
	"github.com/kataras/iris/view"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

type (
	render struct {
		ViewEngine *view.HTMLEngine
		path       string
		minify     *minify.M
	}

	HeaderMetaData struct {
		Title       string
		Image       string
		Desc        string
		SiteName    string
		URL         string
		ImageWidth  int
		ImageHeight int
	}
)

var Render *render

func create() {
	folder := GetEnv("ASSET_FOLDER_VIEWS", "./assets/tmpls")
	vExt := GetEnv("VIEWS_EXT", ".html")

	localRender := iris.HTML(folder, vExt).Layout("layout.html").Reload(true)
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)

	Render = &render{ViewEngine: localRender, path: folder, minify: m}
}

func (render *render) AddFunc(name string, method interface{}) *render {
	render.ViewEngine.AddFunc(name, method)

	return render
}

func (render *render) AddLayoutFunc(name string, method interface{}) *render {
	render.ViewEngine.AddLayoutFunc(name, method)

	return render
}

func (render *render) Defaults() *render {
	render.AddFunc("log", func(format string, params ...string) string {
		newKeys := make([]interface{}, len(params))
		for i, v := range params {
			newKeys[i] = v
		}

		return ""
	})

	render.AddFunc("replace", func(from string, to string, word string, number int) string {
		return strings.Replace(word, from, to, number)
	})

	render.AddFunc("longerThan", func(a string, b int) bool {
		return len(a) > b
	})

	render.AddFunc("lessThan", func(a string, b int) bool {
		return len(a) < b
	})

	render.AddFunc("equalTo", func(a string, b int) bool {
		return len(a) == b
	})

	render.AddFunc("raw", func(a string) template.HTML {
		tmpl, _ := render.minify.String("text/html", a)

		return template.HTML(tmpl)
	})

	render.AddFunc("include", func(filePath string) template.HTML {
		tmpl := AssetManager.FileContents(path.Join(render.path, filePath))
		tmpl, _ = render.minify.String("text/html", tmpl)

		// return template.HTML(tmpl)
		return template.HTML(tmpl)
	})

	render.AddFunc("formatUTC", func(a int64) string {
		if a <= 0 {
			return "N/A"
		}

		tm := time.Unix(0, a*1000000)
		return tm.UTC().Format(time.RFC1123)
	})

	render.AddFunc("eq", func(a interface{}, b interface{}) bool {
		return a == b
	})

	render.AddFunc("neq", func(a interface{}, b interface{}) bool {
		return a != b
	})

	render.AddFunc("gr", func(a int, b int) bool {
		return a > b
	})

	render.AddFunc("lt", func(a int, b int) bool {
		return a < b
	})

	render.AddFunc("TrimSpace", func(body string) string {
		return strings.TrimSpace(body)
	})

	render.AddFunc("ToUpper", func(item string) string {
		return strings.ToUpper(item)
	})

	render.AddFunc("ToLower", func(item string) string {
		return strings.ToLower(item)
	})

	render.AddFunc("ToTitle", func(item string) string {
		return strings.ToTitle(item)
	})

	render.AddFunc("Join", func(joiner string, args ...string) string {
		return strings.Join(args, joiner)
	})

	render.AddFunc("hasField", func(v interface{}, name string) bool {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return false
		}
		return rv.FieldByName(name).IsValid()
	})

	render.AddFunc("formatFS", func(size int) string {
		return datasize.ByteSize(size).HR()
	})

	return render
}

func (render *render) New() *render {
	create()

	return Render
}

func (render *render) Header(title, logo, desc, ogUrl string) HeaderMetaData {
	return HeaderMetaData{
		title, logo, desc, "Go Glitch", ogUrl, 0, 0,
	}
}
