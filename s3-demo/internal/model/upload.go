package model

type GenerateUploadURLRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

type GenerateUploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int    `json:"expires_in"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
