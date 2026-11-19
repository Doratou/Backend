package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	domain_storage "github.com/Doratou/Backend/internal/domain/repository/storage"
	"github.com/minio/minio-go/v7"
	"golang.org/x/sync/errgroup"
)

// EnsureBucketExists checks if cfg.Bucket exists, creating it if missing.
func (r *rustFSRepository) EnsureBucketExists(ctx context.Context) error {
	exists, err := r.client.BucketExists(ctx, r.bucket)
	if err != nil {
		return fmt.Errorf("rustfs: check bucket %q: %w", r.bucket, err)
	}
	if !exists {
		if err := r.client.MakeBucket(ctx, r.bucket, minio.MakeBucketOptions{Region: r.region}); err != nil {
			return fmt.Errorf("rustfs: create bucket %q: %w", r.bucket, err)
		}
	}
	return nil
}

// UploadFile uploads a single file to RustFS and returns the object key.
func (r *rustFSRepository) UploadFile(ctx context.Context, objectKey string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	_, err := r.client.PutObject(ctx, r.bucket, objectKey, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("rustfs: upload %q: %w", objectKey, err)
	}
	return objectKey, nil
}

// DownloadFile opens a stream to the object in RustFS.
// Performs an immediate Stat check so callers don't encounter deferred errors.
func (r *rustFSRepository) DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	obj, err := r.client.GetObject(ctx, r.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("rustfs: download %q: %w", objectKey, err)
	}

	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		return nil, fmt.Errorf("rustfs: stat %q: %w", objectKey, err)
	}

	return obj, nil
}

// StatObject retrieves metadata about an object from RustFS.
func (r *rustFSRepository) StatObject(ctx context.Context, objectKey string) (*domain_storage.ObjectStat, error) {
	info, err := r.client.StatObject(ctx, r.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("rustfs: stat %q: %w", objectKey, err)
	}

	return &domain_storage.ObjectStat{
		Key:          info.Key,
		Size:         info.Size,
		ContentType:  info.ContentType,
		ETag:         info.ETag,
		LastModified: info.LastModified,
	}, nil
}

// GetPresignedURL generates a temporary presigned GET URL.
func (r *rustFSRepository) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presigned, err := r.client.PresignedGetObject(ctx, r.bucket, objectKey, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("rustfs: presign GET %q: %w", objectKey, err)
	}
	return presigned.String(), nil
}

// GetPresignedUploadURL generates a temporary presigned PUT URL allowing client direct upload.
func (r *rustFSRepository) GetPresignedUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presigned, err := r.client.PresignedPutObject(ctx, r.bucket, objectKey, expiry)
	if err != nil {
		return "", fmt.Errorf("rustfs: presign PUT %q: %w", objectKey, err)
	}
	return presigned.String(), nil
}

// DeleteFile deletes a single object from RustFS.
func (r *rustFSRepository) DeleteFile(ctx context.Context, objectKey string) error {
	if err := r.client.RemoveObject(ctx, r.bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("rustfs: delete %q: %w", objectKey, err)
	}
	return nil
}

// UploadFilesBatch concurrently uploads multiple files with bounded concurrency.
func (r *rustFSRepository) UploadFilesBatch(ctx context.Context, files []domain_storage.FileUploadPayload) (map[string]string, error) {
	results := make(map[string]string, len(files))
	var (
		mu       sync.Mutex
		firstErr error
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(r.maxConcurrency)

	for _, f := range files {
		f := f
		g.Go(func() error {
			key, err := r.UploadFile(gCtx, f.ObjectKey, f.Reader, f.Size, f.ContentType)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return nil
			}
			results[f.ObjectKey] = key
			return nil
		})
	}
	_ = g.Wait()

	return results, firstErr
}

// DownloadFilesBatch concurrently downloads multiple files.
// Callers must Close each non-nil ReadCloser returned.
func (r *rustFSRepository) DownloadFilesBatch(ctx context.Context, objectKeys []string) (map[string]io.ReadCloser, error) {
	results := make(map[string]io.ReadCloser, len(objectKeys))
	var (
		mu       sync.Mutex
		firstErr error
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(r.maxConcurrency)

	for _, key := range objectKeys {
		key := key
		g.Go(func() error {
			rc, err := r.DownloadFile(gCtx, key)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return nil
			}
			results[key] = rc
			return nil
		})
	}
	_ = g.Wait()

	return results, firstErr
}

// GetPresignedURLsBatch generates presigned URLs for multiple keys concurrently.
func (r *rustFSRepository) GetPresignedURLsBatch(ctx context.Context, objectKeys []string, expiry time.Duration) (map[string]string, error) {
	results := make(map[string]string, len(objectKeys))
	var (
		mu       sync.Mutex
		firstErr error
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(r.maxConcurrency)

	for _, key := range objectKeys {
		key := key
		g.Go(func() error {
			presignedURL, err := r.GetPresignedURL(gCtx, key, expiry)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return nil
			}
			results[key] = presignedURL
			return nil
		})
	}
	_ = g.Wait()

	return results, firstErr
}

// DeleteFilesBatch performs a native server-side bulk delete using streaming channels.
func (r *rustFSRepository) DeleteFilesBatch(ctx context.Context, objectKeys []string) error {
	if len(objectKeys) == 0 {
		return nil
	}

	objectsCh := make(chan minio.ObjectInfo, len(objectKeys))
	go func() {
		defer close(objectsCh)
		for _, key := range objectKeys {
			select {
			case <-ctx.Done():
				return
			case objectsCh <- minio.ObjectInfo{Key: key}:
			}
		}
	}()

	var failures []string
	for removeErr := range r.client.RemoveObjects(ctx, r.bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if removeErr.Err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", removeErr.ObjectName, removeErr.Err))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("rustfs: batch delete had %d failure(s): %s", len(failures), strings.Join(failures, "; "))
	}
	return nil
}
