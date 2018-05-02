package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
	"github.com/Toyz/GlitchyImageHTTP/routing"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
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

var allowedFileTypes = []string{"image/jpeg", "image/png", "image/jpg", "image/gif"}
var saveMode filemodes.SaveMode
var defaultExpressions []string

func Index(ctx iris.Context) {
	token := filemodes.GetID("sdlkfjdklfjdskjfhdskajfhs")

	//core.RedisManager.Set(fmt.Sprintf("Upload%s", token), token, 30*time.Minute)
	ctx.ViewData("Home", routing.HomePage{
		Error:      "", //ctx.URLParam("error"),
		Token:      token,
		Expression: defaultExpressions[rand.Intn(len(defaultExpressions))],
	})

	ctx.View("index.html")
}

func ViewImage(ctx iris.Context) {
	sess := core.SessionManager.Session.Start(ctx)

	id := ctx.Params().Get("image")

	err, image := database.MongoInstance.GetUploadInfo(id)
	if err != nil {
		ctx.Redirect("/")
		return
	}

	lastViewed := sess.GetStringDefault("LastViewed", "")

	if len(lastViewed) <= 0 || !strings.EqualFold(lastViewed, id) {
		err := database.MongoInstance.UploadInfoUpdateViews(image)
		if err != nil {
			log.Println(err)
		}

		image.Views = image.Views + 1 // hack... Gets around the offset not being defined...
		sess.Set("LastViewed", id)
	}

	// This allows us to add expression to the database if they don't exist...
	// Hacky but it works
	for e := 0; e < len(image.Expressions); e++ {
		exp := image.Expressions[e]

		expItem := database.MongoInstance.GetExpression(exp)
		if len(expItem.Expression) <= 0 {
			expItem = database.ExpressionItem{
				Expression: exp,
				Usage:      1,
			}
			database.MongoInstance.InsertExpression(expItem)
		}
	}

	data := ctx.GetViewData()["Header"].(core.HeaderMetaData)
	header := core.Render.Header(fmt.Sprintf("Viewing %s", image.MGID.Hex()), filemodes.GetFileMode().FullPath(image.Folder, image.FileName), data.Desc, image.ID)
	header.ImageHeight = image.Height
	header.ImageWidth = image.Width

	ctx.ViewData("Header", header)
	ctx.ViewData("Data", image)
	ctx.ViewData("BodyClass", "image")

	ctx.View("viewing.html")
}

func main() {
	rand.Seed(time.Now().Unix())

	core.NewRedis()
	core.NewSessions()

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
	tmpEngine.AddFunc("image_path", filemodes.GetFileMode().FullPath)
	tmpEngine.AddFunc("get_exp_id", func(item string) string {
		exp := database.MongoInstance.GetExpression(item)
		id := bson.NewObjectId()

		if len(exp.Expression) <= 0 {
			exp = database.ExpressionItem{
				id, item, 1,
			}

			database.MongoInstance.InsertExpression(exp)
		} else {
			id = exp.MGID
		}

		return id.Hex()
	})

	app.RegisterView(tmpEngine.ViewEngine)

	app.StaticWeb("/static", "./assets/public")

	app.Get("/", Index)
	app.Post("/upload", routing.Upload)

	api := app.Party("/stats")
	{
		exp := api.Party("/exps")
		{
			exp.Get("/most.json", func(ctx iris.Context) {
				routing.ViewedExpressions("-", ctx)
			})
			exp.Get("/least.json", func(ctx iris.Context) {
				routing.ViewedExpressions("", ctx)
			})
		}

		imgs := api.Party("/imgs")
		{
			imgs.Get("/most.json", func(ctx iris.Context) {
				routing.ViewedImages("-", ctx)
			})
			imgs.Get("/least.json", func(ctx iris.Context) {
				routing.ViewedImages("", ctx)
			})
		}
	}

	app.Get("/{image:string}", ViewImage)
	app.Get("/{image:string}/info.json", routing.ViewImageInfo)

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
