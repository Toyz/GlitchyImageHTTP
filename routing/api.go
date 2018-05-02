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

	items := database.MongoInstance.GetExpressionsByOrder(mode, limit)

	ctx.JSON(items)
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

		artItems[i] = API_ArtInfo{
			ID:          item.ID,
			URL:         item.FullPath,
			Width:       item.Width,
			Height:      item.Height,
			Size:        item.FileSize,
			Views:       item.Views,
			Expressions: item.Expressions,
		}
	}

	ctx.JSON(artItems)
}

func ViewImageInfo(ctx iris.Context) {
	id := ctx.Params().Get("image")
	err, item := database.MongoInstance.GetUploadInfo(id)

	if err != nil {
		ctx.JSON(JsonError{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(API_ArtInfo{
		ID:          item.ID,
		URL:         item.FullPath,
		Width:       item.Width,
		Height:      item.Height,
		Size:        item.FileSize,
		Views:       item.Views,
		Expressions: item.Expressions,
	})
}
