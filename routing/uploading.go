package routing

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
	"github.com/kataras/iris"
	glitch "github.com/sugoiuguu/go-glitch"
)

var allowedFileTypes = []string{"image/jpeg", "image/png", "image/jpg", "image/gif"}
var saveMode filemodes.SaveMode

func validateFormFeilds(ctx iris.Context) (error, []string, string) {
	var expressions []string

	exps := ctx.FormValues()
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
		return errors.New("Only 5 expressions are allowed"), nil, ""
	}

	return nil, expressions, ""
}

func validateFileUpload(ctx iris.Context) (error, multipart.File, *multipart.FileHeader, string) {
	file, fHeader, err := ctx.FormFile("uploadfile")
	if err != nil {
		return errors.New("File cannot be empty"), nil, nil, ""
	}
	defer file.Close()

	cntType := core.GetMimeType(file)
	if ok, _ := core.InArray(cntType, allowedFileTypes); !ok {
		return errors.New("File type is not allowed only PNG, JPEG, GIF allowed"), nil, nil, ""
	}

	return nil, file, fHeader, cntType
}

func SaveImage(dataBuff *bytes.Buffer, cntType string, OrgFileName string, bounds image.Rectangle, expressions []string, userId string) (error, string) {
	buff := dataBuff.Bytes()
	defer dataBuff.Reset()

	md5Sum := core.GetMD5(buff)
	fileName := fmt.Sprintf("%s.%s", md5Sum, core.MimeToExtension(cntType))

	_, folder := saveMode.Write(buff, fileName)

	expressionIds := make([]bson.ObjectId, len(expressions))
	for e := 0; e < len(expressions); e++ {
		exp := expressions[e]
		local := database.MongoInstance.GetExpressionByName(exp)

		if len(local.ExpressionCmp) > 0 {
			expressionIds[e] = local.MGID
			continue
		}

		local = database.MongoInstance.AddExpression(database.ExpressionItem{
			Expression: exp,
		})

		expressionIds[e] = local.MGID
	}

	item := database.ArtItem{
		FileName:    fileName,
		OrgFileName: OrgFileName,
		Folder:      folder,
		Uploaded:    time.Now(),
		FileSize:    binary.Size(buff),
		Width:       bounds.Max.X,
		Height:      bounds.Max.Y,
	}

	var uploaderId bson.ObjectId
	if len(userId) > 0 {
		if bson.IsObjectIdHex(userId) {
			uploaderId = bson.ObjectIdHex(userId)
		} else {
			uploaderId = ""
		}
	} else {
		uploaderId = ""
	}

	art := database.MongoInstance.SetImageInfo(item)
	uploadInfo := database.Upload{
		ImageID:     art.MGID,
		Expressions: expressionIds,
		Views:       0,
		User:        uploaderId,
	}

	upload := database.MongoInstance.AddUpload(uploadInfo)

	return nil, upload.MGID.Hex()
}

func processImage(file multipart.File, mime string, expressions []string) (error, *bytes.Buffer, image.Rectangle) {
	buff := new(bytes.Buffer)
	var bounds image.Rectangle
	switch strings.ToLower(mime) {
	case "image/gif":
		err, by, rect := gifImage(file, expressions)
		bounds = rect

		if err != nil {
			return err, nil, bounds
		}
		buff = by
		break
	default:
		img, _, err := image.Decode(file)
		if err != nil {
			return err, nil, image.Rectangle{}
		}
		bounds = img.Bounds()

		out := img
		for _, expression := range expressions {
			expr, err := glitch.CompileExpression(expression)
			if err != nil {
				return err, nil, bounds
			}

			exp := database.MongoInstance.GetExpressionByName(expression)
			if len(exp.ExpressionCmp) > 0 {
				database.MongoInstance.UpdateExpressionUsage(exp.MGID)
			} else {
				database.MongoInstance.AddExpression(database.ExpressionItem{
					Expression: expression,
				})
			}

			newImage, err := expr.JumblePixels(out)
			if err != nil {
				out = nil
				return err, nil, bounds
			}
			out = newImage
			newImage = nil
		}

		switch strings.ToLower(mime) {
		case "image/png":
			png.Encode(buff, out)
			break
		case "image/jpg", "image/jpeg":
			jpeg.Encode(buff, out, nil)
			break
		}
	}

	return nil, buff, bounds
}

func gifImage(file multipart.File, expressions []string) (error, *bytes.Buffer, image.Rectangle) {
	var bounds image.Rectangle
	buff := new(bytes.Buffer)
	lGif, err := gif.DecodeAll(file)

	bounds = lGif.Image[0].Bounds()
	if err != nil {
		return err, nil, image.Rectangle{}
	}

	out := lGif
	for _, expression := range expressions {
		expr, err := glitch.CompileExpression(expression)
		if err != nil {
			return err, nil, bounds
		}

		exp := database.MongoInstance.GetExpressionByName(expression)
		if len(exp.ExpressionCmp) > 0 {
			database.MongoInstance.UpdateExpressionUsage(exp.MGID)
		} else {
			database.MongoInstance.AddExpression(database.ExpressionItem{
				Expression: expression,
			})
		}

		newImage, err := expr.JumbleGIFPixels(out)
		if err != nil {
			out = nil
			return err, nil, bounds
		}
		out = newImage
		newImage = nil
	}

	err = gif.EncodeAll(buff, out)
	if err != nil {
		return err, nil, bounds
	}

	return nil, buff, bounds
}

func Upload(ctx iris.Context) {
	sess := core.SessionManager.Session.Start(ctx)

	saveMode = filemodes.GetFileMode()

	ctx.SetMaxRequestBodySize(65 << 20) // 65mb because we can

	err, expressions, _ := validateFormFeilds(ctx)
	if err != nil {
		ctx.JSON(JsonError{
			Error: err.Error(),
		})
		return
	}

	err, file, header, mime := validateFileUpload(ctx)
	if err != nil {
		ctx.JSON(JsonError{
			Error: err.Error(),
		})
		return
	}

	err, data, bounds := processImage(file, mime, expressions)
	if err != nil {
		ctx.JSON(JsonError{
			Error: err.Error(),
		})
		return
	}

	userId := ""

	if sess.GetBooleanDefault("logged_in", false) {
		d := sess.Get("user").(map[string]interface{})
		userId = d["id"].(string)
	}
	err, id := SaveImage(data, mime, header.Filename, bounds, expressions, userId)
	if err != nil {
		ctx.JSON(JsonError{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(UploadResult{
		ID: id,
	})
}
