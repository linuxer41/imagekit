package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/vendemas/imagekit/internal/database"
)

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3(project *database.Project, customEndpoint bool) (*S3Client, error) {
	region := project.Region
	if region == "" || customEndpoint {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("s3 config: %w", err)
	}

	if project.AccessKeyID != "" && project.SecretAccessKey != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			project.AccessKeyID,
			project.SecretAccessKey,
			"",
		)
	}

	var client *s3.Client
	if customEndpoint && project.Endpoint != "" {
		ep := project.Endpoint
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &ep
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}

	return &S3Client{
		client: client,
		bucket: project.Bucket,
	}, nil
}

func (s *S3Client) Get(ctx context.Context, objectPath string) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectPath),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get %s: %w", objectPath, err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("s3 read all: %w", err)
	}

	return data, nil
}

func (s *S3Client) Delete(ctx context.Context, objectPath string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectPath),
	})
	if err != nil {
		return fmt.Errorf("s3 delete %s: %w", objectPath, err)
	}
	return nil
}

func (s *S3Client) TestConnection(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("s3 test connection: %w", err)
	}
	return nil
}
