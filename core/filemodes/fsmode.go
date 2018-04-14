package filemodes

import (
	"fmt"
	"os"
	"path"

	"github.com/Toyz/GlitchyImageHTTP/core"
)

type FSMode struct{}

func (fs *FSMode) Setup() {
	if _, err := os.Stat(fs.Path()); os.IsNotExist(err) {
		os.MkdirAll(fs.Path(), 0644)
	}
}

func (fs *FSMode) Write(data []byte, name string) (string, string) {
	staticFilePath := core.GetEnv("HTTP_UPLOADS_URL", "/img/")
	physicalUploadsFolder := fs.Path()

	f, _ := os.Create(path.Join(physicalUploadsFolder, name))
	f.Write(data)
	f.Close()

	return fmt.Sprintf("%s%s", staticFilePath, name), ""
}

func (fs *FSMode) Read(path string) []byte {
	return make([]byte, 0)
}

func (*FSMode) Path() string {
	return core.GetEnv("FS_UPLOADS_FOLDER", fmt.Sprintf("%s%s", core.AssetsFolder(), "uploads"))
}
