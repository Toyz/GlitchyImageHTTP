package filemodes

import (
	"fmt"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/sony/sonyflake"
)

type SaveMode interface {
	Setup()
	Write([]byte, string) (string, string)
	Path() string
}

var flaky *sonyflake.Sonyflake

func init() {
	flaky = sonyflake.NewSonyflake(sonyflake.Settings{})
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
	idx, _ := flaky.NextID()

	return fmt.Sprintf("%x", idx)
}
