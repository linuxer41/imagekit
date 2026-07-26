package router

import (
	"fmt"
	"net/http"

	"github.com/vendemas/imagekit/internal/auth"
	"github.com/vendemas/imagekit/internal/database"
)

func adminAdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.GetRole(r.Context()) != "admin" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type healthHandler struct {
	repo *database.Repo
}

func newHealthHandler(repo *database.Repo) *healthHandler {
	return &healthHandler{repo: repo}
}

func (h *healthHandler) Deep(w http.ResponseWriter, r *http.Request) {
	dbOK := true
	if err := h.repo.Ping(r.Context()); err != nil {
		dbOK = false
	}

	status := http.StatusOK
	if !dbOK {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"status":%d,"database":%v,"version":"1.0.0"}`, status, dbOK)
}

var openAPIHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	spec := `{
  "openapi": "3.0.3",
  "info": { "title": "ImageKit", "version": "1.0.0", "description": "Multi-tenant ImageKit" },
  "servers": [ { "url": "https://images.vendemas.icu", "description": "Production" } ],
  "paths": {
    "/auth/register": { "post": { "tags": ["Auth"], "summary": "Register tenant", "requestBody": { "content": { "application/json": { "schema": { "type": "object", "properties": { "email": {"type":"string"}, "password": {"type":"string"}, "provider":{"type":"string","enum":["gcs","s3","rustfs"]}, "bucket":{"type":"string"} } } } } }, "responses": { "201": { "description": "Created" } } } },
    "/auth/login": { "post": { "tags": ["Auth"], "summary": "Login tenant", "requestBody": { "content": { "application/json": { "schema": { "type": "object", "properties": { "email": {"type":"string"}, "password": {"type":"string"} } } } } }, "responses": { "200": { "description": "Token" } } } },
    "/api/projects": { "get": { "tags": ["Projects"], "summary": "List projects", "security": [{"BearerAuth":[]}], "responses": { "200": { "description": "Project list" } } }, "post": { "tags": ["Projects"], "summary": "Create project", "security": [{"BearerAuth":[]}], "requestBody": { "content": { "application/json": { "schema": { "type": "object", "properties": { "slug":{"type":"string"}, "name":{"type":"string"} } } } } }, "responses": { "201": { "description": "Created" } } } }
  },
  "components": { "securitySchemes": { "BearerAuth": { "type": "http", "scheme": "bearer" } } }
}`
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(spec))
})
