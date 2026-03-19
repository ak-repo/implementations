package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"s3-upload-demo/internal/s3"
)

type UploadService struct {
	s3Client *s3.Client
}

func NewUploadService(s3Client *s3.Client) *UploadService {
	return &UploadService{
		s3Client: s3Client,
	}
}

func (s *UploadService) GenerateUploadURL(filename, contentType string) (string, string, error) {
	id := uuid.New().String()
	extension := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, extension)
	sanitized := sanitizeFilename(baseName)

	key := fmt.Sprintf("uploads/%s-%s%s", id, sanitized, extension)

	url, err := s.s3Client.GeneratePresignedUploadURL(key, contentType)
	if err != nil {
		return "", "", err
	}

	return url, key, nil
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()
	if sanitized == "" {
		sanitized = "file"
	}

	return sanitized
}
