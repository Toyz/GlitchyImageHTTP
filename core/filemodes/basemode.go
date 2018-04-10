package filemodes

import (
	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/rs/xid"
)

type SaveMode interface {
	Setup()
	Write([]byte, string) (string, string)
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

func GetID(id string) string {
	guid := xid.New()

	return guid.String()
}
