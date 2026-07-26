package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/vendemas/imagekit/internal/database"
)

type AccountHandler struct {
	repo *database.Repo
}

func NewAccountHandler(repo *database.Repo) *AccountHandler {
	return &AccountHandler{repo: repo}
}

type accountResp struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil || tenant == nil {
		http.Error(w, `{"error":"tenant not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, accountResp{
		ID:    tenant.ID,
		Name:  tenant.Name,
		Email: tenant.Email,
	})
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromContext(r)
	if tenantID <= 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateTenantInfo(r.Context(), tenantID, req.Name); err != nil {
		slog.Error("update account", "error", err)
		http.Error(w, `{"error":"failed to update account"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}


