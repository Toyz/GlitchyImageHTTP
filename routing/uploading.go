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

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/Toyz/GlitchyImageHTTP/core/database"
	"github.com/Toyz/GlitchyImageHTTP/core/filemodes"
	"github.com/globalsign/mgo"
	"github.com/kataras/iris"
	glitch "github.com/sugoiuguu/go-glitch"
)

var allowedFileTypes = []string{"image/jpeg", "image/png", "image/jpg", "image/gif"}
var saveMode filemodes.SaveMode

func validateFormFeilds(ctx iris.Context) (error, []string) {
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
		return errors.New("Only 5 expressions are allowed"), nil
	}

	return nil, expressions
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

func SaveImage(buff []byte, cntType string, OrgFileName string, bounds image.Rectangle, expressions []string) (error, string) {
	md5Sum := core.GetMD5(buff)
	idx := filemodes.GetID(md5Sum)
	fileName := fmt.Sprintf("%s.%s", md5Sum, core.MimeToExtension(cntType))

	actualFileName, folder := saveMode.Write(buff, fileName)

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

	err := c.Insert(&database.ArtItem{
		ID:          idx,
		FileName:    fileName,
		OrgFileName: OrgFileName,
		Folder:      folder,
		FullPath:    actualFileName,
		Expression:  expression,
		Expressions: expressions,
		Views:       0,
		Uploaded:    time.Now(),
		FileSize:    binary.Size(buff),
		Width:       bounds.Max.X,
		Height:      bounds.Max.Y,
	})

	if err != nil {
		buff = nil
		return err, ""
	}

	return nil, idx
}

func processImage(file multipart.File, mime string, expressions []string) (error, []byte, image.Rectangle) {
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

		if (bounds.Max.X * bounds.Max.Y) > (1920 * 1080) {
			img = nil
			return errors.New("Max image size is 1920x1080 (1080p)"), nil, bounds
		}

		out := img
		for _, expression := range expressions {
			expr, err := glitch.CompileExpression(expression)
			if err != nil {
				return err, nil, bounds
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

	return nil, buff.Bytes(), bounds
}

func gifImage(file multipart.File, expressions []string) (error, *bytes.Buffer, image.Rectangle) {
	var bounds image.Rectangle
	buff := new(bytes.Buffer)
	lGif, err := gif.DecodeAll(file)

	bounds = lGif.Image[0].Bounds()

	if (bounds.Max.X * bounds.Max.Y) > (1024 * 768) {
		return errors.New("Max image size is 1024x768 for GIFs"), nil, bounds
	}

	if err != nil {
		return err, nil, image.Rectangle{}
	}

	out := lGif
	for _, expression := range expressions {
		expr, err := glitch.CompileExpression(expression)
		if err != nil {
			return err, nil, bounds
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
	saveMode = filemodes.GetFileMode()

	ctx.SetMaxRequestBodySize(5 << 20) // 5mb because we can

	err, expressions := validateFormFeilds(ctx)
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

	err, id := SaveImage(data, mime, header.Filename, bounds, expressions)
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
