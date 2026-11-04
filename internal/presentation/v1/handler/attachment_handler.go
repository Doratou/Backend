package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
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

	res, err := h.attachmentService.GetByID(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// List handles GET /api/v1/attachments
func (h *AttachmentHandler) List(c *gin.Context) {
	var filter dto.AttachmentFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.attachmentService.List(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// GetURL handles GET /api/v1/attachments/:id/url
func (h *AttachmentHandler) GetURL(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	var expiry time.Duration
	if expiryStr := c.Query("expiry"); expiryStr != "" {
		if d, err := time.ParseDuration(expiryStr); err == nil && d > 0 {
			expiry = d
		}
	}

	url, err := h.attachmentService.GetFreshAccessURL(c.Request.Context(), id, expiry)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{
		"access_url": url,
	})
}

// DownloadStream handles GET /api/v1/attachments/:id/download
func (h *AttachmentHandler) DownloadStream(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	rc, meta, err := h.attachmentService.DownloadStream(c.Request.Context(), id)
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

// Update handles PATCH /api/v1/attachments/:id (Protected)
func (h *AttachmentHandler) Update(c *gin.Context) {
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

	var req dto.UpdateAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.attachmentService.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
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
