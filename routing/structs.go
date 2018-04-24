package routing

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
