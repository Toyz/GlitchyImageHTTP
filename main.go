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
var defaultExpressions []routing.API_Expression

func Index(ctx iris.Context) {
	token := bson.NewObjectId().Hex()

	//core.RedisManager.Set(fmt.Sprintf("Upload%s", token), token, 30*time.Minute)
	ctx.ViewData("Home", routing.HomePage{
		Error:      "", //ctx.URLParam("error"),
		Token:      token,
		Expression: defaultExpressions[rand.Intn(len(defaultExpressions))].Expression,
	})

	ctx.View("index.html")
}

func ViewImage(ctx iris.Context) {
	sess := core.SessionManager.Session.Start(ctx)

	id := ctx.Params().Get("image")

	if !bson.IsObjectIdHex(id) {
		ctx.Redirect("/")
		return
	}

	upload := database.MongoInstance.GetUpload(bson.ObjectIdHex(id))
	image := database.MongoInstance.GetImageInfo(upload.ImageID)

	lastViewed := sess.GetStringDefault("LastViewed", "")

	if len(lastViewed) <= 0 || !strings.EqualFold(lastViewed, id) {
		err := database.MongoInstance.IncUploadViews(upload.MGID)
		if err != nil {
			log.Println(err)
		}

		upload.Views = upload.Views + 1
		sess.Set("LastViewed", id)
	}

	expressions := make([]database.ExpressionItem, len(upload.Expressions))
	for e := 0; e < len(upload.Expressions); e++ {
		expressions[e] = database.MongoInstance.GetExpression(upload.Expressions[e])
	}

	data := ctx.GetViewData()["Header"].(core.HeaderMetaData)
	header := core.Render.Header(fmt.Sprintf("Viewing %s", image.MGID.Hex()), filemodes.GetFileMode().FullPath(image.Folder, image.FileName), data.Desc, upload.MGID.Hex())
	header.ImageHeight = image.Height
	header.ImageWidth = image.Width

	user := database.User{}
	if upload.User.Valid() {
		user = database.MongoInstance.GetUserByID(upload.User)
		user.Password = ""
		user.Email = ""
	}

	ctx.ViewData("Header", header)
	ctx.ViewData("Image", image)
	ctx.ViewData("Upload", upload)
	ctx.ViewData("Uploader", user)
	ctx.ViewData("Exps", expressions)
	ctx.ViewData("BodyClass", "image")

	ctx.View("viewing.html")
}

func loadExpressions(expressionList []string) {
	var expressCat *routing.API_Category

	for i := 0; i < len(expressionList); i++ {
		line := strings.TrimSpace(expressionList[i])
		if strings.HasPrefix(line, "//") {
			continue
		}

		if strings.HasPrefix(line, "#") {
			catName := strings.TrimSpace(line[1:])
			cat := database.MongoInstance.GetCategoryByName(catName)
			if len(cat.Name) <= 0 {
				cat = database.MongoInstance.AddCategory(database.CategoryItem{
					Name: catName,
				})
			}

			expressCat = &routing.API_Category{
				Name: cat.Name,
				ID:   cat.MGID.Hex(),
			}
			continue
		}

		exp := database.MongoInstance.GetExpressionByName(line)
		if len(exp.ExpressionCmp) <= 0 {
			catIds := make([]bson.ObjectId, 1)
			catIds[0] = bson.ObjectIdHex(expressCat.ID)

			exp = database.MongoInstance.AddExpression(database.ExpressionItem{
				Expression: line,
				Usage:      1,
				CatIDs:     catIds,
			})
		} else {
			catExist := false
			for i := 0; i < len(exp.CatIDs); i++ {
				catLID := exp.CatIDs[i].Hex()
				catExist = strings.EqualFold(catLID, expressCat.ID)
			}

			if !catExist {
				ccc := bson.ObjectIdHex(expressCat.ID)
				exp.CatIDs = append(exp.CatIDs, ccc)
				// Update
				database.MongoInstance.AddCateoryToExpression(exp, ccc)
			}
		}

		catsAll := make([]routing.API_Category, len(exp.CatIDs))
		for c := 0; c < len(exp.CatIDs); c++ {
			cat := database.MongoInstance.GetCategory(exp.CatIDs[c])
			catsAll[c] = routing.API_Category{
				Name: cat.Name,
				ID:   cat.MGID.Hex(),
			}
		}

		defaultExpressions = append(defaultExpressions, routing.API_Expression{
			Expression: exp.Expression,
			Categories: catsAll,
			Usage:      exp.Usage,
			ID:         exp.MGID.Hex(),
		})
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	core.NewRedis()
	core.NewSessions()

	database.NewMongo()
	core.Render.New()
	core.AssetManager.New()

	saveMode = filemodes.GetFileMode()
	expressionList, _ := core.AssetManager.ReadFileLines("./assets/glitches.txt")
	loadExpressions(expressionList)

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

	app.RegisterView(tmpEngine.ViewEngine)

	app.StaticWeb("/static", "./assets/public")
	filemodes.GetFileMode().StaticPath(app)

	app.Get("/", Index)
	app.Post("/upload", routing.Upload)

	user := app.Party("/u")
	{
		tools := user.Party("/tools")
		{
			tools.Get("/register", routing.UserJoin)
			tools.Get("/signin", routing.UserLogin)
			tools.Get("/logout", func(ctx iris.Context) {
				routing.UserTool(routing.LOGOUT_USER, ctx)
			})
			tools.Post("/register", func(ctx iris.Context) {
				routing.UserTool(routing.CREATE_USER, ctx)
			})
			tools.Post("/signin", func(ctx iris.Context) {
				routing.UserTool(routing.LOGIN_USER, ctx)
			})
		}

		user.Get("/{user:string}", routing.UserProfile)
	}

	stats := app.Party("/stats")
	{
		exp := stats.Party("/exps")
		{
			exp.Get("/most.json", func(ctx iris.Context) {
				routing.ViewedExpressions("-", ctx)
			})
			exp.Get("/least.json", func(ctx iris.Context) {
				routing.ViewedExpressions("", ctx)
			})
		}

		imgs := stats.Party("/imgs")
		{
			imgs.Get("/most.json", func(ctx iris.Context) {
				routing.ViewedImages("-", ctx)
			})
			imgs.Get("/least.json", func(ctx iris.Context) {
				routing.ViewedImages("", ctx)
			})
		}
	}

	api := app.Party("/api")
	{
		api.Get("/expressions.json", func(ctx iris.Context) {
			for i := 0; i < len(defaultExpressions); i++ {
				views := database.MongoInstance.GetExpressionByName(defaultExpressions[i].Expression)
				if len(views.ExpressionCmp) <= 0 {
					defaultExpressions[i].Usage = 1
				} else {
					defaultExpressions[i].Usage = views.Usage
				}
			}
			ctx.JSON(defaultExpressions)
		})
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
