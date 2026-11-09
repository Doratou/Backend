package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type AttachmentHandler struct {
	attachmentService adapter.AttachmentService
}

func NewAttachmentHandler(attachmentService adapter.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentService: attachmentService,
	}
}

// Upload handles POST /api/v1/attachments/upload (Protected)
func (h *AttachmentHandler) Upload(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.UploadAttachmentRequest
	if err := c.ShouldBind(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, fmt.Errorf("file multipart field is required: %w", err)))
		return
	}

	res, err := h.attachmentService.Upload(c.Request.Context(), userID, fileHeader, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusCreated, res)
}

// CreatePresignedUpload handles POST /api/v1/attachments/presigned-upload (Protected)
func (h *AttachmentHandler) CreatePresignedUpload(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.PresignedUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.attachmentService.CreatePresignedUpload(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// ConfirmPresignedUpload handles POST /api/v1/attachments/confirm-upload (Protected)
func (h *AttachmentHandler) ConfirmPresignedUpload(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.ConfirmPresignedUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.attachmentService.ConfirmPresignedUpload(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusCreated, res)
}

// GetByID handles GET /api/v1/attachments/:id
func (h *AttachmentHandler) GetByID(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	var requestingUserID *int64
	if uid, ok := tokencontext.GetUserID(c); ok {
		requestingUserID = &uid
	}

	includeRelations := c.Query("include_relations") == "true"
	var res *dto.AttachmentResponse
	if includeRelations {
		res, err = h.attachmentService.GetByIDWithRelations(c.Request.Context(), id, requestingUserID)
	} else {
		res, err = h.attachmentService.GetByID(c.Request.Context(), id, requestingUserID)
	}
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// GetByGroup handles GET /api/v1/attachments/group/:group_id (Protected)
func (h *AttachmentHandler) GetByGroup(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	groupID, err := parseIDParam(c, "group_id")
	if err != nil {
		fail(c, err)
		return
	}

	var attType *enum.AttachmentType
	if t := c.Query("type"); t != "" {
		parsed := enum.AttachmentType(t)
		attType = &parsed
	}

	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if psStr := c.Query("page_size"); psStr != "" {
		if ps, err := strconv.Atoi(psStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	res, err := h.attachmentService.GetByGroup(c.Request.Context(), groupID, userID, attType, page, pageSize)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// GetByExpense handles GET /api/v1/attachments/expense/:expense_id (Protected)
func (h *AttachmentHandler) GetByExpense(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	expenseID, err := parseIDParam(c, "expense_id")
	if err != nil {
		fail(c, err)
		return
	}

	res, err := h.attachmentService.GetByExpense(c.Request.Context(), expenseID, userID)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// DownloadStream handles GET /api/v1/attachments/:id/download
func (h *AttachmentHandler) DownloadStream(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	var requestingUserID *int64
	if uid, ok := tokencontext.GetUserID(c); ok {
		requestingUserID = &uid
	}

	rc, meta, err := h.attachmentService.DownloadStream(c.Request.Context(), id, requestingUserID)
	if err != nil {
		fail(c, err)
		return
	}
	defer rc.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", meta.Filename))
	c.Header("Content-Type", meta.ContentType)
	if meta.Size > 0 {
		c.Header("Content-Length", strconv.FormatInt(meta.Size, 10))
	}

	_, _ = io.Copy(c.Writer, rc)
}

// Delete handles DELETE /api/v1/attachments/:id (Protected)
func (h *AttachmentHandler) Delete(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	hardDelete := c.Query("hard") == "true"
	if err := h.attachmentService.Delete(c.Request.Context(), userID, id, hardDelete); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{
		"message": "attachment deleted successfully",
	})
}

// --- Batch / Bulk Handlers ---

// UploadBatch handles POST /api/v1/attachments/batch/upload (Protected)
func (h *AttachmentHandler) UploadBatch(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, fmt.Errorf("invalid multipart form: %w", err)))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}
	if len(files) == 0 {
		fail(c, exception.Wrap(exception.ErrBadRequest, fmt.Errorf("at least one file in 'files' or 'file' is required")))
		return
	}

	typeStr := c.PostForm("type")
	if typeStr == "" {
		typeStr = string(enum.AttachmentTypeGalleryImage)
	}
	attType := enum.AttachmentType(typeStr)

	var groupID, expenseID *int64
	if gStr := c.PostForm("group_id"); gStr != "" {
		if gid, err := strconv.ParseInt(gStr, 10, 64); err == nil {
			groupID = &gid
		}
	}
	if eStr := c.PostForm("expense_id"); eStr != "" {
		if eid, err := strconv.ParseInt(eStr, 10, 64); err == nil {
			expenseID = &eid
		}
	}

	var isPublic *bool
	if pStr := c.PostForm("is_public"); pStr != "" {
		pVal := strings.EqualFold(pStr, "true") || pStr == "1"
		isPublic = &pVal
	}

	res, err := h.attachmentService.UploadBatch(c.Request.Context(), userID, files, attType, groupID, expenseID, isPublic)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusCreated, res)
}

// CreatePresignedUploadBatch handles POST /api/v1/attachments/batch/presigned-upload (Protected)
func (h *AttachmentHandler) CreatePresignedUploadBatch(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.BatchPresignedUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.attachmentService.CreatePresignedUploadBatch(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// ConfirmPresignedUploadBatch handles POST /api/v1/attachments/batch/confirm-upload (Protected)
func (h *AttachmentHandler) ConfirmPresignedUploadBatch(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.BatchConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.attachmentService.ConfirmPresignedUploadBatch(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusCreated, res)
}

// GetBatchURLs handles POST /api/v1/attachments/batch/urls
func (h *AttachmentHandler) GetBatchURLs(c *gin.Context) {
	var req dto.BatchGetURLsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	var expiry time.Duration
	if expiryStr := c.Query("expiry"); expiryStr != "" {
		if d, err := time.ParseDuration(expiryStr); err == nil && d > 0 {
			expiry = d
		}
	}

	res, err := h.attachmentService.GetBatchURLs(c.Request.Context(), req.IDs, expiry)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// DownloadZipStream handles POST /api/v1/attachments/batch/download
func (h *AttachmentHandler) DownloadZipStream(c *gin.Context) {
	var req dto.BatchDownloadZipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	archiveName := req.ArchiveName
	if archiveName == "" {
		archiveName = "attachments_archive.zip"
	}
	if !strings.HasSuffix(strings.ToLower(archiveName), ".zip") {
		archiveName += ".zip"
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", archiveName))
	c.Header("Content-Type", "application/zip")

	if err := h.attachmentService.DownloadZipStream(c.Request.Context(), req.IDs, c.Writer); err != nil {
		// Response streaming may already have started; log if unable to write header
		return
	}
}

// DeleteBatch handles DELETE /api/v1/attachments/batch (Protected)
func (h *AttachmentHandler) DeleteBatch(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	hardDelete := c.Query("hard") == "true"
	if err := h.attachmentService.DeleteBatch(c.Request.Context(), userID, req.IDs, hardDelete); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{
		"message": "attachments deleted successfully",
	})
}
