package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/vendemas/imagekit/internal/auth"
	"github.com/vendemas/imagekit/internal/config"
	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/middleware"
	"github.com/vendemas/imagekit/internal/router"
	"github.com/vendemas/imagekit/internal/server"
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

	panelDir := os.Getenv("PANEL_DIR")
	adminDir := os.Getenv("ADMIN_DIR")

	handler := router.NewAPIRouter(jwt, repo, corsMW, panelDir, adminDir)

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
