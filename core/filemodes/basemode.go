package filemodes

import "github.com/Toyz/GlitchyImageHTTP/core"

type SaveMode interface {
	Setup()
	Write([]byte, string) string
	Path() string
}

func GetFileMode() SaveMode {
	switch core.GetSaveMode() {
	case "fs":
		return &FSMode{}
	case "aws":
		return &CDNMode{}
	}
	return &FSMode{}
}
