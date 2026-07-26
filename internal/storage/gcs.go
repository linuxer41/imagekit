package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/vendemas/imagekit/internal/database"
)

type GCS struct {
	client *storage.Client
	bucket string
}

func NewGCS(project *database.Project) (*GCS, error) {
	var opts []option.ClientOption
	if project.CredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(project.CredentialsJSON)))
	}

	client, err := storage.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("gcs client: %w", err)
	}

	return &GCS{
		client: client,
		bucket: project.Bucket,
	}, nil
}

func (g *GCS) Get(ctx context.Context, objectPath string) ([]byte, error) {
	rc, err := g.client.Bucket(g.bucket).Object(objectPath).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs read %s: %w", objectPath, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("gcs read all: %w", err)
	}

	return data, nil
}

func (g *GCS) Delete(ctx context.Context, objectPath string) error {
	if err := g.client.Bucket(g.bucket).Object(objectPath).Delete(ctx); err != nil {
		return fmt.Errorf("gcs delete %s: %w", objectPath, err)
	}
	return nil
}

func (g *GCS) TestConnection(ctx context.Context) error {
	_, err := g.client.Bucket(g.bucket).Attrs(ctx)
	if err != nil {
		return fmt.Errorf("gcs test connection: %w", err)
	}
	return nil
}
