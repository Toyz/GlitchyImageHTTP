package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
	"github.com/Toyz/GlitchyImageHTTP/routing"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	glitch "github.com/sugoiuguu/go-glitch"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var allowedFileTypes = []string{"image/jpeg", "image/png"}
var saveMode filemodes.SaveMode
var defaultExpressions []string

func Index(ctx iris.Context) {
	crutime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(crutime, 10))
	token := fmt.Sprintf("%x", h.Sum(nil))

	ctx.ViewData("Home", routing.HomePage{
		Error:      ctx.URLParam("error"),
		Token:      token,
		Expression: defaultExpressions[rand.Intn(len(defaultExpressions))],
	})
	ctx.View("index.html")
}

func Upload(ctx iris.Context) {
	ctx.SetMaxRequestBodySize(15 << 20) // 15mb
	file, _, err := ctx.FormFile("uploadfile")
	defer file.Close()

	if err != nil {
		ctx.Redirect(fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())))
		return
	}

	expression := ctx.FormValue("expression")
	if err != nil {
		ctx.Redirect(fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())))
		return
	}

	// Hack: This is hacky as all hell just to get the damn fileHeader form the bytes
	cntType := core.GetMimeType(file)
	if ok, _ := core.InArray(cntType, allowedFileTypes); !ok {
		ctx.Redirect(fmt.Sprintf("/?error=%s", url.QueryEscape("File type is not allowed only PNG and JPEG allowed")))
		return
	}

	img, _, err := image.Decode(file)
	if err != nil {
		ctx.Redirect(fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())))
		return
	}

	buff := new(bytes.Buffer)

	expr, err := glitch.CompileExpression(expression)
	if err != nil {
		// TODO: make this actually show on the home screen
		// THIS REDIRECT IS ONLY HERE TEMP UNTIL WE WRITE A BETTER ERROR HANDLER... MAYBE USING A "HTTPERROR" STRUCT THAT IS JSON
		ctx.Redirect(fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())))
		return
	}

	out, err := expr.JumblePixels(img)
	if err != nil {
		ctx.Redirect(fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())))
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
		FileSize:   binary.Size(buff.Bytes()),
		Width:      out.Bounds().Max.X,
		Height:     out.Bounds().Max.Y,
	})

	if err != nil {
		ctx.Redirect(fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())))
		return
	}

	ctx.Redirect(fmt.Sprintf("/%s", idx))
}

func ViewImage(ctx iris.Context) {
	id := ctx.Params().Get("image")

	session, c := database.MongoInstance.GetCollection()
	defer session.Close()

	var image database.ArtItem
	c.Find(bson.M{"id": id}).One(&image)

	if len(image.FullPath) <= 0 {
		ctx.Redirect("/")
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

	data := ctx.GetViewData()["Header"].(core.HeaderMetaData)
	header := core.Render.Header(fmt.Sprintf("%s - View Image %s", data.Title, image.ID), image.FullPath, data.Desc, image.ID)
	header.ImageHeight = image.Height
	header.ImageWidth = image.Width
	ctx.ViewData("Header", header)
	ctx.ViewData("Data", image)
	ctx.ViewData("BodyClass", "image")

	ctx.View("img.html")
}

func main() {
	rand.Seed(time.Now().Unix())

	database.NewMongo()
	core.Render.New()
	core.AssetManager.New()

	saveMode = filemodes.GetFileMode()
	defaultExpressions, _ = core.AssetManager.ReadFileLines("./assets/glitches.txt")

	//iris.WithPostMaxMemory((10 * datasize.MB).Bytes())
	app := iris.New()

	app.Logger().SetLevel("debug")
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(func(ctx iris.Context) {
		ctx.ViewData("Header", core.Render.Header(
			"Go Glitch",
			"",
			"A powerful website to glitch art using custom expressions",
			"",
		))

		ctx.ViewData("BodyClass", "index")

		ctx.Next()
	})

	tmpEngine := core.Render.Defaults()
	app.RegisterView(tmpEngine.ViewEngine)

	app.Get("/", Index)
	app.Post("/upload", Upload)
	app.Get("/{image:string}", ViewImage)
	app.StaticWeb("/static", "./assets/public")
	app.Build()

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	mode := fasthttpadaptor.NewFastHTTPHandler(m.Middleware(app))

	listening := core.GetEnv("LISTEN", ":8080")
	fasthttp.ListenAndServe(listening, mode)
}
