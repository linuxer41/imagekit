package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func RunMigrations(pool *pgxpool.Pool, adminUser, adminPass string) error {
	ctx := context.Background()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS image_users (
			id       SERIAL PRIMARY KEY,
			username VARCHAR(100) UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role     VARCHAR(20) DEFAULT 'admin',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS image_tenants (
			id         SERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			email      VARCHAR(255) UNIQUE NOT NULL,
			password   TEXT NOT NULL,
			is_active  BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS image_projects (
			id               SERIAL PRIMARY KEY,
			tenant_id        INT NOT NULL REFERENCES image_tenants(id) ON DELETE CASCADE,
			slug             VARCHAR(100) UNIQUE NOT NULL,
			name             VARCHAR(255) DEFAULT '',
			base_path        VARCHAR(500) DEFAULT '/',
			provider         VARCHAR(10) NOT NULL DEFAULT 'gcs',
			bucket           VARCHAR(255) NOT NULL,
			region           VARCHAR(50) DEFAULT '',
			access_key_id    TEXT DEFAULT '',
			secret_access_key TEXT DEFAULT '',
			credentials_json  TEXT DEFAULT '',
			endpoint         TEXT DEFAULT '',
			default_quality  INT DEFAULT 80,
			default_format   VARCHAR(10) DEFAULT 'webp',
			max_width        INT DEFAULT 4096,
			max_height       INT DEFAULT 4096,
			rps              INT DEFAULT 100,
			burst            INT DEFAULT 200,
			is_active        BOOLEAN DEFAULT true,
			created_at       TIMESTAMPTZ DEFAULT NOW(),
			updated_at       TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_tenant ON image_projects(tenant_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_slug ON image_projects(slug)`,
		`CREATE TABLE IF NOT EXISTS image_invitations (
			id         SERIAL PRIMARY KEY,
			code       VARCHAR(20) UNIQUE NOT NULL,
			is_used    BOOLEAN DEFAULT false,
			created_by INT REFERENCES image_users(id),
			used_by    INT REFERENCES image_tenants(id),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			used_at    TIMESTAMPTZ DEFAULT NULL
		)`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	// seed admin user
	var exists bool
	err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM image_users WHERE username = $1)`, adminUser).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check admin: %w", err)
	}

	if !exists {
		hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash admin: %w", err)
		}
		_, err = pool.Exec(ctx, `INSERT INTO image_users (username, password, role) VALUES ($1, $2, 'admin')`, adminUser, string(hash))
		if err != nil {
			return fmt.Errorf("seed admin: %w", err)
		}
		slog.Info("admin user seeded", "username", adminUser)
	}

	// metrics tables
	metricsQueries := []string{
		`CREATE TABLE IF NOT EXISTS project_metrics (
			id                BIGSERIAL PRIMARY KEY,
			project_id        INT NOT NULL REFERENCES image_projects(id) ON DELETE CASCADE,
			date              DATE NOT NULL,
			hour              INT NOT NULL DEFAULT -1,
			requests          BIGINT NOT NULL DEFAULT 0,
			cache_hits        BIGINT NOT NULL DEFAULT 0,
			cache_misses      BIGINT NOT NULL DEFAULT 0,
			origin_transforms BIGINT NOT NULL DEFAULT 0,
			bandwidth_bytes   BIGINT NOT NULL DEFAULT 0,
			errors            BIGINT NOT NULL DEFAULT 0,
			response_time_sum_ms BIGINT NOT NULL DEFAULT 0,
			response_count    BIGINT NOT NULL DEFAULT 0,
			UNIQUE(project_id, date, hour)
		)`,
		`CREATE TABLE IF NOT EXISTS project_storage (
			id             BIGSERIAL PRIMARY KEY,
			project_id     INT NOT NULL REFERENCES image_projects(id) ON DELETE CASCADE,
			date           DATE NOT NULL,
			storage_bytes  BIGINT NOT NULL DEFAULT 0,
			image_count    INT NOT NULL DEFAULT 0,
			UNIQUE(project_id, date)
		)`,
	}
	for _, q := range metricsQueries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("migrate metrics: %w", err)
		}
	}

	// add endpoint column if missing
	pool.Exec(ctx, `ALTER TABLE image_projects ADD COLUMN IF NOT EXISTS endpoint TEXT DEFAULT ''`)

	slog.Info("migrations complete")
	return nil
}
