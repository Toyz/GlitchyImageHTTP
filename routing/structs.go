package routing

import "github.com/Toyz/GlitchyImageHTTP/core/database"

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
	ID          string                    `json:"id"`
	URL         string                    `json:"URL"`
	Width       int                       `json:"width"`
	Height      int                       `json:"height"`
	Size        int                       `json:"size"`
	Views       int                       `json:"views"`
	Expressions []database.ExpressionItem `json:"expressions"`
}
