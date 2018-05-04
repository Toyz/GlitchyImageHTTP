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

	page, err := ctx.URLParamInt("page")
	if page <= 0 || err != nil {
		page = 1
	}

	items := database.MongoInstance.OrderExpression(mode, page-1, limit)

	expItems := make([]API_Expression, len(items))

	for e := 0; e < len(items); e++ {
		item := items[e]
		expItem := API_Expression{
			Expression: item.Expression,
			Usage:      item.Usage,
			ID:         item.MGID.Hex(),
			Categories: make([]API_Category, len(item.CatIDs)),
		}

		if len(item.CatIDs) > 0 {
			for c := 0; c < len(item.CatIDs); c++ {
				cat := database.MongoInstance.GetCategory(item.CatIDs[c])

				expItem.Categories[c] = API_Category{
					Name: cat.Name,
					ID:   cat.MGID.Hex(),
				}
			}
		}

		expItems[e] = expItem
	}

	ctx.JSON(expItems)
}

func ViewedImages(mode string, ctx iris.Context) {
	limit, err := ctx.URLParamInt("limit")
	if limit < 0 || err != nil {
		limit = 20
	}

	page, err := ctx.URLParamInt("page")
	if page <= 0 || err != nil {
		page = 1
	}

	items := database.MongoInstance.OrderUploads(mode, page-1, limit)

	artItems := make([]API_ArtInfo, len(items))

	for i := 0; i < len(items); i++ {
		item := items[i]
		imageMeta := database.MongoInstance.GetImageInfo(item.ImageID)

		artItems[i] = API_ArtInfo{
			ID:          item.MGID.Hex(),
			URL:         filemodes.GetFileMode().FullPath(imageMeta.Folder, imageMeta.FileName),
			Width:       imageMeta.Width,
			Height:      imageMeta.Height,
			Size:        imageMeta.FileSize,
			Views:       item.Views,
			Uploaded:    imageMeta.Uploaded,
			Expressions: make([]API_Expression, len(item.Expressions)),
		}

		for e := 0; e < len(item.Expressions); e++ {
			exp := item.Expressions[e]
			item := database.MongoInstance.GetExpression(exp)
			expItem := API_Expression{
				Expression: item.Expression,
				Usage:      item.Usage,
				ID:         item.MGID.Hex(),
				Categories: make([]API_Category, len(item.CatIDs)),
			}

			if len(item.CatIDs) > 0 {
				for c := 0; c < len(item.CatIDs); c++ {
					cat := database.MongoInstance.GetCategory(item.CatIDs[c])

					expItem.Categories[c] = API_Category{
						Name: cat.Name,
						ID:   cat.MGID.Hex(),
					}
				}
			}

			artItems[i].Expressions[e] = expItem
		}
	}

	ctx.JSON(artItems)
}

func ViewImageInfo(ctx iris.Context) {
	id := ctx.Params().Get("image")
	if !bson.IsObjectIdHex(id) {
		ctx.JSON(JsonError{
			Error: "Invalid ID",
		})
		return
	}

	upload := database.MongoInstance.GetUpload(bson.ObjectIdHex(id))
	image := database.MongoInstance.GetImageInfo(upload.ImageID)

	artItem := API_ArtInfo{
		ID:          upload.MGID.Hex(),
		URL:         filemodes.GetFileMode().FullPath(image.Folder, image.FileName),
		Width:       image.Width,
		Height:      image.Height,
		Size:        image.FileSize,
		Views:       upload.Views,
		Uploaded:    image.Uploaded,
		Expressions: make([]API_Expression, len(upload.Expressions)),
	}

	for e := 0; e < len(upload.Expressions); e++ {
		exp := upload.Expressions[e]
		item := database.MongoInstance.GetExpression(exp)
		expItem := API_Expression{
			Expression: item.Expression,
			Usage:      item.Usage,
			ID:         item.MGID.Hex(),
			Categories: make([]API_Category, len(item.CatIDs)),
		}

		if len(item.CatIDs) > 0 {
			for c := 0; c < len(item.CatIDs); c++ {
				cat := database.MongoInstance.GetCategory(item.CatIDs[c])

				expItem.Categories[c] = API_Category{
					Name: cat.Name,
					ID:   cat.MGID.Hex(),
				}
			}
		}

		artItem.Expressions[e] = expItem
	}

	ctx.JSON(artItem)
}
