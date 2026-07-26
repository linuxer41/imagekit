package database

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *Repo) RegisterTenant(ctx context.Context, name, email, password string) (*Tenant, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}

	var t Tenant
	err = r.pool.QueryRow(ctx, `
		INSERT INTO image_tenants (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, is_active, created_at, updated_at`,
		name, email, string(hash),
	).Scan(&t.ID, &t.Name, &t.Email, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert tenant: %w", err)
	}

	return &t, nil
}

func (r *Repo) LoginTenant(ctx context.Context, email, password string) (*Tenant, error) {
	var t Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password, is_active, created_at, updated_at
		FROM image_tenants WHERE email = $1 AND is_active = true`, email,
	).Scan(&t.ID, &t.Name, &t.Email, &t.Password, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find tenant: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(t.Password), []byte(password)); err != nil {
		return nil, nil
	}

	return &t, nil
}

func (r *Repo) GetTenant(ctx context.Context, id int) (*Tenant, error) {
	var t Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password, is_active, created_at, updated_at
		FROM image_tenants WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Email, &t.Password, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return &t, nil
}

func (r *Repo) GetTenantByID(ctx context.Context, id int) (*Tenant, error) {
	var t Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password, is_active, created_at, updated_at
		FROM image_tenants WHERE id = $1 AND is_active = true`, id,
	).Scan(&t.ID, &t.Name, &t.Email, &t.Password, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return &t, nil
}

func (r *Repo) UpdateTenantInfo(ctx context.Context, id int, name string) error {
	_, err := r.pool.Exec(ctx, `UPDATE image_tenants SET name=$1, updated_at=$2 WHERE id=$3`,
		name, time.Now(), id)
	return err
}

func (r *Repo) LoginUser(ctx context.Context, username, password string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, password, role, is_active, created_at
		FROM image_users WHERE username = $1 AND is_active = true`, username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, nil
	}

	return &u, nil
}

func (r *Repo) CreateProject(ctx context.Context, tenantID int, slug, name, basePath, provider, bucket, region, accessKey, secretKey, credsJSON, endpoint string, defaultQuality int, defaultFormat string, maxWidth, maxHeight, rps, burst int) (*Project, error) {
	var p Project
	err := r.pool.QueryRow(ctx, `
		INSERT INTO image_projects (tenant_id, slug, name, base_path, provider, bucket, region, access_key_id, secret_access_key, credentials_json, endpoint, default_quality, default_format, max_width, max_height, rps, burst)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, tenant_id, slug, name, base_path, provider, bucket, region, access_key_id, secret_access_key, credentials_json, endpoint, default_quality, default_format, max_width, max_height, rps, burst, is_active, created_at, updated_at`,
		tenantID, slug, name, basePath, provider, bucket, region, accessKey, secretKey, credsJSON, endpoint, defaultQuality, defaultFormat, maxWidth, maxHeight, rps, burst,
	).Scan(&p.ID, &p.TenantID, &p.Slug, &p.Name, &p.BasePath, &p.Provider, &p.Bucket, &p.Region, &p.AccessKeyID, &p.SecretAccessKey, &p.CredentialsJSON, &p.Endpoint, &p.DefaultQuality, &p.DefaultFormat, &p.MaxWidth, &p.MaxHeight, &p.RPSSec, &p.Burst, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return &p, nil
}

func (r *Repo) GetProject(ctx context.Context, id int) (*Project, error) {
	var p Project
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, slug, name, base_path, provider, bucket, region, access_key_id, secret_access_key, credentials_json, endpoint, default_quality, default_format, max_width, max_height, rps, burst, is_active, created_at, updated_at
		FROM image_projects WHERE id = $1`, id,
	).Scan(&p.ID, &p.TenantID, &p.Slug, &p.Name, &p.BasePath, &p.Provider, &p.Bucket, &p.Region, &p.AccessKeyID, &p.SecretAccessKey, &p.CredentialsJSON, &p.Endpoint, &p.DefaultQuality, &p.DefaultFormat, &p.MaxWidth, &p.MaxHeight, &p.RPSSec, &p.Burst, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &p, nil
}

func (r *Repo) GetProjectBySlug(ctx context.Context, slug string) (*Project, error) {
	var p Project
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, slug, name, base_path, provider, bucket, region, access_key_id, secret_access_key, credentials_json, endpoint, default_quality, default_format, max_width, max_height, rps, burst, is_active, created_at, updated_at
		FROM image_projects WHERE slug = $1 AND is_active = true`, slug,
	).Scan(&p.ID, &p.TenantID, &p.Slug, &p.Name, &p.BasePath, &p.Provider, &p.Bucket, &p.Region, &p.AccessKeyID, &p.SecretAccessKey, &p.CredentialsJSON, &p.Endpoint, &p.DefaultQuality, &p.DefaultFormat, &p.MaxWidth, &p.MaxHeight, &p.RPSSec, &p.Burst, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get project by slug: %w", err)
	}
	return &p, nil
}

func (r *Repo) ListProjects(ctx context.Context, tenantID int) ([]*Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, slug, name, base_path, provider, bucket, region, access_key_id, secret_access_key, credentials_json, endpoint, default_quality, default_format, max_width, max_height, rps, burst, is_active, created_at, updated_at
		FROM image_projects WHERE tenant_id = $1 ORDER BY id`, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Slug, &p.Name, &p.BasePath, &p.Provider, &p.Bucket, &p.Region, &p.AccessKeyID, &p.SecretAccessKey, &p.CredentialsJSON, &p.Endpoint, &p.DefaultQuality, &p.DefaultFormat, &p.MaxWidth, &p.MaxHeight, &p.RPSSec, &p.Burst, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, &p)
	}
	return projects, nil
}

func (r *Repo) UpdateProject(ctx context.Context, id int, slug, name, basePath, provider, bucket, region, accessKey, secretKey, credsJSON, endpoint string, defaultQuality int, defaultFormat string, maxWidth, maxHeight, rps, burst int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE image_projects SET slug=$1, name=$2, base_path=$3, provider=$4, bucket=$5, region=$6, access_key_id=$7, secret_access_key=$8, credentials_json=$9, endpoint=$10, default_quality=$11, default_format=$12, max_width=$13, max_height=$14, rps=$15, burst=$16, updated_at=$17
		WHERE id=$18`,
		slug, name, basePath, provider, bucket, region, accessKey, secretKey, credsJSON, endpoint, defaultQuality, defaultFormat, maxWidth, maxHeight, rps, burst, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	return nil
}

func (r *Repo) DeleteProject(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `UPDATE image_projects SET is_active=false, updated_at=$1 WHERE id=$2`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}

func (r *Repo) ListAllActiveProjects(ctx context.Context) ([]*Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, slug, name, base_path, provider, bucket, region, access_key_id, secret_access_key, credentials_json, endpoint, default_quality, default_format, max_width, max_height, rps, burst, is_active, created_at, updated_at
		FROM image_projects WHERE is_active = true ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list active projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Slug, &p.Name, &p.BasePath, &p.Provider, &p.Bucket, &p.Region, &p.AccessKeyID, &p.SecretAccessKey, &p.CredentialsJSON, &p.Endpoint, &p.DefaultQuality, &p.DefaultFormat, &p.MaxWidth, &p.MaxHeight, &p.RPSSec, &p.Burst, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		projects = append(projects, &p)
	}
	return projects, nil
}

func (r *Repo) ListAllTenants(ctx context.Context) ([]*Tenant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, email, password, is_active, created_at, updated_at
		FROM image_tenants ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Password, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		tenants = append(tenants, &t)
	}
	return tenants, nil
}

func (r *Repo) ListAllUsers(ctx context.Context) ([]*User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, username, role, is_active, created_at
		FROM image_users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *Repo) CountTenants(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM image_tenants`).Scan(&n)
	return n, err
}

func (r *Repo) CountProjects(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM image_projects`).Scan(&n)
	return n, err
}

func (r *Repo) CountUsers(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM image_users`).Scan(&n)
	return n, err
}

func (r *Repo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, password, role, is_active, created_at
		FROM image_users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &u, nil
}

// Invitations

func generateCode() (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 8)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		code[i] = chars[n.Int64()]
	}
	return string(code), nil
}

func (r *Repo) CreateInvitation(ctx context.Context, createdBy int) (*Invitation, error) {
	code, err := generateCode()
	if err != nil {
		return nil, fmt.Errorf("generate code: %w", err)
	}

	var inv Invitation
	err = r.pool.QueryRow(ctx, `
		INSERT INTO image_invitations (code, created_by)
		VALUES ($1, $2)
		RETURNING id, code, is_used, created_by, used_by, created_at, used_at`,
		code, createdBy,
	).Scan(&inv.ID, &inv.Code, &inv.IsUsed, &inv.CreatedBy, &inv.UsedBy, &inv.CreatedAt, &inv.UsedAt)
	if err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}
	return &inv, nil
}

func (r *Repo) ListInvitations(ctx context.Context) ([]*Invitation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, is_used, created_by, used_by, created_at, used_at
		FROM image_invitations ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	defer rows.Close()

	var invs []*Invitation
	for rows.Next() {
		var inv Invitation
		if err := rows.Scan(&inv.ID, &inv.Code, &inv.IsUsed, &inv.CreatedBy, &inv.UsedBy, &inv.CreatedAt, &inv.UsedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		invs = append(invs, &inv)
	}
	return invs, nil
}

func (r *Repo) ValidateAndUseInvitation(ctx context.Context, code string, tenantID int) error {
	var inv Invitation
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, is_used FROM image_invitations WHERE code = $1`, code,
	).Scan(&inv.ID, &inv.Code, &inv.IsUsed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("invalid invitation code")
		}
		return fmt.Errorf("find invitation: %w", err)
	}

	if inv.IsUsed {
		return fmt.Errorf("invitation code already used")
	}

	_, err = r.pool.Exec(ctx, `UPDATE image_invitations SET is_used=true, used_by=$1, used_at=$2 WHERE id=$3`,
		tenantID, time.Now(), inv.ID)
	return err
}

// Metrics

func (r *Repo) UpsertProjectMetrics(ctx context.Context, pm *ProjectMetrics) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO project_metrics (project_id, date, hour, requests, cache_hits, cache_misses, origin_transforms, bandwidth_bytes, errors, response_time_sum_ms, response_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (project_id, date, hour) DO UPDATE SET
			requests = project_metrics.requests + EXCLUDED.requests,
			cache_hits = project_metrics.cache_hits + EXCLUDED.cache_hits,
			cache_misses = project_metrics.cache_misses + EXCLUDED.cache_misses,
			origin_transforms = project_metrics.origin_transforms + EXCLUDED.origin_transforms,
			bandwidth_bytes = project_metrics.bandwidth_bytes + EXCLUDED.bandwidth_bytes,
			errors = project_metrics.errors + EXCLUDED.errors,
			response_time_sum_ms = project_metrics.response_time_sum_ms + EXCLUDED.response_time_sum_ms,
			response_count = project_metrics.response_count + EXCLUDED.response_count`,
		pm.ProjectID, pm.Date, pm.Hour, pm.Requests, pm.CacheHits, pm.CacheMisses, pm.OriginTransforms, pm.BandwidthBytes, pm.Errors, pm.ResponseTimeSumMs, pm.ResponseCount,
	)
	return err
}

func (r *Repo) GetProjectMetricsDaily(ctx context.Context, projectID int, days int) ([]ProjectMetricsDaily, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT date,
			SUM(requests) as requests,
			SUM(cache_hits) as cache_hits,
			SUM(cache_misses) as cache_misses,
			SUM(origin_transforms) as origin_transforms,
			SUM(bandwidth_bytes) as bandwidth_bytes,
			SUM(errors) as errors,
			CASE WHEN SUM(response_count) > 0 THEN SUM(response_time_sum_ms) / SUM(response_count) ELSE 0 END as avg_response_time_ms
		FROM project_metrics
		WHERE project_id = $1 AND date >= CURRENT_DATE - $2::INT
		GROUP BY date
		ORDER BY date`, projectID, days)
	if err != nil {
		return nil, fmt.Errorf("get project metrics: %w", err)
	}
	defer rows.Close()

	var result []ProjectMetricsDaily
	for rows.Next() {
		var m ProjectMetricsDaily
		if err := rows.Scan(&m.Date, &m.Requests, &m.CacheHits, &m.CacheMisses, &m.OriginTransforms, &m.BandwidthBytes, &m.Errors, &m.AvgResponseTimeMs); err != nil {
			return nil, fmt.Errorf("scan metrics: %w", err)
		}
		result = append(result, m)
	}
	return result, nil
}

func (r *Repo) GetProjectSummary(ctx context.Context, projectID int) (*ProjectSummary, error) {
	ps := &ProjectSummary{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, slug, name, base_path, provider, bucket, region, access_key_id, secret_access_key, credentials_json, endpoint,
			default_quality, default_format, max_width, max_height, rps, burst, is_active, created_at, updated_at
		FROM image_projects WHERE id = $1`, projectID,
	).Scan(&ps.ID, &ps.TenantID, &ps.Slug, &ps.Name, &ps.BasePath, &ps.Provider, &ps.Bucket, &ps.Region, &ps.AccessKeyID, &ps.SecretAccessKey, &ps.CredentialsJSON, &ps.Endpoint, &ps.DefaultQuality, &ps.DefaultFormat, &ps.MaxWidth, &ps.MaxHeight, &ps.RPSSec, &ps.Burst, &ps.IsActive, &ps.CreatedAt, &ps.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get project for summary: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(requests), 0) as requests_today,
			COALESCE(SUM(cache_hits), 0) as cache_hits_today,
			COALESCE(SUM(cache_misses), 0) as cache_misses_today,
			COALESCE(SUM(bandwidth_bytes), 0) as bandwidth_bytes_today,
			COALESCE(SUM(errors), 0) as errors_today
		FROM project_metrics
		WHERE project_id = $1 AND date = CURRENT_DATE AND hour >= 0`, projectID)
	if err := row.Scan(&ps.RequestsToday, &ps.CacheHitsToday, &ps.CacheMissesToday, &ps.BandwidthBytesToday, &ps.ErrorsToday); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get today metrics: %w", err)
	}

	row2 := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(requests), 0) as total_requests,
			COALESCE(SUM(bandwidth_bytes), 0) as total_bandwidth,
			COALESCE(SUM(errors), 0) as total_errors,
			COALESCE(SUM(cache_hits), 0) + COALESCE(SUM(cache_misses), 0) as total_cacheable
		FROM project_metrics WHERE project_id = $1`, projectID)
	if err := row2.Scan(&ps.RequestsTotal, &ps.BandwidthBytesTotal, &ps.ErrorsTotal, &ps.CacheHitsToday); err != nil {
		_ = err // ignore partial data
	}
	_ = row2 // reuse cache_hits_today for the total: compute ratio
	totalCacheHits := ps.CacheHitsToday // reused field as total
	totalCacheable := ps.CacheMissesToday + totalCacheHits
	if totalCacheable > 0 {
		ps.CacheHitRatio = float64(totalCacheHits) / float64(totalCacheable) * 100
	}

	return ps, nil
}

func (r *Repo) GetTenantMetricsSummary(ctx context.Context, tenantID int, days int) (*TenantMetricsSummary, error) {
	ts := &TenantMetricsSummary{}

	totalRow := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(pm.requests), 0),
			COALESCE(SUM(pm.cache_hits), 0),
			COALESCE(SUM(pm.errors), 0),
			COALESCE(SUM(pm.bandwidth_bytes), 0)
		FROM project_metrics pm
		JOIN image_projects p ON p.id = pm.project_id
		WHERE p.tenant_id = $1 AND pm.date >= CURRENT_DATE - $2::INT`, tenantID, days)
	if err := totalRow.Scan(&ts.TotalRequests, &ts.TotalCacheHits, &ts.TotalErrors, &ts.TotalBandwidthBytes); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get tenant totals: %w", err)
	}
	totalCacheable := ts.TotalCacheHits + (ts.TotalRequests - ts.TotalCacheHits)
	if totalCacheable > 0 {
		ts.CacheHitRatio = float64(ts.TotalCacheHits) / float64(totalCacheable) * 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.slug, p.name,
			COALESCE(SUM(pm.requests), 0),
			COALESCE(SUM(pm.cache_hits), 0),
			COALESCE(SUM(pm.bandwidth_bytes), 0),
			COALESCE(SUM(pm.errors), 0)
		FROM image_projects p
		LEFT JOIN project_metrics pm ON pm.project_id = p.id AND pm.date >= CURRENT_DATE - $2::INT
		WHERE p.tenant_id = $1 AND p.is_active = true
		GROUP BY p.id, p.slug, p.name
		ORDER BY p.slug`, tenantID, days)
	if err != nil {
		return nil, fmt.Errorf("get per-project metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pb ProjectBriefMetrics
		if err := rows.Scan(&pb.ProjectID, &pb.Slug, &pb.Name, &pb.Requests, &pb.CacheHits, &pb.BandwidthBytes, &pb.Errors); err != nil {
			return nil, fmt.Errorf("scan per-project: %w", err)
		}
		ts.PerProject = append(ts.PerProject, pb)
	}
	return ts, nil
}

func (r *Repo) GetAdminGlobalMetrics(ctx context.Context, days int) (*TenantMetricsSummary, error) {
	ts := &TenantMetricsSummary{}

	totalRow := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(requests), 0),
			COALESCE(SUM(cache_hits), 0),
			COALESCE(SUM(errors), 0),
			COALESCE(SUM(bandwidth_bytes), 0)
		FROM project_metrics
		WHERE date >= CURRENT_DATE - $1::INT`, days)
	if err := totalRow.Scan(&ts.TotalRequests, &ts.TotalCacheHits, &ts.TotalErrors, &ts.TotalBandwidthBytes); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get admin totals: %w", err)
	}
	totalCacheable := ts.TotalCacheHits + (ts.TotalRequests - ts.TotalCacheHits)
	if totalCacheable > 0 {
		ts.CacheHitRatio = float64(ts.TotalCacheHits) / float64(totalCacheable) * 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.slug, p.name,
			COALESCE(SUM(pm.requests), 0),
			COALESCE(SUM(pm.cache_hits), 0),
			COALESCE(SUM(pm.bandwidth_bytes), 0),
			COALESCE(SUM(pm.errors), 0)
		FROM image_projects p
		LEFT JOIN project_metrics pm ON pm.project_id = p.id AND pm.date >= CURRENT_DATE - $1::INT
		WHERE p.is_active = true
		GROUP BY p.id, p.slug, p.name
		ORDER BY p.slug`, days)
	if err != nil {
		return nil, fmt.Errorf("get admin per-project: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pb ProjectBriefMetrics
		if err := rows.Scan(&pb.ProjectID, &pb.Slug, &pb.Name, &pb.Requests, &pb.CacheHits, &pb.BandwidthBytes, &pb.Errors); err != nil {
			return nil, fmt.Errorf("scan admin per-project: %w", err)
		}
		ts.PerProject = append(ts.PerProject, pb)
	}
	return ts, nil
}
