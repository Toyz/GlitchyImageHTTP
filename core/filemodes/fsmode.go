package filemodes

import (
	"fmt"
	"os"
	"path"

	"github.com/Toyz/GlitchyImageHTTP/core"
)

type FSMode struct{}

func (FSMode) Write(data []byte, name string) string {
	staticFilePath := core.GetEnv("HTTP_UPLOADS_URL", "/img/")

	f, _ := os.Create(path.Join(core.UploadsFolder(), name))
	f.Write(data)
	f.Close()

	return fmt.Sprintf("%s%s", staticFilePath, name)
}
