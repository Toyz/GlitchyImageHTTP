package filemodes

import (
	"fmt"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/sony/sonyflake"
)

type SaveMode interface {
	Setup()
	Write([]byte, string) (string, string)
	Read(string) []byte
	Path() string
	FullPath(folder, name string) string
}

var flaky *sonyflake.Sonyflake

func init() {
	flaky = sonyflake.NewSonyflake(sonyflake.Settings{})
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

func GetID(id string) string {
	idx, _ := flaky.NextID()

	return fmt.Sprintf("%x", idx)
}
