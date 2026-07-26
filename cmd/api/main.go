package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/vendemas/imagekit/internal/auth"
	"github.com/vendemas/imagekit/internal/cache"
	"github.com/vendemas/imagekit/internal/config"
	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/middleware"
	"github.com/vendemas/imagekit/internal/router"
	"github.com/vendemas/imagekit/internal/server"
	"github.com/vendemas/imagekit/internal/tenant"
)

func main() {
	_ = godotenv.Load()
	setupLogger()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

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
	jwt := auth.NewJWT(cfg.JWTSecret)
	corsMW := middleware.NewCORS(cfg.CORSOrigins)

	projectCache := tenant.NewProjectCache(repo)
	lruCache := cache.NewLRUCache(cfg.CacheSizeMB, cfg.CacheTTL)

	ctx := context.Background()
	projectCache.Start(ctx)
	defer projectCache.Stop()

	panelDir := os.Getenv("PANEL_DIR")
	adminDir := os.Getenv("ADMIN_DIR")

	handler := router.NewAPIRouter(jwt, repo, corsMW, projectCache, lruCache, panelDir, adminDir)

	server.Run(cfg.HTTPAddr, handler)
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
