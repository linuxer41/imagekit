package image

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vendemas/imagekit/internal/cache"
	"github.com/vendemas/imagekit/internal/metrics"
	"github.com/vendemas/imagekit/internal/params"
	"github.com/vendemas/imagekit/internal/tenant"
	"github.com/vendemas/imagekit/internal/transform"
)

type MetricsRecorder interface {
	RecordRequest(projectID int, isCacheHit bool, bandwidthBytes int64, durationMs int64, hasError bool, hasTransform bool)
}

type Handler struct {
	projectCache *tenant.ProjectCache
	lruCache     *cache.LRUCache
	recorder     MetricsRecorder
}

func NewHandler(pc *tenant.ProjectCache, c *cache.LRUCache, recorder MetricsRecorder) *Handler {
	return &Handler{projectCache: pc, lruCache: c, recorder: recorder}
}

func (h *Handler) Serve(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	slug := chi.URLParam(r, "slug")
	filePath := chi.URLParam(r, "*")

	var (
		projectID    int
		isCacheHit   bool
		bandwidth    int64
		hasError     bool
		hasTransform bool
	)

	defer func() {
		if h.recorder == nil || projectID == 0 {
			return
		}
		h.recorder.RecordRequest(projectID, isCacheHit, bandwidth, time.Since(start).Milliseconds(), hasError, hasTransform)
	}()

	if slug == "" || filePath == "" {
		hasError = true
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := validateFilePath(filePath); err != nil {
		slog.Warn("path traversal blocked", "slug", slug, "path", filePath)
		hasError = true
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	project, ok := h.projectCache.Get(slug)
	if !ok {
		hasError = true
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	projectID = project.ID

	trParam := r.URL.Query().Get("tr")
	imgParams := params.ParseTR(trParam)

	fullPath := path.Join(project.BasePath, filePath)
	fullPath = strings.TrimPrefix(fullPath, "/")

	if imgParams.HasParams() {
		if imgParams.Width > project.MaxWidth {
			imgParams.Width = project.MaxWidth
		}
		if imgParams.Height > project.MaxHeight {
			imgParams.Height = project.MaxHeight
		}
		if imgParams.Quality <= 0 {
			imgParams.Quality = project.DefaultQuality
		}
		if imgParams.Format == "" {
			accept := r.Header.Get("Accept")
			if supportsFormat(accept, "avif") {
				imgParams.Format = "avif"
			} else if supportsFormat(accept, "webp") {
				imgParams.Format = "webp"
			} else {
				imgParams.Format = params.Format(project.DefaultFormat)
			}
		}
	}

	cacheKey := cacheKeyString(slug, fullPath, imgParams)
	if cached, ct, ok := h.lruCache.Get(cacheKey); ok {
		isCacheHit = true
		bandwidth = int64(len(cached))
		metrics.CacheHits.Inc()
		serveBytes(w, cached, ct, project.MaxWidth)
		metrics.RequestDuration.WithLabelValues(slug, "cache_hit").Observe(time.Since(start).Seconds())
		return
	}
	metrics.CacheMisses.Inc()

	storageStart := time.Now()
	data, err := project.Provider.Get(r.Context(), fullPath)
	metrics.StorageGetDuration.WithLabelValues(project.Name).Observe(time.Since(storageStart).Seconds())
	if err != nil {
		slog.Error("storage get", "slug", slug, "path", fullPath, "error", err)
		metrics.ErrorsTotal.WithLabelValues(slug, "storage_error").Inc()
		hasError = true
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}

	if imgParams.HasParams() {
		transformStart := time.Now()
		out, contentType, err := transform.ProcessImage(data, imgParams)
		metrics.TransformDuration.WithLabelValues(string(imgParams.Format)).Observe(time.Since(transformStart).Seconds())
		if err != nil {
			slog.Error("transform", "slug", slug, "error", err)
			metrics.ErrorsTotal.WithLabelValues(slug, "transform_error").Inc()
			hasError = true
			http.Error(w, "transform failed", http.StatusInternalServerError)
			return
		}

		hasTransform = true
		bandwidth = int64(len(out))
		h.lruCache.Set(cacheKey, out, contentType)
		metrics.RequestsTotal.WithLabelValues(slug, contentType, "true").Inc()
		metrics.RequestDuration.WithLabelValues(slug, "transform").Observe(time.Since(start).Seconds())
		serveBytes(w, out, contentType, project.MaxWidth)
		return
	}

	bandwidth = int64(len(data))
	contentType := detectContentType(data, filePath)
	metrics.RequestsTotal.WithLabelValues(slug, contentType, "false").Inc()
	metrics.RequestDuration.WithLabelValues(slug, "original").Observe(time.Since(start).Seconds())
	serveBytes(w, data, contentType, project.MaxWidth)
}

func serveBytes(w http.ResponseWriter, data []byte, contentType string, maxAge int) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", maxAge))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(data))
	w.Header().Set("ETag", etag)
	w.Write(data)
}

func cacheKeyString(slug, fullPath string, imgParams params.Params) string {
	return fmt.Sprintf("%s:%s:%s", slug, fullPath, imgParams.CacheKey())
}
