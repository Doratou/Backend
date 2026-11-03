package storage

import (
	"context"
	"fmt"

	domain_storage "github.com/Doratou/Backend/internal/domain/repository/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config holds the RustFS / MinIO connection settings.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
	Bucket          string
	MaxConcurrency  int
}

type rustFSRepository struct {
	client         *minio.Client
	bucket         string
	region         string
	maxConcurrency int
}

// NewRustFSRepository connects to RustFS / MinIO, verifies connection and bucket existence,
// and returns the domain StorageRepository interface.
func NewRustFSRepository(ctx context.Context, cfg *Config) (domain_storage.StorageRepository, error) {
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	concurrency := cfg.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 8
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("rustfs: init client: %w", err)
	}

	repo := &rustFSRepository{
		client:         client,
		bucket:         cfg.Bucket,
		region:         cfg.Region,
		maxConcurrency: concurrency,
	}

	if err := repo.EnsureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("rustfs: check/create bucket %q: %w", cfg.Bucket, err)
	}

	return repo, nil
}
