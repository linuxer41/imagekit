package tenant

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/storage"
)

type ProjectCache struct {
	mu       sync.RWMutex
	projects map[string]*CachedProject
	repo     *database.Repo
	stopCh   chan struct{}
}

type CachedProject struct {
	*database.Project
	Provider storage.Provider
}

func NewProjectCache(repo *database.Repo) *ProjectCache {
	return &ProjectCache{
		projects: make(map[string]*CachedProject),
		repo:     repo,
		stopCh:   make(chan struct{}),
	}
}

func (c *ProjectCache) Start(ctx context.Context) {
	if err := c.Refresh(ctx); err != nil {
		slog.Warn("initial tenant cache refresh", "error", err)
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.Refresh(ctx); err != nil {
					slog.Error("refresh tenant cache", "error", err)
				}
			case <-c.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	slog.Info("project cache started, refresh every 30s")
}

func (c *ProjectCache) Stop() {
	close(c.stopCh)
}

func (c *ProjectCache) Refresh(ctx context.Context) error {
	projects, err := c.repo.ListAllActiveProjects(ctx)
	if err != nil {
		return err
	}

	projectMap := make(map[string]*CachedProject, len(projects))

	for _, p := range projects {
		prov, err := storage.NewProvider(p)
		if err != nil {
			slog.Error("creating storage provider for project", "slug", p.Slug, "error", err)
			continue
		}

		projectMap[p.Slug] = &CachedProject{
			Project:  p,
			Provider: prov,
		}
	}

	c.mu.Lock()
	c.projects = projectMap
	c.mu.Unlock()

	slog.Info("project cache refreshed", "count", len(projectMap))
	return nil
}

func (c *ProjectCache) Get(slug string) (*CachedProject, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.projects[slug]
	return p, ok
}

func (c *ProjectCache) Set(cp *CachedProject) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.projects[cp.Slug] = cp
}
