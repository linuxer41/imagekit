package image

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

func validateFilePath(filePath string) error {
	clean := filepath.Clean(filePath)
	if strings.Contains(clean, "..") {
		return fmt.Errorf("path traversal detected")
	}
	if strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "\\") {
		return fmt.Errorf("path must not start with separator")
	}
	return nil
}

func detectContentType(data []byte, path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != "" {
		ct := mime.TypeByExtension(ext)
		if ct != "" {
			return ct
		}
	}
	return http.DetectContentType(data)
}

func supportsFormat(accept string, f string) bool {
	if accept == "" {
		return false
	}
	switch f {
	case "webp":
		return strings.Contains(accept, "image/webp")
	case "avif":
		return strings.Contains(accept, "image/avif")
	}
	return false
}
