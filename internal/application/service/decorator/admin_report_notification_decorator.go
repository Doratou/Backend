package decorator

import (
	"context"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/enum"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type adminReportNotificationDecorator struct {
	next     adapter.AdminReportService
	notifSvc adapter.NotificationService
	userRepo domain_db.UserRepository
}

// NewAdminReportNotificationDecorator wraps an adapter.AdminReportService to notify the
// original reporter when an admin resolves, rejects, or dismisses their report.
func NewAdminReportNotificationDecorator(
	next adapter.AdminReportService,
	notifSvc adapter.NotificationService,
	userRepo domain_db.UserRepository,
) adapter.AdminReportService {
	return &adminReportNotificationDecorator{
		next:     next,
		notifSvc: notifSvc,
		userRepo: userRepo,
	}
}

func (d *adminReportNotificationDecorator) ReviewReport(ctx context.Context, id int64, reviewerUserID int64, req dto.ReviewReportRequest) (*dto.ReportResponse, error) {
	res, err := d.next.ReviewReport(ctx, id, reviewerUserID, req)
	if err != nil {
		return nil, err
	}

	if d.notifSvc != nil && isTerminalReportStatus(res.Status) && res.ReporterUserID != nil && *res.ReporterUserID > 0 {
		reporterID := *res.ReporterUserID
		answer := req.Response
		if res.Response != nil && *res.Response != "" {
			answer = *res.Response
		}
		var adminName string
		if res.ReviewedByUser != nil {
			adminName = res.ReviewedByUser.Name
		}
		if adminName == "" {
			adminName, _ = resolveUserProfile(ctx, d.userRepo, reviewerUserID)
		}
		answeredAt := time.Now().UTC()
		if res.ResolvedAt != nil {
			answeredAt = *res.ResolvedAt
		} else if !res.UpdatedAt.IsZero() {
			answeredAt = res.UpdatedAt
		}

		dispatchUserNotification(ctx, d.notifSvc, reporterID, &reviewerUserID, nil, payload.WSEventAdminReportAnswered, payload.AdminReportAnsweredData{
			ReportID:   res.ID,
			UserID:     strconv.FormatInt(reporterID, 10),
			Subject:    string(res.Reason),
			Answer:     answer,
			AdminName:  adminName,
			AnsweredAt: answeredAt,
		})
	}

	return res, nil
}

func isTerminalReportStatus(status enum.ReportStatus) bool {
	return status == enum.ReportStatusResolved || status == enum.ReportStatusRejected || status == enum.ReportStatusDismissed
}

func (d *adminReportNotificationDecorator) CreateReport(ctx context.Context, reporterUserID int64, req dto.CreateReportRequest) (*dto.ReportResponse, error) {
	return d.next.CreateReport(ctx, reporterUserID, req)
}

func (d *adminReportNotificationDecorator) GetReportByID(ctx context.Context, id int64, requesterUserID int64, requesterRole enum.UserRole) (*dto.ReportResponse, error) {
	return d.next.GetReportByID(ctx, id, requesterUserID, requesterRole)
}

func (d *adminReportNotificationDecorator) ListMyReports(ctx context.Context, userID int64, page, limit int) (*dto.ReportListResponse, error) {
	return d.next.ListMyReports(ctx, userID, page, limit)
}

func (d *adminReportNotificationDecorator) UpdateMyPendingReport(ctx context.Context, id int64, userID int64, req dto.UpdateReportRequest) (*dto.ReportResponse, error) {
	return d.next.UpdateMyPendingReport(ctx, id, userID, req)
}

func (d *adminReportNotificationDecorator) CancelMyReport(ctx context.Context, id int64, userID int64) error {
	return d.next.CancelMyReport(ctx, id, userID)
}

func (d *adminReportNotificationDecorator) ListReports(ctx context.Context, req dto.ReportFilterRequest) (*dto.ReportListResponse, error) {
	return d.next.ListReports(ctx, req)
}

func (d *adminReportNotificationDecorator) DeleteReport(ctx context.Context, id int64) error {
	return d.next.DeleteReport(ctx, id)
}

func (d *adminReportNotificationDecorator) RestoreReport(ctx context.Context, id int64) error {
	return d.next.RestoreReport(ctx, id)
}

func (d *adminReportNotificationDecorator) HardDeleteReport(ctx context.Context, id int64) error {
	return d.next.HardDeleteReport(ctx, id)
}

func (d *adminReportNotificationDecorator) GetReportStats(ctx context.Context) (*dto.ReportStatsResponse, error) {
	return d.next.GetReportStats(ctx)
}
