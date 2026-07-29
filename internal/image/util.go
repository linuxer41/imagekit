package image

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func validateFilePath(filePath string) error {
	decoded, err := url.PathUnescape(filePath)
	if err != nil {
		decoded = filePath
	}
	clean := filepath.Clean(decoded)
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		if part == ".." {
			return fmt.Errorf("path traversal detected")
		}
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
