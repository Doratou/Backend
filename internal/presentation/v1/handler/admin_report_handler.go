package handler

import (
	"net/http"
	"strconv"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type AdminReportHandler struct {
	reportService adapter.AdminReportService
}

func NewAdminReportHandler(reportService adapter.AdminReportService) *AdminReportHandler {
	return &AdminReportHandler{
		reportService: reportService,
	}
}

// ==========================================
// User Endpoints
// ==========================================

// SubmitReport handles POST /api/v1/reports
func (h *AdminReportHandler) SubmitReport(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.reportService.CreateReport(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusCreated, res)
}

// ListMyReports handles GET /api/v1/reports/my
func (h *AdminReportHandler) ListMyReports(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.reportService.ListMyReports(c.Request.Context(), userID, page, limit)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// GetReport handles GET /api/v1/reports/:id
func (h *AdminReportHandler) GetReport(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	role, _ := tokencontext.GetUserRole(c)

	res, err := h.reportService.GetReportByID(c.Request.Context(), id, userID, role)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// UpdateMyPendingReport handles PATCH /api/v1/reports/:id
func (h *AdminReportHandler) UpdateMyPendingReport(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.reportService.UpdateMyPendingReport(c.Request.Context(), id, userID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// CancelMyReport handles DELETE /api/v1/reports/:id
func (h *AdminReportHandler) CancelMyReport(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	if err := h.reportService.CancelMyReport(c.Request.Context(), id, userID); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{"message": "report cancelled successfully"})
}

// ==========================================
// Admin Endpoints
// ==========================================

// AdminListReports handles GET /api/v1/admin/reports
func (h *AdminReportHandler) AdminListReports(c *gin.Context) {
	var req dto.ReportFilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.reportService.ListReports(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// AdminGetStats handles GET /api/v1/admin/reports/stats
func (h *AdminReportHandler) AdminGetStats(c *gin.Context) {
	res, err := h.reportService.GetReportStats(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// AdminGetReport handles GET /api/v1/admin/reports/:id
func (h *AdminReportHandler) AdminGetReport(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	res, err := h.reportService.GetReportByID(c.Request.Context(), id, 0, enum.UserRoleAdmin)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// AdminReviewReport handles PUT /api/v1/admin/reports/:id/review
func (h *AdminReportHandler) AdminReviewReport(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	reviewerID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.ReviewReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.reportService.ReviewReport(c.Request.Context(), id, reviewerID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// AdminDeleteReport handles DELETE /api/v1/admin/reports/:id (soft delete)
func (h *AdminReportHandler) AdminDeleteReport(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	if err := h.reportService.DeleteReport(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{"message": "report soft deleted successfully"})
}

// AdminRestoreReport handles POST /api/v1/admin/reports/:id/restore
func (h *AdminReportHandler) AdminRestoreReport(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	if err := h.reportService.RestoreReport(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{"message": "report restored successfully"})
}

// AdminHardDeleteReport handles DELETE /api/v1/admin/reports/:id/hard
func (h *AdminReportHandler) AdminHardDeleteReport(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	if err := h.reportService.HardDeleteReport(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{"message": "report permanently deleted"})
}
