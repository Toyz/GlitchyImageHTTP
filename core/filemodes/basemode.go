package filemodes

import (
	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/kataras/iris"
)

type SaveMode interface {
	Setup()
	Write([]byte, string) (string, string)
	Read(string) []byte
	Path() string
	FullPath(folder, name string) string
	StaticPath(*iris.Application)
}

func GetFileMode() SaveMode {
	var mode SaveMode

	switch core.GetSaveMode() {
	case "fs":
		mode = &FSMode{}
	case "aws":
		mode = &CDNMode{}
	}

	mode.Setup()
	return mode
}
