package core

import (
	"fmt"
	"os"
	"path"
)

func AssetsFolder() string {
	return GetEnv("ASSETS_FOLDER", "./assets/")
}

func UploadsFolder() string {
	return GetEnv("UPLOAD_FOLDER", "./assets/uploads")
}

func GetTemplateFilePath(name string) string {
	return path.Join(AssetsFolder(), "tmpls", fmt.Sprintf("%s.html", name))
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
