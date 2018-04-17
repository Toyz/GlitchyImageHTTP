package database

import "time"

type ArtItem struct {
	ID          string    `json:"id"`
	Folder      string    `json:"folder"`
	FileName    string    `json:"filename"`
	FullPath    string    `json:"fullpath"`
	Expression  string    `json:"expression"`  // set if only one expression is used (can be empty)
	Expressions []string  `json:"expressions"` // set if multiable are used (will always be 0 if empty)
	Views       int       `json:"views"`
	Uploaded    time.Time `json:"uploaded_on"`
	FileSize    int       `json:"file_size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
}
