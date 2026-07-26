package database

import "time"

type Tenant struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Project struct {
	ID               int       `json:"id"`
	TenantID         int       `json:"tenant_id"`
	Slug             string    `json:"slug"`
	Name             string    `json:"name"`
	BasePath         string    `json:"base_path"`
	Provider         string    `json:"provider"`
	Bucket           string    `json:"bucket"`
	Region           string    `json:"region"`
	AccessKeyID      string    `json:"access_key_id,omitempty"`
	SecretAccessKey  string    `json:"secret_access_key,omitempty"`
	CredentialsJSON  string    `json:"credentials_json,omitempty"`
	Endpoint         string    `json:"endpoint,omitempty"`
	DefaultQuality   int       `json:"default_quality"`
	DefaultFormat    string    `json:"default_format"`
	MaxWidth         int       `json:"max_width"`
	MaxHeight        int       `json:"max_height"`
	RPSSec           int       `json:"rps"`
	Burst            int       `json:"burst"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Invitation struct {
	ID        int        `json:"id"`
	Code      string     `json:"code"`
	IsUsed    bool       `json:"is_used"`
	CreatedBy int        `json:"created_by"`
	UsedBy    *int       `json:"used_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

type ProjectMetrics struct {
	ID                int64 `json:"id"`
	ProjectID         int   `json:"project_id"`
	Date              string `json:"date"`
	Hour              int   `json:"hour"`
	Requests          int64 `json:"requests"`
	CacheHits         int64 `json:"cache_hits"`
	CacheMisses       int64 `json:"cache_misses"`
	OriginTransforms  int64 `json:"origin_transforms"`
	BandwidthBytes    int64 `json:"bandwidth_bytes"`
	Errors            int64 `json:"errors"`
	ResponseTimeSumMs int64 `json:"response_time_sum_ms"`
	ResponseCount     int64 `json:"response_count"`
}

type ProjectMetricsDaily struct {
	Date              string `json:"date"`
	Requests          int64  `json:"requests"`
	CacheHits         int64  `json:"cache_hits"`
	CacheMisses       int64  `json:"cache_misses"`
	OriginTransforms  int64  `json:"origin_transforms"`
	BandwidthBytes    int64  `json:"bandwidth_bytes"`
	Errors            int64  `json:"errors"`
	AvgResponseTimeMs int64  `json:"avg_response_time_ms"`
}

type ProjectSummary struct {
	Project
	RequestsToday        int64   `json:"requests_today"`
	CacheHitsToday       int64   `json:"cache_hits_today"`
	CacheMissesToday     int64   `json:"cache_misses_today"`
	BandwidthBytesToday  int64   `json:"bandwidth_bytes_today"`
	ErrorsToday          int64   `json:"errors_today"`
	RequestsTotal        int64   `json:"requests_total"`
	BandwidthBytesTotal  int64   `json:"bandwidth_bytes_total"`
	ErrorsTotal          int64   `json:"errors_total"`
	CacheHitRatio        float64 `json:"cache_hit_ratio"`
}

type TenantMetricsSummary struct {
	TotalRequests       int64   `json:"total_requests"`
	TotalCacheHits      int64   `json:"total_cache_hits"`
	TotalErrors         int64   `json:"total_errors"`
	TotalBandwidthBytes int64   `json:"total_bandwidth_bytes"`
	CacheHitRatio       float64 `json:"cache_hit_ratio"`
	PerProject          []ProjectBriefMetrics `json:"per_project"`
}

type ProjectBriefMetrics struct {
	ProjectID     int    `json:"project_id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Requests      int64  `json:"requests"`
	CacheHits     int64  `json:"cache_hits"`
	BandwidthBytes int64 `json:"bandwidth_bytes"`
	Errors        int64  `json:"errors"`
}
