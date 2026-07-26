package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/vendemas/imagekit/internal/auth"
	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/storage"
)

type ProjectsHandler struct {
	repo *database.Repo
}

func NewProjectsHandler(repo *database.Repo) *ProjectsHandler {
	return &ProjectsHandler{repo: repo}
}

type createProjectReq struct {
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	BasePath        string `json:"base_path"`
	Provider        string `json:"provider"`
	Bucket          string `json:"bucket"`
	Region          string `json:"region,omitempty"`
	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	CredentialsJSON string `json:"credentials_json,omitempty"`
	Endpoint       string `json:"endpoint,omitempty"`
	DefaultQuality  int    `json:"default_quality"`
	DefaultFormat   string `json:"default_format"`
	MaxWidth        int    `json:"max_width"`
	MaxHeight       int    `json:"max_height"`
	RPSSec          int    `json:"rps"`
	Burst           int    `json:"burst"`
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	projects, err := h.repo.ListProjects(r.Context(), tenantID)
	if err != nil {
		slog.Error("list projects", "error", err)
		http.Error(w, `{"error":"failed to list projects"}`, http.StatusInternalServerError)
		return
	}

	if projects == nil {
		projects = []*database.Project{}
	}

	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req createProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Slug = strings.TrimSpace(req.Slug)
	if req.Slug == "" {
		http.Error(w, `{"error":"slug is required"}`, http.StatusBadRequest)
		return
	}

	if req.Provider == "" {
		req.Provider = "gcs"
	}
	if req.Bucket == "" {
		http.Error(w, `{"error":"bucket is required"}`, http.StatusBadRequest)
		return
	}
	if req.Provider != "gcs" && req.Provider != "s3" && req.Provider != "rustfs" {
		http.Error(w, `{"error":"provider must be 'gcs', 's3', or 'rustfs'"}`, http.StatusBadRequest)
		return
	}
	if req.DefaultQuality <= 0 {
		req.DefaultQuality = 80
	}
	if req.DefaultFormat == "" {
		req.DefaultFormat = "webp"
	}
	if req.MaxWidth <= 0 {
		req.MaxWidth = 4096
	}
	if req.MaxHeight <= 0 {
		req.MaxHeight = 4096
	}
	if req.RPSSec <= 0 {
		req.RPSSec = 100
	}
	if req.Burst <= 0 {
		req.Burst = 200
	}

	// test connection before saving
	dummyProject := &database.Project{
		Provider:        req.Provider,
		Bucket:          req.Bucket,
		Region:          req.Region,
		AccessKeyID:     req.AccessKeyID,
		SecretAccessKey: req.SecretAccessKey,
		CredentialsJSON: req.CredentialsJSON,
		Endpoint:        req.Endpoint,
	}
	prov, err := storage.NewProvider(dummyProject)
	if err != nil {
		slog.Error("create provider for test", "error", err)
		http.Error(w, `{"error":"invalid storage configuration"}`, http.StatusBadRequest)
		return
	}
	if err := prov.TestConnection(r.Context()); err != nil {
		slog.Error("storage test connection", "error", err)
		http.Error(w, `{"error":"storage connection failed: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	project, err := h.repo.CreateProject(r.Context(), tenantID, req.Slug, req.Name, req.BasePath, req.Provider, req.Bucket, req.Region, req.AccessKeyID, req.SecretAccessKey, req.CredentialsJSON, req.Endpoint, req.DefaultQuality, req.DefaultFormat, req.MaxWidth, req.MaxHeight, req.RPSSec, req.Burst)

	if err != nil {
		slog.Error("create project", "error", err)
		http.Error(w, `{"error":"failed to create project, slug may already exist"}`, http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	project, err := h.repo.GetProject(r.Context(), id)
	if err != nil || project == nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	if project.TenantID != tenantID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectsHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	project, err := h.repo.GetProject(r.Context(), id)
	if err != nil || project == nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	if project.TenantID != tenantID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	var req createProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	slug := ifStr(req.Slug != "", req.Slug, project.Slug)
	name := ifStr(req.Name != "", req.Name, project.Name)
	basePath := ifStr(req.BasePath != "", req.BasePath, project.BasePath)
	provider := ifStr(req.Provider != "", req.Provider, project.Provider)
	bucket := ifStr(req.Bucket != "", req.Bucket, project.Bucket)
	region := ifStr(req.Region != "" || req.Provider != "", req.Region, project.Region)
	ak := ifStr(req.AccessKeyID != "", req.AccessKeyID, project.AccessKeyID)
	sk := ifStr(req.SecretAccessKey != "", req.SecretAccessKey, project.SecretAccessKey)
	cj := ifStr(req.CredentialsJSON != "", req.CredentialsJSON, project.CredentialsJSON)
	ep := ifStr(req.Endpoint != "", req.Endpoint, project.Endpoint)
	dq := ifVal(req.DefaultQuality > 0, req.DefaultQuality, project.DefaultQuality)
	df := ifStr(req.DefaultFormat != "", req.DefaultFormat, project.DefaultFormat)
	mw := ifVal(req.MaxWidth > 0, req.MaxWidth, project.MaxWidth)
	mh := ifVal(req.MaxHeight > 0, req.MaxHeight, project.MaxHeight)
	rps := ifVal(req.RPSSec > 0, req.RPSSec, project.RPSSec)
	burst := ifVal(req.Burst > 0, req.Burst, project.Burst)

	// test connection if storage fields changed
	if req.Bucket != "" || req.Provider != "" || req.AccessKeyID != "" || req.SecretAccessKey != "" || req.CredentialsJSON != "" || req.Endpoint != "" {
		dummyProject := &database.Project{
			Provider:        provider,
			Bucket:          bucket,
			Region:          region,
			AccessKeyID:     ak,
			SecretAccessKey: sk,
			CredentialsJSON: cj,
			Endpoint:        ep,
		}
		prov, err := storage.NewProvider(dummyProject)
		if err != nil {
			http.Error(w, `{"error":"invalid storage configuration"}`, http.StatusBadRequest)
			return
		}
		if err := prov.TestConnection(r.Context()); err != nil {
			http.Error(w, `{"error":"storage connection failed"}`, http.StatusBadRequest)
			return
		}
	}

	if err := h.repo.UpdateProject(r.Context(), id, slug, name, basePath, provider, bucket, region, ak, sk, cj, ep, dq, df, mw, mh, rps, burst); err != nil {
		slog.Error("update project", "error", err)
		http.Error(w, `{"error":"failed to update project"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	project, err := h.repo.GetProject(r.Context(), id)
	if err != nil || project == nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	if project.TenantID != tenantID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	if err := h.repo.DeleteProject(r.Context(), id); err != nil {
		slog.Error("delete project", "error", err)
		http.Error(w, `{"error":"failed to delete project"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func tenantIDFromContext(r *http.Request) int {
	sub := auth.GetSubject(r.Context())
	if sub == "" {
		return 0
	}
	id, err := strconv.Atoi(sub)
	if err != nil {
		return 0
	}
	return id
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func ifStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func ifVal(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}
