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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
	"github.com/Toyz/GlitchyImageHTTP/core/tmplengine"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	glitch "github.com/sugoiuguu/go-glitch"
)

var allowedFileTypes = []string{"image/jpeg", "image/png"}
var saveMode filemodes.SaveMode

func index(w http.ResponseWriter, r *http.Request) {
	crutime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(crutime, 10))
	token := fmt.Sprintf("%x", h.Sum(nil))

	tmplengine.CntRender.HTML(w, http.StatusOK, "index", map[string]string{
		"Error": r.URL.Query().Get("error"),
		"Token": token,
	})
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("uploadfile")
		defer file.Close()

		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())), 302)
			return
		}

		expression := r.FormValue("expression")
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())), 302)
			return
		}

		// Hack: This is hacky as all hell just to get the damn fileHeader form the bytes
		cntType := core.GetMimeType(file)
		if ok, _ := core.InArray(cntType, allowedFileTypes); !ok {
			http.Redirect(w, r, fmt.Sprintf("/?error=%s", url.QueryEscape("File type is not allowed only PNG and JPEG allowed")), 302)
			return
		}

		img, _, err := image.Decode(file)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())), 302)
			return
		}

		buff := new(bytes.Buffer)

		expr, err := glitch.CompileExpression(expression)
		if err != nil {
			//log.Println(err)
			// TODO: make this actually show on the home screen
			// THIS REDIRECT IS ONLY HERE TEMP UNTIL WE WRITE A BETTER ERROR HANDLER... MAYBE USING A "HTTPERROR" STRUCT THAT IS JSON
			http.Redirect(w, r, fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())), 302)
			return
		}

		out, err := expr.JumblePixels(img)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())), 302)
			return
		}

		png.Encode(buff, out)

		md5Sum := core.GetMD5(buff.Bytes())
		idx := filemodes.GetID(md5Sum)
		fileName := fmt.Sprintf("%s.png", md5Sum)

		actualFileName, folder := saveMode.Write(buff.Bytes(), fileName)

		session, c := database.MongoInstance.GetCollection()
		defer session.Close()

		index := mgo.Index{
			Key:        []string{"id", "filename"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		}
		c.EnsureIndex(index)

		err = c.Insert(&database.ArtItem{
			ID:         idx,
			FileName:   fileName,
			Folder:     folder,
			FullPath:   actualFileName,
			Expression: expression,
			Views:      0,
			Uploaded:   time.Now(),
		})

		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())), 302)
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

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"views": 1}},
		ReturnNew: false,
	}
	_, err := c.Find(bson.M{"id": id}).Apply(change, &image)
	if err != nil {
		log.Println(err)
	}
	tmplengine.CntRender.HTML(w, http.StatusOK, "img", image)
}

func main() {
	database.NewMongo()
	tmplengine.New()

	staticFilePath := core.GetEnv("HTTP_UPLOADS_URL", "/img/")

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
