package service

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	domain_storage "github.com/Doratou/Backend/internal/domain/repository/storage"
	"github.com/google/uuid"
)

type AttachmentServiceConfig struct {
	PublicURLPrefix     string
	DefaultPresignedTTL time.Duration
}

type attachmentService struct {
	attachmentRepo domain_db.AttachmentRepository
	storageRepo    domain_storage.StorageRepository
	cfg            AttachmentServiceConfig
	consts         *bootstrap.Constant
}

func NewAttachmentService(
	attachmentRepo domain_db.AttachmentRepository,
	storageRepo domain_storage.StorageRepository,
	cfg AttachmentServiceConfig,
	consts *bootstrap.Constant,
) adapter.AttachmentService {
	if cfg.DefaultPresignedTTL <= 0 {
		defaultTTL := consts.Attachment.DefaultPresignedTTL
		log.Printf("[CONFIG NOTICE] DefaultPresignedTTL not configured or non-positive (%v), applying default: %v", cfg.DefaultPresignedTTL, defaultTTL)
		cfg.DefaultPresignedTTL = defaultTTL
	}

	return &attachmentService{
		attachmentRepo: attachmentRepo,
		storageRepo:    storageRepo,
		cfg:            cfg,
		consts:         consts,
	}
}

// Upload handles receiving a multipart file, validating, uploading to storage, and storing in DB.
func (s *attachmentService) Upload(ctx context.Context, userID int64, fileHeader *multipart.FileHeader, req dto.UploadAttachmentRequest) (*dto.AttachmentResponse, error) {
	if fileHeader == nil {
		return nil, exception.Wrap(exception.ErrBadRequest, fmt.Errorf("file is required"))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, exception.Wrap(exception.ErrBadRequest, fmt.Errorf("unable to read file: %w", err))
	}
	defer file.Close()

	// 1. Detect actual MIME type from magic bytes
	headerBytes := make([]byte, 512)
	n, err := file.Read(headerBytes)
	if err != nil && err != io.EOF {
		return nil, exception.Wrap(exception.ErrBadRequest, fmt.Errorf("failed to read file header: %w", err))
	}
	detectedMime := http.DetectContentType(headerBytes[:n])

	// 2. Validate MIME type and file size against limits
	if err := s.validateFile(req.Type, fileHeader.Size, detectedMime); err != nil {
		return nil, err
	}

	// 3. Compute SHA-256 hash of file content
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// 4. Extract image dimensions if applicable
	var width, height *int
	if strings.HasPrefix(detectedMime, "image/") {
		if _, err := file.Seek(0, io.SeekStart); err == nil {
			if cfg, _, err := image.DecodeConfig(file); err == nil {
				w := cfg.Width
				h := cfg.Height
				width = &w
				height = &h
			}
		}
	}

	// 5. Reset seek for upload
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}

	// 6. Determine public status and generate safe object key
	isPublic := s.resolveIsPublic(req.Type, req.IsPublic)
	objectKey := s.generateObjectKey(req.Type, fileHeader.Filename)

	// 7. Upload to RustFS / S3 storage
	if _, err := s.storageRepo.UploadFile(ctx, objectKey, file, fileHeader.Size, detectedMime); err != nil {
		return nil, exception.Wrap(exception.ErrStorageUnavailable, err)
	}

	// 8. Persist metadata in Postgres database
	now := time.Now().UTC()
	attachment := &entity.Attachment{
		UploadedByUserID: &userID,
		GroupID:          req.GroupID,
		ExpenseID:        req.ExpenseID,
		Type:             req.Type,
		OriginalFilename: &fileHeader.Filename,
		ObjectKey:        &objectKey,
		MimeType:         &detectedMime,
		FileSizeBytes:    &fileHeader.Size,
		FileHash:         &fileHash,
		Width:            width,
		Height:           height,
		IsPublic:         isPublic,
		UploadedAt:       &now,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		// Rollback file in storage to prevent orphaned files
		_ = s.storageRepo.DeleteFile(context.Background(), objectKey)
		return nil, err
	}

	accessURL, _ := s.buildAccessURL(ctx, objectKey, isPublic)
	return dto.ToAttachmentResponse(attachment, accessURL), nil
}

// UploadBatch handles uploading multiple multipart files with partial-success tolerance.
func (s *attachmentService) UploadBatch(
	ctx context.Context,
	userID int64,
	files []*multipart.FileHeader,
	defaultType enum.AttachmentType,
	groupID, expenseID *int64,
	isPublic *bool,
) (*dto.BatchUploadResponse, error) {
	if len(files) == 0 {
		return nil, exception.Wrap(exception.ErrBadRequest, fmt.Errorf("no files provided for batch upload"))
	}

	var (
		validEntities  []*entity.Attachment
		uploadedKeys   []string
		failures       []dto.BatchUploadItemFailure
		resolvedPublic = s.resolveIsPublic(defaultType, isPublic)
	)

	for _, fileHeader := range files {
		if fileHeader == nil {
			continue
		}

		file, err := fileHeader.Open()
		if err != nil {
			failures = append(failures, dto.BatchUploadItemFailure{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("cannot open file: %v", err),
			})
			continue
		}

		headerBytes := make([]byte, 512)
		n, err := file.Read(headerBytes)
		if err != nil && err != io.EOF {
			file.Close()
			failures = append(failures, dto.BatchUploadItemFailure{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("failed to read file header: %v", err),
			})
			continue
		}
		detectedMime := http.DetectContentType(headerBytes[:n])

		if err := s.validateFile(defaultType, fileHeader.Size, detectedMime); err != nil {
			file.Close()
			failures = append(failures, dto.BatchUploadItemFailure{
				Filename: fileHeader.Filename,
				Error:    err.Error(),
			})
			continue
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			file.Close()
			failures = append(failures, dto.BatchUploadItemFailure{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("seek error: %v", err),
			})
			continue
		}
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			file.Close()
			failures = append(failures, dto.BatchUploadItemFailure{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("hash error: %v", err),
			})
			continue
		}
		fileHash := hex.EncodeToString(hasher.Sum(nil))

		var width, height *int
		if strings.HasPrefix(detectedMime, "image/") {
			if _, err := file.Seek(0, io.SeekStart); err == nil {
				if cfg, _, err := image.DecodeConfig(file); err == nil {
					w := cfg.Width
					h := cfg.Height
					width = &w
					height = &h
				}
			}
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			file.Close()
			failures = append(failures, dto.BatchUploadItemFailure{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("seek before upload error: %v", err),
			})
			continue
		}

		objectKey := s.generateObjectKey(defaultType, fileHeader.Filename)

		if _, err := s.storageRepo.UploadFile(ctx, objectKey, file, fileHeader.Size, detectedMime); err != nil {
			file.Close()
			failures = append(failures, dto.BatchUploadItemFailure{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("storage upload error: %v", err),
			})
			continue
		}
		file.Close()
		uploadedKeys = append(uploadedKeys, objectKey)

		now := time.Now().UTC()
		validEntities = append(validEntities, &entity.Attachment{
			UploadedByUserID: &userID,
			GroupID:          groupID,
			ExpenseID:        expenseID,
			Type:             defaultType,
			OriginalFilename: &fileHeader.Filename,
			ObjectKey:        &objectKey,
			MimeType:         &detectedMime,
			FileSizeBytes:    &fileHeader.Size,
			FileHash:         &fileHash,
			Width:            width,
			Height:           height,
			IsPublic:         resolvedPublic,
			UploadedAt:       &now,
		})
	}

	if len(validEntities) > 0 {
		if err := s.attachmentRepo.CreateBatch(ctx, validEntities); err != nil {
			_ = s.storageRepo.DeleteFilesBatch(context.Background(), uploadedKeys)
			return nil, err
		}
	}

	items := make([]*dto.AttachmentResponse, len(validEntities))
	for i, ent := range validEntities {
		url, _ := s.buildAccessURL(ctx, *ent.ObjectKey, ent.IsPublic)
		items[i] = dto.ToAttachmentResponse(ent, url)
	}

	return &dto.BatchUploadResponse{
		SuccessCount: len(items),
		FailureCount: len(failures),
		Items:        items,
		Failures:     failures,
	}, nil
}

// CreatePresignedUpload generates a presigned PUT URL allowing clients to upload directly to RustFS.
func (s *attachmentService) CreatePresignedUpload(ctx context.Context, userID int64, req dto.PresignedUploadRequest) (*dto.PresignedUploadResponse, error) {
	if err := s.validateFile(req.Type, req.FileSizeBytes, req.MimeType); err != nil {
		return nil, err
	}

	objectKey := s.generateObjectKey(req.Type, req.Filename)
	uploadURL, err := s.storageRepo.GetPresignedUploadURL(ctx, objectKey, s.cfg.DefaultPresignedTTL)
	if err != nil {
		return nil, exception.Wrap(exception.ErrStorageUnavailable, err)
	}

	return &dto.PresignedUploadResponse{
		UploadURL:        uploadURL,
		ObjectKey:        objectKey,
		ExpiresInSeconds: int64(s.cfg.DefaultPresignedTTL.Seconds()),
	}, nil
}

// CreatePresignedUploadBatch generates direct upload PUT URLs for multiple items.
func (s *attachmentService) CreatePresignedUploadBatch(ctx context.Context, userID int64, req dto.BatchPresignedUploadRequest) (*dto.BatchPresignedUploadResponse, error) {
	items := make([]dto.BatchPresignedUploadItemResponse, 0, len(req.Items))

	for _, item := range req.Items {
		if err := s.validateFile(item.Type, item.FileSizeBytes, item.MimeType); err != nil {
			return nil, err
		}

		objectKey := s.generateObjectKey(item.Type, item.Filename)
		uploadURL, err := s.storageRepo.GetPresignedUploadURL(ctx, objectKey, s.cfg.DefaultPresignedTTL)
		if err != nil {
			return nil, exception.Wrap(exception.ErrStorageUnavailable, err)
		}

		items = append(items, dto.BatchPresignedUploadItemResponse{
			Filename:         item.Filename,
			UploadURL:        uploadURL,
			ObjectKey:        objectKey,
			ExpiresInSeconds: int64(s.cfg.DefaultPresignedTTL.Seconds()),
		})
	}

	return &dto.BatchPresignedUploadResponse{
		Items: items,
	}, nil
}

// ConfirmPresignedUpload registers the metadata in Postgres after client completes direct upload to RustFS.
func (s *attachmentService) ConfirmPresignedUpload(ctx context.Context, userID int64, req dto.ConfirmPresignedUploadRequest) (*dto.AttachmentResponse, error) {
	stat, err := s.storageRepo.StatObject(ctx, req.ObjectKey)
	if err != nil {
		return nil, exception.Wrap(exception.ErrAttachmentNotFound, fmt.Errorf("uploaded file not found in storage: %w", err))
	}

	isPublic := s.resolveIsPublic(req.Type, req.IsPublic)
	now := time.Now().UTC()
	fileSize := stat.Size
	mimeType := stat.ContentType

	attachment := &entity.Attachment{
		UploadedByUserID: &userID,
		GroupID:          req.GroupID,
		ExpenseID:        req.ExpenseID,
		Type:             req.Type,
		OriginalFilename: &req.OriginalFilename,
		ObjectKey:        &req.ObjectKey,
		MimeType:         &mimeType,
		FileSizeBytes:    &fileSize,
		FileHash:         &stat.ETag,
		IsPublic:         isPublic,
		UploadedAt:       &now,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		return nil, err
	}

	accessURL, _ := s.buildAccessURL(ctx, req.ObjectKey, isPublic)
	return dto.ToAttachmentResponse(attachment, accessURL), nil
}

// ConfirmPresignedUploadBatch confirms and creates database records for multiple uploaded items.
func (s *attachmentService) ConfirmPresignedUploadBatch(ctx context.Context, userID int64, req dto.BatchConfirmUploadRequest) (*dto.BatchUploadResponse, error) {
	var (
		validEntities []*entity.Attachment
		failures      []dto.BatchUploadItemFailure
	)

	now := time.Now().UTC()
	for _, item := range req.Items {
		stat, err := s.storageRepo.StatObject(ctx, item.ObjectKey)
		if err != nil {
			failures = append(failures, dto.BatchUploadItemFailure{
				Filename: item.OriginalFilename,
				Error:    fmt.Sprintf("object %q not found in storage: %v", item.ObjectKey, err),
			})
			continue
		}

		isPublic := s.resolveIsPublic(item.Type, item.IsPublic)
		size := stat.Size
		mime := stat.ContentType

		validEntities = append(validEntities, &entity.Attachment{
			UploadedByUserID: &userID,
			GroupID:          item.GroupID,
			ExpenseID:        item.ExpenseID,
			Type:             item.Type,
			OriginalFilename: &item.OriginalFilename,
			ObjectKey:        &item.ObjectKey,
			MimeType:         &mime,
			FileSizeBytes:    &size,
			FileHash:         &stat.ETag,
			IsPublic:         isPublic,
			UploadedAt:       &now,
		})
	}

	if len(validEntities) > 0 {
		if err := s.attachmentRepo.CreateBatch(ctx, validEntities); err != nil {
			return nil, err
		}
	}

	items := make([]*dto.AttachmentResponse, len(validEntities))
	for i, ent := range validEntities {
		url, _ := s.buildAccessURL(ctx, *ent.ObjectKey, ent.IsPublic)
		items[i] = dto.ToAttachmentResponse(ent, url)
	}

	return &dto.BatchUploadResponse{
		SuccessCount: len(items),
		FailureCount: len(failures),
		Items:        items,
		Failures:     failures,
	}, nil
}

// GetByID retrieves attachment metadata and generates an access URL.
func (s *attachmentService) GetByID(ctx context.Context, id int64) (*dto.AttachmentResponse, error) {
	attachment, err := s.attachmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	accessURL, _ := s.buildAccessURL(ctx, *attachment.ObjectKey, attachment.IsPublic)
	return dto.ToAttachmentResponse(attachment, accessURL), nil
}

// GetByObjectKey retrieves attachment metadata by object storage key.
func (s *attachmentService) GetByObjectKey(ctx context.Context, objectKey string) (*dto.AttachmentResponse, error) {
	attachment, err := s.attachmentRepo.FindByObjectKey(ctx, objectKey)
	if err != nil {
		return nil, err
	}

	accessURL, _ := s.buildAccessURL(ctx, *attachment.ObjectKey, attachment.IsPublic)
	return dto.ToAttachmentResponse(attachment, accessURL), nil
}

// List retrieves paginated attachments matching filter criteria.
func (s *attachmentService) List(ctx context.Context, filter dto.AttachmentFilterRequest) (*dto.AttachmentListResponse, error) {
	repoFilter := domain_db.AttachmentFilter{
		UploadedByUserID: filter.UploadedByUserID,
		GroupID:          filter.GroupID,
		ExpenseID:        filter.ExpenseID,
		Type:             filter.Type,
		IsPublic:         filter.IsPublic,
		Search:           filter.Search,
	}

	repoPagination := domain_db.Pagination{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	items, total, err := s.attachmentRepo.List(ctx, repoFilter, repoPagination)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AttachmentResponse, len(items))
	for i, item := range items {
		accessURL, _ := s.buildAccessURL(ctx, *item.ObjectKey, item.IsPublic)
		responses[i] = dto.ToAttachmentResponse(item, accessURL)
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.AttachmentListResponse{
		Items:      responses,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetBatchURLs generates active access URLs for multiple attachment IDs.
func (s *attachmentService) GetBatchURLs(ctx context.Context, ids []int64, expiry time.Duration) (*dto.BatchURLsResponse, error) {
	if len(ids) == 0 {
		return &dto.BatchURLsResponse{Items: []dto.BatchURLItem{}}, nil
	}

	attachments, err := s.attachmentRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	if expiry <= 0 {
		expiry = s.cfg.DefaultPresignedTTL
	}

	var privateKeys []string
	for _, att := range attachments {
		if !att.IsPublic && att.ObjectKey != nil {
			privateKeys = append(privateKeys, *att.ObjectKey)
		}
	}

	var presignedMap map[string]string
	if len(privateKeys) > 0 {
		presignedMap, _ = s.storageRepo.GetPresignedURLsBatch(ctx, privateKeys, expiry)
	}

	items := make([]dto.BatchURLItem, 0, len(attachments))
	for _, att := range attachments {
		if att.ObjectKey == nil {
			continue
		}
		var accessURL string
		if att.IsPublic {
			prefix := strings.TrimRight(s.cfg.PublicURLPrefix, "/")
			accessURL = fmt.Sprintf("%s/%s", prefix, *att.ObjectKey)
		} else if presignedMap != nil {
			accessURL = presignedMap[*att.ObjectKey]
		}

		items = append(items, dto.BatchURLItem{
			ID:        att.ID,
			AccessURL: accessURL,
		})
	}

	return &dto.BatchURLsResponse{Items: items}, nil
}

// DownloadZipStream streams an on-the-fly zip archive containing the requested attachments.
func (s *attachmentService) DownloadZipStream(ctx context.Context, ids []int64, w io.Writer) error {
	if len(ids) == 0 {
		return exception.Wrap(exception.ErrBadRequest, fmt.Errorf("no ids provided for zip download"))
	}

	attachments, err := s.attachmentRepo.FindByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if len(attachments) == 0 {
		return exception.ErrAttachmentNotFound
	}

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	usedNames := make(map[string]int)

	for _, att := range attachments {
		if att.ObjectKey == nil {
			continue
		}

		rc, err := s.storageRepo.DownloadFile(ctx, *att.ObjectKey)
		if err != nil {
			log.Printf("[ZIP DOWNLOAD WARNING] skipping object %q: %v", *att.ObjectKey, err)
			continue
		}

		baseName := "file"
		if att.OriginalFilename != nil && *att.OriginalFilename != "" {
			baseName = *att.OriginalFilename
		}

		cleanName := baseName
		count := usedNames[baseName]
		if count > 0 {
			ext := filepath.Ext(baseName)
			raw := strings.TrimSuffix(baseName, ext)
			cleanName = fmt.Sprintf("%s_%d%s", raw, count, ext)
		}
		usedNames[baseName]++

		entryHeader := &zip.FileHeader{
			Name:     cleanName,
			Method:   zip.Deflate,
			Modified: time.Now().UTC(),
		}
		if att.FileSizeBytes != nil {
			entryHeader.UncompressedSize64 = uint64(*att.FileSizeBytes)
		}

		entryWriter, err := zipWriter.CreateHeader(entryHeader)
		if err != nil {
			rc.Close()
			return exception.Wrap(exception.ErrInternal, err)
		}

		if _, err := io.Copy(entryWriter, rc); err != nil {
			rc.Close()
			return exception.Wrap(exception.ErrInternal, err)
		}
		rc.Close()
	}

	return zipWriter.Flush()
}

// Update modifies metadata fields of an existing attachment.
func (s *attachmentService) Update(ctx context.Context, userID int64, id int64, req dto.UpdateAttachmentRequest) (*dto.AttachmentResponse, error) {
	attachment, err := s.attachmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ownership check
	if attachment.UploadedByUserID != nil && *attachment.UploadedByUserID != userID {
		return nil, exception.ErrAttachmentForbidden
	}

	if req.OriginalFilename != nil {
		attachment.OriginalFilename = req.OriginalFilename
	}
	if req.GroupID != nil {
		attachment.GroupID = req.GroupID
	}
	if req.ExpenseID != nil {
		attachment.ExpenseID = req.ExpenseID
	}
	if req.Type != nil {
		attachment.Type = *req.Type
	}
	if req.IsPublic != nil {
		attachment.IsPublic = *req.IsPublic
	}

	if err := s.attachmentRepo.Update(ctx, attachment); err != nil {
		return nil, err
	}

	accessURL, _ := s.buildAccessURL(ctx, *attachment.ObjectKey, attachment.IsPublic)
	return dto.ToAttachmentResponse(attachment, accessURL), nil
}

// UpdateBatch updates common metadata fields for multiple attachments owned by the user.
func (s *attachmentService) UpdateBatch(ctx context.Context, userID int64, req dto.BatchUpdateAttachmentRequest) error {
	if len(req.IDs) == 0 {
		return nil
	}

	attachments, err := s.attachmentRepo.FindByIDs(ctx, req.IDs)
	if err != nil {
		return err
	}
	if len(attachments) == 0 {
		return exception.ErrAttachmentNotFound
	}

	for _, att := range attachments {
		if att.UploadedByUserID != nil && *att.UploadedByUserID != userID {
			return exception.ErrAttachmentForbidden
		}
	}

	updates := make(map[string]any)
	if req.OriginalFilename != nil {
		updates["original_filename"] = *req.OriginalFilename
	}
	if req.GroupID != nil {
		updates["group_id"] = *req.GroupID
	}
	if req.ExpenseID != nil {
		updates["expense_id"] = *req.ExpenseID
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	return s.attachmentRepo.UpdateBatch(ctx, req.IDs, updates)
}

// Delete removes the attachment metadata, optionally hard-deleting the underlying storage object.
func (s *attachmentService) Delete(ctx context.Context, userID int64, id int64, hardDelete bool) error {
	attachment, err := s.attachmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Ownership check
	if attachment.UploadedByUserID != nil && *attachment.UploadedByUserID != userID {
		return exception.ErrAttachmentForbidden
	}

	if hardDelete {
		if attachment.ObjectKey != nil {
			_ = s.storageRepo.DeleteFile(ctx, *attachment.ObjectKey)
		}
		if attachment.ThumbnailObjectKey != nil {
			_ = s.storageRepo.DeleteFile(ctx, *attachment.ThumbnailObjectKey)
		}
		return s.attachmentRepo.HardDeleteByID(ctx, id)
	}

	return s.attachmentRepo.DeleteByID(ctx, id)
}

// DeleteBatch deletes multiple attachments (soft or hard).
func (s *attachmentService) DeleteBatch(ctx context.Context, userID int64, ids []int64, hardDelete bool) error {
	if len(ids) == 0 {
		return nil
	}

	attachments, err := s.attachmentRepo.FindByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if len(attachments) == 0 {
		return exception.ErrAttachmentNotFound
	}

	for _, att := range attachments {
		if att.UploadedByUserID != nil && *att.UploadedByUserID != userID {
			return exception.ErrAttachmentForbidden
		}
	}

	if hardDelete {
		var keys []string
		for _, att := range attachments {
			if att.ObjectKey != nil {
				keys = append(keys, *att.ObjectKey)
			}
			if att.ThumbnailObjectKey != nil {
				keys = append(keys, *att.ThumbnailObjectKey)
			}
		}
		if len(keys) > 0 {
			_ = s.storageRepo.DeleteFilesBatch(ctx, keys)
		}
		return s.attachmentRepo.HardDeleteByIDs(ctx, ids)
	}

	return s.attachmentRepo.DeleteByIDs(ctx, ids)
}

// GetFreshAccessURL generates an updated direct or presigned URL.
func (s *attachmentService) GetFreshAccessURL(ctx context.Context, id int64, expiry time.Duration) (string, error) {
	attachment, err := s.attachmentRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	if expiry <= 0 {
		expiry = s.cfg.DefaultPresignedTTL
	}

	if attachment.IsPublic {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.PublicURLPrefix, "/"), *attachment.ObjectKey), nil
	}

	return s.storageRepo.GetPresignedURL(ctx, *attachment.ObjectKey, expiry)
}

// DownloadStream streams the object from storage for direct HTTP proxy download.
func (s *attachmentService) DownloadStream(ctx context.Context, id int64) (io.ReadCloser, *dto.FileStreamMeta, error) {
	attachment, err := s.attachmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	rc, err := s.storageRepo.DownloadFile(ctx, *attachment.ObjectKey)
	if err != nil {
		return nil, nil, exception.Wrap(exception.ErrStorageUnavailable, err)
	}

	filename := "file"
	if attachment.OriginalFilename != nil && *attachment.OriginalFilename != "" {
		filename = *attachment.OriginalFilename
	}

	contentType := "application/octet-stream"
	if attachment.MimeType != nil && *attachment.MimeType != "" {
		contentType = *attachment.MimeType
	}

	var size int64
	if attachment.FileSizeBytes != nil {
		size = *attachment.FileSizeBytes
	}

	meta := &dto.FileStreamMeta{
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
	}

	return rc, meta, nil
}

// --- Internal Helper Functions ---

func (s *attachmentService) validateFile(attType enum.AttachmentType, size int64, mimeType string) error {
	var maxAllowed int64
	switch attType {
	case enum.AttachmentTypeProfileImage:
		maxAllowed = s.consts.Attachment.MaxProfileImageSize
		if !strings.HasPrefix(mimeType, "image/") {
			return exception.ErrInvalidFileType
		}
	case enum.AttachmentTypeGroupCover:
		maxAllowed = s.consts.Attachment.MaxGroupCoverSize
		if !strings.HasPrefix(mimeType, "image/") {
			return exception.ErrInvalidFileType
		}
	case enum.AttachmentTypeGalleryImage:
		maxAllowed = s.consts.Attachment.MaxGalleryImageSize
		if !strings.HasPrefix(mimeType, "image/") {
			return exception.ErrInvalidFileType
		}
	case enum.AttachmentTypeGalleryVideo:
		maxAllowed = s.consts.Attachment.MaxGalleryVideoSize
		if !strings.HasPrefix(mimeType, "video/") {
			return exception.ErrInvalidFileType
		}
	case enum.AttachmentTypeExpenseReceipt, enum.AttachmentTypeSettlementReceipt:
		maxAllowed = s.consts.Attachment.MaxReceiptSize
		if !strings.HasPrefix(mimeType, "image/") && mimeType != "application/pdf" {
			return exception.ErrInvalidFileType
		}
	default:
		return exception.ErrInvalidFileType
	}

	if maxAllowed > 0 && size > maxAllowed {
		return exception.ErrFileTooLarge
	}

	return nil
}

func (s *attachmentService) resolveIsPublic(attType enum.AttachmentType, isPublicReq *bool) bool {
	if isPublicReq != nil {
		return *isPublicReq
	}
	// Profile images and group covers are public by default
	if attType == enum.AttachmentTypeProfileImage || attType == enum.AttachmentTypeGroupCover {
		return true
	}
	return false
}

func (s *attachmentService) generateObjectKey(attType enum.AttachmentType, filename string) string {
	now := time.Now().UTC()
	cleanExt := strings.ToLower(filepath.Ext(filename))
	cleanBase := strings.TrimSuffix(filepath.Base(filename), cleanExt)
	cleanBase = strings.ReplaceAll(cleanBase, " ", "_")
	if len(cleanBase) > 30 {
		cleanBase = cleanBase[:30]
	}
	uniqueID := uuid.New().String()

	return fmt.Sprintf("attachments/%s/%04d/%02d/%s_%s%s",
		strings.ToLower(string(attType)),
		now.Year(),
		now.Month(),
		uniqueID,
		cleanBase,
		cleanExt,
	)
}

func (s *attachmentService) buildAccessURL(ctx context.Context, objectKey string, isPublic bool) (string, error) {
	if isPublic {
		prefix := strings.TrimRight(s.cfg.PublicURLPrefix, "/")
		return fmt.Sprintf("%s/%s", prefix, objectKey), nil
	}
	return s.storageRepo.GetPresignedURL(ctx, objectKey, s.cfg.DefaultPresignedTTL)
}
