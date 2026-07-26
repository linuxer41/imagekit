package storage

import (
	"context"
	"fmt"

	"github.com/vendemas/imagekit/internal/database"
)

type Provider interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	TestConnection(ctx context.Context) error
}

func NewProvider(project *database.Project) (Provider, error) {
	switch project.Provider {
	case "gcs":
		return NewGCS(project)
	case "s3":
		return NewS3(project, false)
	case "rustfs":
		return NewS3(project, true)
	default:
		return nil, fmt.Errorf("unknown provider: %s", project.Provider)
	}
}
