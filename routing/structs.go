package routing

type HomePage struct {
	Token      string
	Error      string
	Expression string
}

type UploadResult struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}
