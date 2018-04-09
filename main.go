package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
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

// Exampel expression:
// 	128 & (c + 255) : (s ^ (c ^ 255)) + 25
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

		img, _, _ := image.Decode(file)

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

		sum := md5.Sum(buff.Bytes())

		actualFileName := saveMode.Write(buff.Bytes(), fmt.Sprintf("%x.png", sum))

		htmlRender.HTML(w, http.StatusOK, "img", fmt.Sprintf("%s", actualFileName))
	}
}

func main() {
	staticFilePath := core.GetEnv("HTTP_UPLOADS_URL", "/img/")

	htmlRender = render.New(render.Options{
		Directory:  core.GetTemplateFolder(),
		Extensions: []string{".html"},
		Layout:     "layout",
	})

	saveMode = filemodes.GetFileMode()
	saveMode.Setup()

	r := mux.NewRouter()

	if len(staticFilePath) > 0 && strings.EqualFold(core.GetSaveMode(), "fs") {
		fs := filemodes.FSMode{}
		r.PathPrefix(staticFilePath).Handler(http.StripPrefix(staticFilePath, http.FileServer(http.Dir(fs.Path()))))
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(core.GetPublicFolder()))))

	r.HandleFunc("/", index)
	r.HandleFunc("/upload", upload)

	log.Fatal(http.ListenAndServe(core.GetEnv("LISTEN", ":8080"), r))
}
