package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/vendemas/imagekit/internal/cache"
	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/metrics"
	"github.com/vendemas/imagekit/internal/tenant"
)

type ImagesHandler struct {
	repo         *database.Repo
	projectCache *tenant.ProjectCache
	cache        *cache.LRUCache
}

func NewImagesHandler(repo *database.Repo, pc *tenant.ProjectCache, c *cache.LRUCache) *ImagesHandler {
	return &ImagesHandler{repo: repo, projectCache: pc, cache: c}
}

func (h *ImagesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	slug := chi.URLParam(r, "slug")
	filePath := chi.URLParam(r, "*")

	if slug == "" || filePath == "" {
		http.Error(w, `{"error":"slug and path are required"}`, http.StatusBadRequest)
		return
	}

	if err := validateImagePath(filePath); err != nil {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	cached, ok := h.projectCache.Get(slug)
	if !ok {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	if cached.TenantID != tenantID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	fullPath := path.Join(cached.BasePath, filePath)
	fullPath = strings.TrimPrefix(fullPath, "/")

	if err := cached.Provider.Delete(r.Context(), fullPath); err != nil {
		slog.Error("delete image", "slug", slug, "path", fullPath, "error", err)
		metrics.ErrorsTotal.WithLabelValues(slug, "delete_error").Inc()
		http.Error(w, `{"error":"failed to delete image"}`, http.StatusInternalServerError)
		return
	}

	// invalidate cache for this path
	// we clear all cache entries that match this slug+path prefix
	h.cache.InvalidatePrefix(slug + ":" + fullPath)

	metrics.RequestsTotal.WithLabelValues(slug, "", "delete").Inc()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func validateImagePath(filePath string) error {
	if filepath.Clean(filePath) != filePath {
		return fmt.Errorf("invalid path")
	}
	for _, part := range strings.Split(filePath, string(filepath.Separator)) {
		if part == ".." {
			return fmt.Errorf("path traversal")
		}
	}
	return nil
}
