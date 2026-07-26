package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vendemas/imagekit/internal/cache"
	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/image"
	"github.com/vendemas/imagekit/internal/middleware"
	"github.com/vendemas/imagekit/internal/tenant"
)

func NewImagerRouter(
	projectCache *tenant.ProjectCache,
	lruCache *cache.LRUCache,
	rateLimiter *middleware.PerProjectLimiter,
	corsMiddleware *middleware.CORS,
	repo *database.Repo,
	recorder image.MetricsRecorder,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(corsMiddleware.Handler)

	healthHandler := newHealthHandler(repo)
	r.Get("/health", healthHandler.Deep)
	r.Handle("/metrics", promhttp.Handler())

	imageHandler := image.NewHandler(projectCache, lruCache, recorder)
	r.With(middleware.RateLimit(rateLimiter)).Get("/{slug}/{*}", imageHandler.Serve)

	return r
}
