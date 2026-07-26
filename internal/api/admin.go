package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vendemas/imagekit/internal/auth"
	"github.com/vendemas/imagekit/internal/database"
)

type AdminHandler struct {
	repo *database.Repo
}

func NewAdminHandler(repo *database.Repo) *AdminHandler {
	return &AdminHandler{repo: repo}
}

func adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.GetRole(r.Context()) != "admin" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.repo.CountTenants(r.Context())
	if err != nil {
		slog.Error("count tenants", "error", err)
	}
	projects, err := h.repo.CountProjects(r.Context())
	if err != nil {
		slog.Error("count projects", "error", err)
	}
	users, err := h.repo.CountUsers(r.Context())
	if err != nil {
		slog.Error("count users", "error", err)
	}
	writeJSON(w, http.StatusOK, map[string]int{
		"tenants":  tenants,
		"projects": projects,
		"users":    users,
	})
}

func (h *AdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.repo.ListAllTenants(r.Context())
	if err != nil {
		slog.Error("list tenants", "error", err)
		http.Error(w, `{"error":"failed to list tenants"}`, http.StatusInternalServerError)
		return
	}
	if tenants == nil {
		tenants = []*database.Tenant{}
	}
	writeJSON(w, http.StatusOK, tenants)
}

func (h *AdminHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), id)
	if err != nil || tenant == nil {
		http.Error(w, `{"error":"tenant not found"}`, http.StatusNotFound)
		return
	}

	projects, err := h.repo.ListProjects(r.Context(), id)
	if err != nil {
		projects = []*database.Project{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tenant":   tenant,
		"projects": projects,
	})
}

func (h *AdminHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	allTenants, err := h.repo.ListAllActiveProjects(r.Context())
	if err != nil {
		slog.Error("list all projects", "error", err)
		http.Error(w, `{"error":"failed to list projects"}`, http.StatusInternalServerError)
		return
	}
	if allTenants == nil {
		allTenants = []*database.Project{}
	}
	writeJSON(w, http.StatusOK, allTenants)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.ListAllUsers(r.Context())
	if err != nil {
		slog.Error("list users", "error", err)
		http.Error(w, `{"error":"failed to list users"}`, http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []*database.User{}
	}
	writeJSON(w, http.StatusOK, users)
}

type adminCreateProjectReq struct {
	createProjectReq
	TenantID int `json:"tenant_id"`
}

func (h *AdminHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	invs, err := h.repo.ListInvitations(r.Context())
	if err != nil {
		slog.Error("list invitations", "error", err)
		http.Error(w, `{"error":"failed to list invitations"}`, http.StatusInternalServerError)
		return
	}
	if invs == nil {
		invs = []*database.Invitation{}
	}
	writeJSON(w, http.StatusOK, invs)
}

func (h *AdminHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	username := auth.GetSubject(r.Context())
	user, err := h.repo.GetUserByUsername(r.Context(), username)
	if err != nil || user == nil {
		slog.Error("admin user not found", "username", username)
		http.Error(w, `{"error":"admin user not found"}`, http.StatusInternalServerError)
		return
	}

	inv, err := h.repo.CreateInvitation(r.Context(), user.ID)
	if err != nil {
		slog.Error("create invitation", "error", err)
		http.Error(w, `{"error":"failed to create invitation"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, inv)
}

func (h *AdminHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req adminCreateProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.TenantID <= 0 {
		http.Error(w, `{"error":"tenant_id is required"}`, http.StatusBadRequest)
		return
	}

	if req.Slug == "" {
		http.Error(w, `{"error":"slug is required"}`, http.StatusBadRequest)
		return
	}

	if req.DefaultQuality <= 0 { req.DefaultQuality = 80 }
	if req.DefaultFormat == "" { req.DefaultFormat = "webp" }
	if req.MaxWidth <= 0 { req.MaxWidth = 4096 }
	if req.MaxHeight <= 0 { req.MaxHeight = 4096 }
	if req.RPSSec <= 0 { req.RPSSec = 100 }
	if req.Burst <= 0 { req.Burst = 200 }

	project, err := h.repo.CreateProject(r.Context(), req.TenantID, req.Slug, req.Name, req.BasePath, req.Provider, req.Bucket, req.Region, req.AccessKeyID, req.SecretAccessKey, req.CredentialsJSON, req.Endpoint, req.DefaultQuality, req.DefaultFormat, req.MaxWidth, req.MaxHeight, req.RPSSec, req.Burst)
	if err != nil {
		slog.Error("admin create project", "error", err)
		http.Error(w, `{"error":"failed to create project"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}
