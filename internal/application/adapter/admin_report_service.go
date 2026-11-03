package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type AdminReportService interface {
	// User operations
	CreateReport(ctx context.Context, reporterUserID int64, req dto.CreateReportRequest) (*dto.ReportResponse, error)
	GetReportByID(ctx context.Context, id int64, requesterUserID int64, requesterRole enum.UserRole) (*dto.ReportResponse, error)
	ListMyReports(ctx context.Context, userID int64, page, limit int) (*dto.ReportListResponse, error)
	UpdateMyPendingReport(ctx context.Context, id int64, userID int64, req dto.UpdateReportRequest) (*dto.ReportResponse, error)
	CancelMyReport(ctx context.Context, id int64, userID int64) error

	// Admin operations
	ListReports(ctx context.Context, req dto.ReportFilterRequest) (*dto.ReportListResponse, error)
	ReviewReport(ctx context.Context, id int64, reviewerUserID int64, req dto.ReviewReportRequest) (*dto.ReportResponse, error)
	DeleteReport(ctx context.Context, id int64) error
	RestoreReport(ctx context.Context, id int64) error
	HardDeleteReport(ctx context.Context, id int64) error
	GetReportStats(ctx context.Context) (*dto.ReportStatsResponse, error)
}
