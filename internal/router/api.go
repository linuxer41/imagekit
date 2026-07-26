package router

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vendemas/imagekit/internal/api"
	"github.com/vendemas/imagekit/internal/auth"
	"github.com/vendemas/imagekit/internal/cache"
	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/middleware"
	"github.com/vendemas/imagekit/internal/tenant"
)

func NewAPIRouter(
	jwt *auth.JWT,
	repo *database.Repo,
	corsMiddleware *middleware.CORS,
	projectCache *tenant.ProjectCache,
	lruCache *cache.LRUCache,
	panelDir string,
	adminDir string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(corsMiddleware.Handler)

	healthHandler := newHealthHandler(repo)
	r.Get("/health", healthHandler.Deep)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/openapi.json", openAPIHandler)

	authHandler := auth.NewHandler(repo, jwt)

	// Auth endpoints (no JWT required)
	r.Post("/api/panel/auth/register", authHandler.Register)
	r.Post("/api/panel/auth/login", authHandler.Login)
	r.Post("/api/admin/auth/login", authHandler.LoginUser)

	// Panel JSON API (JWT required, tenant role)
	panelAPI := chi.NewRouter()
	panelAPI.Use(auth.Middleware(jwt))

	panelAPI.Post("/auth/refresh", authHandler.Refresh)

	projectsHandler := api.NewProjectsHandler(repo)
	panelAPI.Get("/projects", projectsHandler.List)
	panelAPI.Post("/projects", projectsHandler.Create)
	panelAPI.Get("/projects/{id}", projectsHandler.Get)
	panelAPI.Put("/projects/{id}", projectsHandler.Update)
	panelAPI.Delete("/projects/{id}", projectsHandler.Delete)

	accountHandler := api.NewAccountHandler(repo)
	panelAPI.Get("/account", accountHandler.Get)
	panelAPI.Put("/account", accountHandler.Update)

	imagesHandler := api.NewImagesHandler(repo, projectCache, lruCache)
	panelAPI.Delete("/images/{slug}/*", imagesHandler.Delete)

	metricsHandler := api.NewMetricsHandler(repo)
	panelAPI.Get("/projects/{id}/metrics", metricsHandler.ProjectMetrics)
	panelAPI.Get("/projects/{id}/summary", metricsHandler.ProjectSummary)
	panelAPI.Get("/metrics", metricsHandler.TenantMetrics)

	r.Mount("/api/panel", panelAPI)

	// Admin JSON API (JWT required, admin role only)
	adminAPI := chi.NewRouter()
	adminAPI.Use(auth.Middleware(jwt))
	adminAPI.Use(adminAdminOnly)

	adminHandler := api.NewAdminHandler(repo)
	adminAPI.Get("/stats", adminHandler.Stats)
	adminAPI.Get("/tenants", adminHandler.ListTenants)
	adminAPI.Get("/tenants/{id}", adminHandler.GetTenant)
	adminAPI.Get("/projects", adminHandler.ListProjects)
	adminAPI.Get("/users", adminHandler.ListUsers)
	adminAPI.Post("/projects", adminHandler.CreateProject)
	adminAPI.Get("/invitations", adminHandler.ListInvitations)
	adminAPI.Post("/invitations", adminHandler.CreateInvitation)
	adminAPI.Get("/projects/{id}/metrics", metricsHandler.AdminProjectMetrics)
	adminAPI.Get("/projects/{id}/summary", metricsHandler.AdminProjectSummary)
	adminAPI.Get("/metrics", metricsHandler.AdminGlobalMetrics)

	r.Mount("/api/admin", adminAPI)

	// Static files — panel SvelteKit build at /
	if panelDir != "" {
		serveStatic(r, "/", panelDir)
	}

	// Static files — admin SvelteKit build at /admin
	if adminDir != "" {
		adminRouter := chi.NewRouter()
		serveStatic(adminRouter, "/", adminDir)
		r.Mount("/admin", adminRouter)
	}

	return r
}

func serveStatic(r chi.Router, prefix string, dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return
	}

	// SvelteKit adapter-static generates:
	//   /index.html         → root page
	//   /login/index.html   → login page
	//   etc.
	// We need to serve files and use SPA fallback for non-file routes.
	fsys := os.DirFS(dir)
	staticFS, err := fs.Sub(fsys, ".")
	if err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(staticFS))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Try existing file first
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		fullPath := filepath.Join(dir, path)
		if _, err := os.Stat(fullPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback — serve index.html
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}
