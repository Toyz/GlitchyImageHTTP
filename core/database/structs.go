package database

type ArtItem struct {
	ID         string `json:"id"`
	Folder     string `json:"folder"`
	FileName   string `json:"filename"`
	FullPath   string `json:"fullpath"`
	Expression string `json:"expression"`
}
