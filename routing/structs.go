package routing

import (
	"time"
)

type HomePage struct {
	Token      string
	Error      string
	Expression string
}

type UploadResult struct {
	ID string `json:"id,omitempty"`
}

type JsonError struct {
	Error string `json:"error,omitempty"`
}

type API_ArtInfo struct {
	ID          string           `json:"id"`
	URL         string           `json:"URL"`
	Width       int              `json:"width"`
	Height      int              `json:"height"`
	Size        int              `json:"size"`
	Views       int              `json:"views"`
	Uploaded    time.Time        `json:"uploaded"`
	Expressions []API_Expression `json:"expressions"`
}

type API_Expression struct {
	Expression string `json:"expression"`
	Category   string `json:"category,omitempty"`
	Usage      int    `json:"usage"`
	ID         string `json:"id"`
}
