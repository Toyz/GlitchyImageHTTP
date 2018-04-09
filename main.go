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
	"text/template"
	"time"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
	"github.com/gorilla/mux"
	glitch "github.com/sugoiuguu/go-glitch"
)

var allowedFileTypes = []string{"image/jpeg", "image/png"}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles(core.GetTemplateFilePath("index"))
		t.Execute(w, token)
	}
}

// Exampel expression:
// 	128 & (c + 255) : (s ^ (c ^ 255)) + 25
func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("uploadfile")
		expression := r.FormValue("expression")
		if err != nil {
			fmt.Println(err)
			return
		}

		// Hack: This is hacky as all hell just to get the damn fileHeader form the bytes
		fileHeader := make([]byte, 512)
		if _, err := file.Read(fileHeader); err != nil {
			return
		}
		if _, err := file.Seek(0, 0); err != nil {
			return
		}
		cntType := http.DetectContentType(fileHeader)
		if ok, _ := core.InArray(cntType, allowedFileTypes); !ok {
			return
		}

		defer file.Close()
		img, _, _ := image.Decode(file)

		buff := new(bytes.Buffer)

		expr, _ := glitch.CompileExpression(expression)
		out, _ := expr.JumblePixels(img)
		png.Encode(buff, out)

		sum := md5.Sum(buff.Bytes())

		var saveMode filemodes.SaveMode
		switch strings.ToLower(core.GetEnv("SAVE_MODE", "disk")) {
		case "disk":
			saveMode = filemodes.FSMode{}
			break
		}
		actualFileName := saveMode.Write(buff.Bytes(), fmt.Sprintf("%x.png", sum))

		t, _ := template.ParseFiles(core.GetTemplateFilePath("img"))
		t.Execute(w, fmt.Sprintf("%s", actualFileName))
	}
}

func main() {
	staticFilePath := core.GetEnv("HTTP_UPLOADS_URL", "/img/")

	r := mux.NewRouter()

	r.PathPrefix(staticFilePath).Handler(http.StripPrefix(staticFilePath, http.FileServer(http.Dir(core.UploadsFolder()))))
	r.HandleFunc("/", index)
	r.HandleFunc("/upload", upload)

	log.Fatal(http.ListenAndServe(core.GetEnv("LISTEN", ":8080"), r))
}
