package routing

import (
	"strings"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/kataras/iris"
	ss "github.com/kataras/iris/sessions"
)

const (
	CREATE_USER = 0x1
	LOGIN_USER  = 0x2
	LOGOUT_USER = 0x3
)

func UserJoin(ctx iris.Context) {
}

func UserLogin(ctx iris.Context) {
	data := ctx.GetViewData()["Header"].(core.HeaderMetaData)

	data.Title = "Login"
	ctx.ViewData("Header", data)
	ctx.View("login.html")
}

func UserProfile(ctx iris.Context) {

}

func UserTool(tool int, ctx iris.Context) {
	sess := core.SessionManager.Session.Start(ctx)

	switch tool {
	case CREATE_USER:
		join(ctx, sess)
		return
	case LOGIN_USER:
		login(ctx, sess)
		return
	case LOGOUT_USER:
		sess.Delete("logged_in")
		sess.Delete("user")
		ctx.Redirect("/")
		return
	}
}

func login(ctx iris.Context, sess *ss.Session) {
	email := strings.TrimSpace(ctx.FormValueDefault("email", ""))
	pass := ctx.FormValueDefault("pass", "")

	if len(email) <= 0 || len(pass) <= 0 {
		ctx.JSON(JsonError{
			Error: "Email or Password must not be empty!",
		})

		return
	}

	user := database.MongoInstance.GetUserByEmail(email)
	if len(user.Email) <= 0 {
		ctx.JSON(JsonError{
			Error: "Email/Password is invalid",
		})

		return
	}

	if user.Password != pass {
		ctx.JSON(JsonError{
			Error: "Email/Password is invalid",
		})

		return
	}

	sess.Set("logged_in", true)

	user.Password = ""
	sess.Set("user", user)

	user = database.MongoInstance.SetLastLogin(user)

	ctx.JSON(UploadResult{
		ID: user.MGID.Hex(), // UserID for redirect to there profile
	})
}

func join(ctx iris.Context, sess *ss.Session) {
	email := strings.TrimSpace(ctx.FormValueDefault("email", ""))
	pass := ctx.FormValueDefault("pass", "")
	username := ctx.FormValueDefault("uname", "")

	if len(email) <= 0 || len(pass) <= 0 || len(username) <= 0 {
		ctx.JSON(JsonError{
			Error: "Email, Password or Username must not be empty!",
		})

		return
	}

	user := database.MongoInstance.GetUserByEmail(email)
	if len(user.Email) > 0 && strings.EqualFold(email, user.Email) {
		ctx.JSON(JsonError{
			Error: "Email already in use",
		})

		return
	}

	user = database.MongoInstance.InsertUser(database.User{
		Email:    email,
		Username: username,
		Password: pass,
	})

	sess.Set("logged_in", true)

	user.Password = ""
	sess.Set("user", user)

	ctx.JSON(UploadResult{
		ID: user.MGID.Hex(), // UserID for redirect to there profile
	})
}
