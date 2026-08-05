//go:build cgo

package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/vendemas/imagekit/internal/cache"
	"github.com/vendemas/imagekit/internal/config"
	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/metricsrecorder"
	"github.com/vendemas/imagekit/internal/middleware"
	"github.com/vendemas/imagekit/internal/router"
	"github.com/vendemas/imagekit/internal/server"
	"github.com/vendemas/imagekit/internal/tenant"

	"github.com/davidbyttow/govips/v2/vips"
)

func main() {
	_ = godotenv.Load()
	setupLogger()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	vips.Startup(nil)
	defer vips.Shutdown()
	slog.Info("libvips initialized", "version", vips.Version)

	pool, err := database.InitPool(cfg)
	if err != nil {
		slog.Error("database init", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database connected")

	if err := database.RunMigrations(pool, cfg.AdminUser, cfg.AdminPass); err != nil {
		slog.Error("migrations", "error", err)
		os.Exit(1)
	}

	repo := database.NewRepo(pool)
	projectCache := tenant.NewProjectCache(repo)
	lruCache := cache.NewLRUCache(cfg.CacheSizeMB, cfg.CacheTTL)
	rateLimiter := middleware.NewPerProjectLimiter()
	corsMW := middleware.NewCORS(cfg.CORSOrigins)

	ctx := context.Background()
	projectCache.Start(ctx)
	defer projectCache.Stop()

	recorder := metricsrecorder.NewRecorder(repo)
	recorder.Start(ctx)
	defer recorder.Stop()

	handler := router.NewImagerRouter(projectCache, lruCache, rateLimiter, corsMW, repo, recorder, cfg.TransformConcurrency)

	server.Run(cfg.ImagerHTTPAddr, handler)
}

func setupLogger() {
	var level slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))
}
