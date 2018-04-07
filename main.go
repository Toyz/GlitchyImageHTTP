package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	glitch "github.com/sugoiuguu/go-glitch"
)

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("./tmpls/index.html")
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
		defer file.Close()
		img, _, _ := image.Decode(file)

		buff := new(bytes.Buffer)

		expr, _ := glitch.CompileExpression(expression)
		out, _ := expr.JumblePixels(img)
		png.Encode(buff, out)

		imgBase64Str := base64.StdEncoding.EncodeToString(buff.Bytes())

		t, _ := template.ParseFiles("./tmpls/img.html")
		t.Execute(w, imgBase64Str)
	}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", index)
	r.HandleFunc("/upload", upload)

	log.Fatal(http.ListenAndServe(getEnv("LISTEN", ":8080"), r))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
