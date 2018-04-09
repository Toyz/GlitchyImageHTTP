package filemodes

type SaveMode interface {
	Write([]byte, string) string
	Path() string
}
