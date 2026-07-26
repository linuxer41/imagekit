package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vendemas/imagekit/internal/database"
)

type MetricsHandler struct {
	repo *database.Repo
}

func NewMetricsHandler(repo *database.Repo) *MetricsHandler {
	return &MetricsHandler{repo: repo}
}

func (h *MetricsHandler) ProjectMetrics(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	project, err := h.repo.GetProject(r.Context(), projectID)
	if err != nil || project == nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	tenantID := tenantIDFromContext(r)
	if tenantID > 0 && project.TenantID != tenantID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	days := intParam(r, "days", 30)
	if days > 90 {
		days = 90
	}

	daily, err := h.repo.GetProjectMetricsDaily(r.Context(), projectID, days)
	if err != nil {
		slog.Error("get project metrics", "project_id", projectID, "error", err)
		http.Error(w, `{"error":"failed to get metrics"}`, http.StatusInternalServerError)
		return
	}
	if daily == nil {
		daily = []database.ProjectMetricsDaily{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"daily": daily,
	})
}

func (h *MetricsHandler) ProjectSummary(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	summary, err := h.repo.GetProjectSummary(r.Context(), projectID)
	if err != nil || summary == nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	tenantID := tenantIDFromContext(r)
	if tenantID > 0 && summary.TenantID != tenantID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *MetricsHandler) TenantMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	days := intParam(r, "days", 7)
	if days > 90 {
		days = 90
	}

	ts, err := h.repo.GetTenantMetricsSummary(r.Context(), tenantID, days)
	if err != nil {
		slog.Error("get tenant metrics", "tenant_id", tenantID, "error", err)
		http.Error(w, `{"error":"failed to get metrics"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, ts)
}

func (h *MetricsHandler) AdminGlobalMetrics(w http.ResponseWriter, r *http.Request) {
	days := intParam(r, "days", 7)
	if days > 90 {
		days = 90
	}

	ts, err := h.repo.GetAdminGlobalMetrics(r.Context(), days)
	if err != nil {
		slog.Error("get admin global metrics", "error", err)
		http.Error(w, `{"error":"failed to get metrics"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, ts)
}

func (h *MetricsHandler) AdminProjectMetrics(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	days := intParam(r, "days", 30)
	if days > 90 {
		days = 90
	}

	daily, err := h.repo.GetProjectMetricsDaily(r.Context(), projectID, days)
	if err != nil {
		slog.Error("get admin project metrics", "project_id", projectID, "error", err)
		http.Error(w, `{"error":"failed to get metrics"}`, http.StatusInternalServerError)
		return
	}
	if daily == nil {
		daily = []database.ProjectMetricsDaily{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"daily": daily,
	})
}

func (h *MetricsHandler) AdminProjectSummary(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	summary, err := h.repo.GetProjectSummary(r.Context(), projectID)
	if err != nil || summary == nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func intParam(r *http.Request, name string, def int) int {
	v := r.URL.Query().Get(name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
