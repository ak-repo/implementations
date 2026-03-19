package handler

import (
	"encoding/json"
	"net/http"

	"s3-upload-demo/internal/model"
	"s3-upload-demo/internal/service"
)

type UploadHandler struct {
	service *service.UploadService
}

func NewUploadHandler(service *service.UploadService) *UploadHandler {
	return &UploadHandler{
		service: service,
	}
}

func (h *UploadHandler) GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.GenerateUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Filename == "" {
		h.respondError(w, http.StatusBadRequest, "filename is required")
		return
	}

	if req.ContentType == "" {
		h.respondError(w, http.StatusBadRequest, "content_type is required")
		return
	}

	uploadURL, fileKey, err := h.service.GenerateUploadURL(req.Filename, req.ContentType)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	resp := model.GenerateUploadURLResponse{
		UploadURL: uploadURL,
		FileKey:   fileKey,
		ExpiresIn: 300,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *UploadHandler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{Error: message})
}
