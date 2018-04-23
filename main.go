package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
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

var allowedFileTypes = []string{"image/jpeg", "image/png", "image/jpg"}
var saveMode filemodes.SaveMode
var defaultExpressions []string

func Index(ctx iris.Context) {
	crutime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(crutime, 10))
	token := fmt.Sprintf("%x", h.Sum(nil))

	ctx.ViewData("Home", routing.HomePage{
		Error:      "", //ctx.URLParam("error"),
		Token:      token,
		Expression: defaultExpressions[rand.Intn(len(defaultExpressions))],
	})
	ctx.View("index.html")
}

func Upload(ctx iris.Context) {
	ctx.SetMaxRequestBodySize(5 << 20) // 5mb because we can
	file, fHeader, err := ctx.FormFile("uploadfile")
	if err != nil {
		ctx.JSON(&routing.UploadResult{
			Error: "Upload cannot be empty",
		})
		return
	}
	defer file.Close()

	exps := ctx.FormValues()
	expressions := make([]string, 0, len(exps))
	for k, v := range exps {
		if strings.EqualFold(k, "expression") {
			for _, item := range v {
				if len(strings.TrimSpace(item)) > 0 {
					expressions = append(expressions, item)
				}
			}
		}
	}

	if len(expressions) > 5 {
		ctx.JSON(&routing.UploadResult{
			Error: "Only 5 expressions are allowed",
		})
		return
	}

	// Hack: This is hacky as all hell just to get the damn fileHeader form the bytes
	cntType := core.GetMimeType(file)
	if ok, _ := core.InArray(cntType, allowedFileTypes); !ok {
		ctx.JSON(&routing.UploadResult{
			Error: "File type is not allowed only PNG and JPEG allowed",
		})
		return
	}

	img, _, err := image.Decode(file)
	if err != nil {
		ctx.JSON(&routing.UploadResult{
			Error: err.Error(),
		})
		return
	}

	if (img.Bounds().Max.X * img.Bounds().Max.Y) > (1920 * 1080) {
		img = nil
		ctx.JSON(&routing.UploadResult{
			Error: "Max image size is 1920x1080 (1080p)",
		})
		return
	}

	buff := new(bytes.Buffer)
	out := img

	for _, expression := range expressions {
		expr, err := glitch.CompileExpression(expression)
		if err != nil {
			ctx.JSON(&routing.UploadResult{
				Error: err.Error(),
			})
			return
		}

		newImage, err := expr.JumblePixels(out)
		if err != nil {
			out = nil
			ctx.JSON(&routing.UploadResult{
				Error: err.Error(),
			})
			return
		}
		out = newImage
		newImage = nil
	}

	switch strings.ToLower(cntType) {
	case "image/png":
		png.Encode(buff, out)
		break
	case "image/jpg", "image/jpeg":
		jpeg.Encode(buff, out, nil)
		break
	}

	bounds := out.Bounds()
	out = nil
	img = nil

	md5Sum := core.GetMD5(buff.Bytes())
	idx := filemodes.GetID(md5Sum)
	fileName := fmt.Sprintf("%s.%s", md5Sum, core.MimeToExtension(cntType))

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

	expression := ""
	if len(expressions) == 1 {
		expression = expressions[0]
	}

	err = c.Insert(&database.ArtItem{
		ID:          idx,
		FileName:    fileName,
		OrgFileName: fHeader.Filename,
		Folder:      folder,
		FullPath:    actualFileName,
		Expression:  expression,
		Expressions: expressions,
		Views:       0,
		Uploaded:    time.Now(),
		FileSize:    binary.Size(buff.Bytes()),
		Width:       bounds.Max.X,
		Height:      bounds.Max.Y,
	})

	if err != nil {
		buff = nil
		//ctx.Redirect(fmt.Sprintf("/?error=%s", url.QueryEscape(err.Error())))
		ctx.JSON(&routing.UploadResult{
			Error: err.Error(),
		})
		return
	}

	buff = nil
	ctx.JSON(&routing.UploadResult{
		ID: idx,
	})
	//ctx.Redirect(fmt.Sprintf("/%s", idx))
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

	//image.FullPath = fmt.Sprintf("%s://%s/%s/%s", "https", ctx.Host(), "img", image.FileName)

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
