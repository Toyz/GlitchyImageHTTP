package routing

import (
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
	"github.com/globalsign/mgo/bson"
	"github.com/kataras/iris"
)

func ViewedExpressions(mode string, ctx iris.Context) {
	limit, err := ctx.URLParamInt("limit")
	if limit < 0 || err != nil {
		limit = 20
	}

	items := database.MongoInstance.GetExpressionsByOrder(mode, limit)

	expItems := make([]API_Expression, len(items))

	for i := 0; i < len(items); i++ {
		item := items[i]

		expItems[i] = API_Expression{
			Expression: item.Expression,
			ID:         item.MGID.Hex(),
			Usage:      item.Usage,
		}
	}

	ctx.JSON(expItems)
}

func ViewedImages(mode string, ctx iris.Context) {
	limit, err := ctx.URLParamInt("limit")
	if limit < 0 || err != nil {
		limit = 20
	}

	items := database.MongoInstance.GetArtByOrder(mode, limit)

	artItems := make([]API_ArtInfo, len(items))

	for i := 0; i < len(items); i++ {
		item := items[i]

		if len(item.Expressions) <= 0 && len(item.Expression) > 0 {
			item.Expressions = append(item.Expressions, item.Expression)
		}

		artItems[i] = API_ArtInfo{
			ID:          item.MGID.Hex(),
			URL:         filemodes.GetFileMode().FullPath(item.Folder, item.FileName),
			Width:       item.Width,
			Height:      item.Height,
			Size:        item.FileSize,
			Views:       item.Views,
			Uploaded:    item.Uploaded,
			Expressions: make([]API_Expression, len(item.Expressions)),
		}

		for e := 0; e < len(item.Expressions); e++ {
			exp := item.Expressions[e]

			expItem := database.MongoInstance.GetExpression(exp)
			if len(expItem.Expression) <= 0 {
				expItem = database.ExpressionItem{
					Expression: exp,
					Usage:      1,
					MGID:       bson.NewObjectId(),
				}

				database.MongoInstance.InsertExpression(expItem)
			}

			artItems[i].Expressions[e] = API_Expression{
				Expression: expItem.Expression,
				ID:         expItem.MGID.Hex(),
				Usage:      expItem.Usage,
			}
		}
	}

	ctx.JSON(artItems)
}

func ViewImageInfo(ctx iris.Context) {
	id := ctx.Params().Get("image")
	err, item := database.MongoInstance.GetUploadInfo(id)

	if len(item.Expressions) <= 0 && len(item.Expression) > 0 {
		item.Expressions = append(item.Expressions, item.Expression)
	}

	if err != nil {
		ctx.JSON(JsonError{
			Error: err.Error(),
		})
		return
	}

	artItem := API_ArtInfo{
		ID:          item.MGID.Hex(),
		URL:         filemodes.GetFileMode().FullPath(item.Folder, item.FileName),
		Width:       item.Width,
		Height:      item.Height,
		Size:        item.FileSize,
		Views:       item.Views,
		Uploaded:    item.Uploaded,
		Expressions: make([]API_Expression, len(item.Expressions)),
	}

	for e := 0; e < len(item.Expressions); e++ {
		exp := item.Expressions[e]

		expItem := database.MongoInstance.GetExpression(exp)
		if len(expItem.Expression) <= 0 {
			expItem = database.ExpressionItem{
				Expression: exp,
				Usage:      1,
				MGID:       bson.NewObjectId(),
			}
			database.MongoInstance.InsertExpression(expItem)
		}

		artItem.Expressions[e] = API_Expression{
			Expression: expItem.Expression,
			ID:         expItem.MGID.Hex(),
			Usage:      expItem.Usage,
		}
	}

	ctx.JSON(artItem)
}
