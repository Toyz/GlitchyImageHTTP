package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	glitch "github.com/sugoiuguu/go-glitch"
	"github.com/unrolled/render"
)

var allowedFileTypes = []string{"image/jpeg", "image/png"}
var htmlRender *render.Render
var saveMode filemodes.SaveMode

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		htmlRender.HTML(w, http.StatusOK, "index", token)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("uploadfile")
		defer file.Close()

		expression := r.FormValue("expression")
		if err != nil {
			log.Println(err)
			return
		}

		// Hack: This is hacky as all hell just to get the damn fileHeader form the bytes
		cntType := core.GetMimeType(file)
		if ok, _ := core.InArray(cntType, allowedFileTypes); !ok {
			return
		}

		img, _, err := image.Decode(file)
		if err != nil {
			log.Println(err)
			return
		}

		buff := new(bytes.Buffer)

		expr, err := glitch.CompileExpression(expression)
		if err != nil {
			log.Println(err)
			return
		}

		out, err := expr.JumblePixels(img)
		if err != nil {
			log.Println(err)
			return
		}

		png.Encode(buff, out)

		md5Sum := core.GetMD5(buff.Bytes())
		idx := filemodes.GetID(md5Sum)
		fileName := fmt.Sprintf("%x.png", md5Sum)

		actualFileName, folder := saveMode.Write(buff.Bytes(), fileName)

		session, c := database.MongoInstance.GetCollection()
		defer session.Close()

		err = c.Insert(&database.ArtItem{
			ID:         idx,
			FileName:   fileName,
			Folder:     folder,
			FullPath:   actualFileName,
			Expression: expression,
		})

		if err != nil {
			log.Println(err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/%s", idx), 302)
	}
}

func viewImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["imageid"]

	session, c := database.MongoInstance.GetCollection()
	defer session.Close()

	var image database.ArtItem
	c.Find(bson.M{"id": id}).One(&image)

	if len(image.FullPath) <= 0 {
		http.Redirect(w, r, "/", 302)
		return
	}

	htmlRender.HTML(w, http.StatusOK, "img", image)
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

func main() {
	database.NewMongo()

	staticFilePath := core.GetEnv("HTTP_UPLOADS_URL", "/img/")

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

	htmlRender = render.New(htmlOptions)

	saveMode = filemodes.GetFileMode()
	saveMode.Setup()

	r := mux.NewRouter()

	if len(staticFilePath) > 0 && strings.EqualFold(core.GetSaveMode(), "fs") {
		fs := filemodes.FSMode{}
		r.PathPrefix(staticFilePath).Handler(http.StripPrefix(staticFilePath, http.FileServer(http.Dir(fs.Path()))))
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(core.GetPublicFolder()))))

	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/{imageid}", viewImage).Methods("GET")
	r.HandleFunc("/upload", upload).Methods("POST")

	log.Fatal(http.ListenAndServe(core.GetEnv("LISTEN", ":8080"), r))
}
