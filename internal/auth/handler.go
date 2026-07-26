package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/vendemas/imagekit/internal/database"
)

type Handler struct {
	repo *database.Repo
	jwt  *JWT
}

func NewHandler(repo *database.Repo, jwt *JWT) *Handler {
	return &Handler{repo: repo, jwt: jwt}
}

type registerReq struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	InvitationCode  string `json:"invitation_code"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResp struct {
	Token    string `json:"token"`
	TenantID int    `json:"tenant_id"`
}

type userLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.InvitationCode == "" {
		http.Error(w, `{"error":"name, email, password, and invitation_code are required"}`, http.StatusBadRequest)
		return
	}

	tenant, err := h.repo.RegisterTenant(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		slog.Error("register tenant", "error", err)
		http.Error(w, `{"error":"registration failed, email may already exist"}`, http.StatusConflict)
		return
	}

	if err := h.repo.ValidateAndUseInvitation(r.Context(), req.InvitationCode, tenant.ID); err != nil {
		// rollback tenant creation
		slog.Error("invalid invitation code", "code", req.InvitationCode, "error", err)
		http.Error(w, `{"error":"invalid or used invitation code"}`, http.StatusBadRequest)
		return
	}

	token, err := h.jwt.Generate(strconv.Itoa(tenant.ID), "tenant", 72*time.Hour)
	if err != nil {
		slog.Error("generate token", "error", err)
		http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, tokenResp{Token: token, TenantID: tenant.ID})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	tenant, err := h.repo.LoginTenant(r.Context(), req.Email, req.Password)
	if err != nil {
		slog.Error("login tenant", "error", err)
		http.Error(w, `{"error":"login failed"}`, http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.jwt.Generate(strconv.Itoa(tenant.ID), "tenant", 72*time.Hour)
	if err != nil {
		slog.Error("generate token", "error", err)
		http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tokenResp{Token: token, TenantID: tenant.ID})
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req userLoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	user, err := h.repo.LoginUser(r.Context(), req.Username, req.Password)
	if err != nil || user == nil {
		http.Error(w, `{"error":"invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.jwt.Generate(user.Username, user.Role, 72*time.Hour)
	if err != nil {
		slog.Error("generate token", "error", err)
		http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token, "role": user.Role})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.jwt.Generate(claims.Subject, claims.Role, 72*time.Hour)
	if err != nil {
		slog.Error("refresh token", "error", err)
		http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
