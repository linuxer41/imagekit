package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr      string
	ImagerHTTPAddr string
	DBHost        string
	DBPort        int
	DBUser        string
	DBPass        string
	DBName        string
	SSLMode       string
	MinConns      int
	MaxConns      int
	JWTSecret     string
	LogLevel      string
	CacheSizeMB   int
	CacheTTL      int
	AdminUser     string
	AdminPass     string
	CORSOrigins   string
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func Load() *Config {
	return &Config{
		HTTPAddr:       addPort(env("HTTP_ADDR", ""), ":8080"),
		ImagerHTTPAddr: addPort(env("IMAGER_HTTP_ADDR", ""), ":9000"),
		DBHost:      env("DB_HOST", "127.0.0.1"),
		DBPort:      envInt("DB_PORT", 5432),
		DBUser:      env("DB_USER", "linuxer"),
		DBPass:      env("DB_PASSWORD", ""),
		DBName:      env("DB_NAME", "vendemas_v2"),
		SSLMode:     env("DB_SSLMODE", "disable"),
		MinConns:    envInt("DB_MIN_CONNS", 2),
		MaxConns:    envInt("DB_MAX_CONNS", 10),
		JWTSecret:   env("JWT_SECRET", ""),
		LogLevel:    env("LOG_LEVEL", "info"),
		CacheSizeMB: envInt("CACHE_SIZE_MB", 256),
		CacheTTL:    envInt("CACHE_TTL_SEC", 86400),
		AdminUser:    env("ADMIN_USER", "admin"),
		AdminPass:    env("ADMIN_PASSWORD", "admin123"),
		CORSOrigins:  env("CORS_ORIGINS", "*"),
	}
}

func addPort(v string, def string) string {
	if v == "" {
		return def
	}
	if v[0] != ':' {
		return ":" + v
	}
	return v
}

func (c *Config) Validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.MaxConns <= 0 {
		return fmt.Errorf("DB_MAX_CONNS must be > 0")
	}
	if c.MinConns > c.MaxConns {
		return fmt.Errorf("DB_MIN_CONNS (%d) must be <= DB_MAX_CONNS (%d)", c.MinConns, c.MaxConns)
	}
	return nil
}
