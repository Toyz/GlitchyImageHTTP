package core

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
)

func AssetsFolder() string {
	return GetEnv("ASSETS_FOLDER", "./assets/")
}

// TODO: Move template stuff into it's own "render" class
func GetTemplateFilePath(name string) string {
	return path.Join(AssetsFolder(), "tmpls", fmt.Sprintf("%s.html", name))
}

func GetTemplateFolder() string {
	return path.Join(AssetsFolder(), "tmpls")
}

// TODO: Move this into a "public" type system to let me extend the render for easier loading of assets
func GetPublicFolder() string {
	return path.Join(AssetsFolder(), "public")
}

func GetSaveMode() string {
	return strings.ToLower(GetEnv("SAVE_MODE", "fs"))
}

// TODO: add things we actually need helpers for here
func InArray(val string, array []string) (exists bool, index int) {
	exists = false
	index = -1

	for i, v := range array {
		if val == v {
			index = i
			exists = true
			return
		}
	}

	return
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetMimeType(file multipart.File) string {
	fileHeader := make([]byte, 512)
	if _, err := file.Read(fileHeader); err != nil {
		log.Println(err)
		return ""
	}
	if _, err := file.Seek(0, 0); err != nil {
		log.Println(err)
		return ""
	}

	return http.DetectContentType(fileHeader)
}

func GetMimeTypeFromBytes(fileHeader []byte) string {
	return http.DetectContentType(fileHeader)
}

func MimeToExtension(mime string) string {
	switch strings.ToLower(mime) {
	case "image/png":
		return "png"
	case "image/jpg", "image/jpeg":
		return "jpg"
	default:
		return ""
	}
}
