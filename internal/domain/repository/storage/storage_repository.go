package storage

import (
	"context"
	"io"
	"time"
)

// FileUploadPayload represents a single file item in a batch upload operation.
type FileUploadPayload struct {
	ObjectKey   string
	Reader      io.Reader
	Size        int64
	ContentType string
}

// ObjectStat represents metadata about an object in storage.
type ObjectStat struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

// StorageRepository abstracts object storage (RustFS / MinIO / S3) operations.
// All methods use standard Go types (io.Reader, io.ReadCloser, primitives).
// No implementation-specific types (e.g. minio-go) leak outside the infrastructure layer.
type StorageRepository interface {
	// --- Single-file operations ---

	// UploadFile puts an object into storage and returns the stored objectKey.
	UploadFile(ctx context.Context, objectKey string, reader io.Reader, objectSize int64, contentType string) (string, error)

	// DownloadFile retrieves a stream of the object. Callers MUST close the returned ReadCloser.
	DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error)

	// StatObject fetches metadata for an object without downloading its payload.
	StatObject(ctx context.Context, objectKey string) (*ObjectStat, error)

	// GetPresignedURL generates a temporary presigned GET URL for public or authenticated downloads.
	GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)

	// GetPresignedUploadURL generates a temporary presigned PUT URL allowing clients to upload directly.
	GetPresignedUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)

	// DeleteFile removes a single object from storage.
	DeleteFile(ctx context.Context, objectKey string) error

	// --- Batch operations (Bounded concurrency via errgroup) ---

	// UploadFilesBatch uploads multiple files concurrently with bounded concurrency.
	UploadFilesBatch(ctx context.Context, files []FileUploadPayload) (map[string]string, error)

	// DownloadFilesBatch downloads multiple files concurrently. Callers MUST Close() each returned ReadCloser.
	DownloadFilesBatch(ctx context.Context, objectKeys []string) (map[string]io.ReadCloser, error)

	// GetPresignedURLsBatch generates presigned URLs for multiple keys concurrently.
	GetPresignedURLsBatch(ctx context.Context, objectKeys []string, expiry time.Duration) (map[string]string, error)

	// DeleteFilesBatch bulk-deletes multiple objects using native server-side batch removal.
	DeleteFilesBatch(ctx context.Context, objectKeys []string) error

	// --- Bucket maintenance ---

	// EnsureBucketExists checks if the default bucket exists, creating it if necessary.
	EnsureBucketExists(ctx context.Context) error
}
