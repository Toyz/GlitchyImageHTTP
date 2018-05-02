package routing

import (
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/kataras/iris"
)

func ViewedExpressions(mode string, ctx iris.Context) {
	limit, err := ctx.URLParamInt("limit")
	if limit < 0 || err != nil {
		limit = 20
	}

	items := database.MongoInstance.GetMostUsedExpression(mode, limit)

	ctx.JSON(items)
}
