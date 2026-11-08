package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

// CreateReportRequest holds parameters required when a user submits a moderation report.
type CreateReportRequest struct {
	ReportedUserID  *int64            `json:"reported_user_id,omitempty"`
	ReportedGroupID *int64            `json:"reported_group_id,omitempty"`
	Reason          enum.ReportReason `json:"reason" binding:"required"`
	Description     string            `json:"description" binding:"required,min=5,max=3000"`
}

// UpdateReportRequest allows a reporter to update their pending report's reason or description.
type UpdateReportRequest struct {
	Reason      *enum.ReportReason `json:"reason,omitempty"`
	Description *string            `json:"description,omitempty" binding:"omitempty,min=5,max=3000"`
}

// ReviewReportRequest holds parameters used by an administrator to review and resolve a report.
type ReviewReportRequest struct {
	Status   enum.ReportStatus `json:"status" binding:"required"`
	Response string            `json:"response" binding:"required,min=2,max=3000"`
}

// ReportFilterRequest holds query parameters for filtering and paginating admin reports.
type ReportFilterRequest struct {
	Page             int                `form:"page,default=1"`
	Limit            int                `form:"limit,default=20"`
	Status           *enum.ReportStatus `form:"status"`
	Reason           *enum.ReportReason `form:"reason"`
	ReporterUserID   *int64             `form:"reporter_user_id"`
	ReportedUserID   *int64             `form:"reported_user_id"`
	ReportedGroupID  *int64             `form:"reported_group_id"`
	ReviewedByUserID *int64             `form:"reviewed_by_user_id"`
	Search           string             `form:"search"`
	OrderBy          string             `form:"order_by,default=created_at"`
	OrderDesc        *bool              `form:"order_desc"`
}

// ReportUserSummary encapsulates non-sensitive user identity details for response serialization.
type ReportUserSummary struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

// ReportResponse represents the detailed serialized representation of an AdminReport.
type ReportResponse struct {
	ID               int64              `json:"id"`
	ReporterUserID   *int64             `json:"reporter_user_id,omitempty"`
	ReporterUser     *ReportUserSummary `json:"reporter_user,omitempty"`
	ReportedUserID   *int64             `json:"reported_user_id,omitempty"`
	ReportedUser     *ReportUserSummary `json:"reported_user,omitempty"`
	ReportedGroupID  *int64             `json:"reported_group_id,omitempty"`
	Reason           enum.ReportReason  `json:"reason"`
	Description      string             `json:"description"`
	Status           enum.ReportStatus  `json:"status"`
	ReviewedByUserID *int64             `json:"reviewed_by_user_id,omitempty"`
	ReviewedByUser   *ReportUserSummary `json:"reviewed_by_user,omitempty"`
	Response         *string            `json:"response,omitempty"`
	ResolvedAt       *time.Time         `json:"resolved_at,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// ReportListResponse represents a standard paginated response envelope for report listings.
type ReportListResponse struct {
	Items      []*ReportResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

// ReportStatsResponse holds summary statistics for reports on the admin moderation dashboard.
type ReportStatsResponse struct {
	TotalReports int64                      `json:"total_reports"`
	ByStatus     map[enum.ReportStatus]int64 `json:"by_status"`
}

// ToUserSummary extracts non-sensitive user fields into ReportUserSummary.
func ToUserSummary(u *entity.User) *ReportUserSummary {
	if u == nil {
		return nil
	}
	return &ReportUserSummary{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Phone: u.Phone,
	}
}

// ToReportResponse maps an entity.AdminReport into a presentation-safe DTO.
func ToReportResponse(r *entity.AdminReport) *ReportResponse {
	if r == nil {
		return nil
	}
	return &ReportResponse{
		ID:               r.ID,
		ReporterUserID:   r.ReporterUserID,
		ReporterUser:     ToUserSummary(r.ReporterUser),
		ReportedUserID:   r.ReportedUserID,
		ReportedUser:     ToUserSummary(r.ReportedUser),
		ReportedGroupID:  r.ReportedGroupID,
		Reason:           r.Reason,
		Description:      r.Description,
		Status:           r.Status,
		ReviewedByUserID: r.ReviewedByUserID,
		ReviewedByUser:   ToUserSummary(r.ReviewedByUser),
		Response:         r.Response,
		ResolvedAt:       r.ResolvedAt,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}
