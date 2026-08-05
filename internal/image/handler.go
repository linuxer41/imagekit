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

const (
	// immutableCacheAge es el max-age para imágenes inmutables (1 año).
	immutableCacheAge = 31536000
	// notFoundCacheTTL es cuánto tiempo se cachean los 404 (evita golpear storage repetidamente).
	notFoundCacheTTL = 30 * time.Second
)

type MetricsRecorder interface {
	RecordRequest(projectID int, isCacheHit bool, bandwidthBytes int64, durationMs int64, hasError bool, hasTransform bool)
}

type Handler struct {
	projectCache *tenant.ProjectCache
	lruCache     *cache.LRUCache
	recorder     MetricsRecorder
	// transformSem limita las transformaciones concurrentes para evitar
	// saturar la CPU de libvips con picos de misses.
	transformSem chan struct{}
}

func NewHandler(pc *tenant.ProjectCache, c *cache.LRUCache, recorder MetricsRecorder, maxConcurrentTransforms int) *Handler {
	if maxConcurrentTransforms <= 0 {
		maxConcurrentTransforms = 8
	}
	return &Handler{
		projectCache: pc,
		lruCache:     c,
		recorder:     recorder,
		transformSem: make(chan struct{}, maxConcurrentTransforms),
	}
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
		// Formato fijo: si no se pide f- explícito, se usa el default del
		// proyecto. NO se negocia por header Accept para mantener la clave de
		// caché estable entre navegadores (Chrome avif / Safari webp) y evitar
		// transformaciones duplicadas + encodes AVIF lentos.
		if imgParams.Format == "" {
			imgParams.Format = params.Format(project.DefaultFormat)
		}
	}

	cacheKey := cacheKeyString(slug, fullPath, imgParams)
	if cached, ct, ok := h.lruCache.Get(cacheKey); ok {
		isCacheHit = true
		bandwidth = int64(len(cached))
		metrics.CacheHits.Inc()
		serveBytes(w, cached, ct, immutableCacheAge)
		metrics.RequestDuration.WithLabelValues(slug, "cache_hit").Observe(time.Since(start).Seconds())
		return
	}
	metrics.CacheMisses.Inc()

	// Negative caching: si la imagen no existe, no volver a golpear el
	// storage durante unos segundos.
	if h.lruCache.GetNotFound(cacheKey) {
		metrics.CacheHits.Inc()
		slog.Warn("image not found (cached 404)", "slug", slug, "path", fullPath)
		hasError = true
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}

	storageStart := time.Now()
	data, err := project.Provider.Get(r.Context(), fullPath)
	metrics.StorageGetDuration.WithLabelValues(project.Name).Observe(time.Since(storageStart).Seconds())
	if err != nil {
		slog.Error("storage get", "slug", slug, "path", fullPath, "error", err)
		metrics.ErrorsTotal.WithLabelValues(slug, "storage_error").Inc()
		hasError = true
		h.lruCache.SetNotFound(cacheKey, notFoundCacheTTL)
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}

	if imgParams.HasParams() {
		h.transformSem <- struct{}{}
		transformStart := time.Now()
		out, contentType, err := transform.ProcessImage(data, imgParams)
		<-h.transformSem
		metrics.TransformDuration.WithLabelValues(string(imgParams.Format)).Observe(time.Since(transformStart).Seconds())
		if err != nil {
			slog.Warn("transform failed, serving original", "slug", slug, "path", fullPath, "error", err)
			metrics.ErrorsTotal.WithLabelValues(slug, "transform_error").Inc()
			contentType = detectContentType(data, filePath)
			out = data
		}

		hasTransform = err == nil
		bandwidth = int64(len(out))
		h.lruCache.Set(cacheKey, out, contentType)
		metrics.RequestsTotal.WithLabelValues(slug, contentType, "true").Inc()
		metrics.RequestDuration.WithLabelValues(slug, "transform").Observe(time.Since(start).Seconds())
		serveBytes(w, out, contentType, immutableCacheAge)
		return
	}

	bandwidth = int64(len(data))
	contentType := detectContentType(data, filePath)
	metrics.RequestsTotal.WithLabelValues(slug, contentType, "false").Inc()
	metrics.RequestDuration.WithLabelValues(slug, "original").Observe(time.Since(start).Seconds())
	serveBytes(w, data, contentType, immutableCacheAge)
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
