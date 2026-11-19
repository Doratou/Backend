package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type adminReportService struct {
	adminReportRepo domain_repo.AdminReportRepository
	userRepo        domain_repo.UserRepository
}

func NewAdminReportService(
	adminReportRepo domain_repo.AdminReportRepository,
	userRepo domain_repo.UserRepository,
) adapter.AdminReportService {
	return &adminReportService{
		adminReportRepo: adminReportRepo,
		userRepo:        userRepo,
	}
}

func (s *adminReportService) CreateReport(ctx context.Context, reporterUserID int64, req dto.CreateReportRequest) (*dto.ReportResponse, error) {
	if !req.Reason.IsValid() {
		return nil, exception.ErrInvalidReportReason
	}

	if req.ReportedUserID == nil && req.ReportedGroupID == nil {
		return nil, exception.ErrReportTargetRequired
	}

	if req.ReportedUserID != nil {
		if *req.ReportedUserID == reporterUserID {
			return nil, exception.ErrReportSelfForbidden
		}
		exists, err := s.userRepo.ExistsByID(ctx, *req.ReportedUserID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, exception.ErrUserNotFound
		}
	}

	report := &entity.AdminReport{
		ReporterUserID:  &reporterUserID,
		ReportedUserID:  req.ReportedUserID,
		ReportedGroupID: req.ReportedGroupID,
		Reason:          req.Reason,
		Description:     strings.TrimSpace(req.Description),
		Status:          enum.ReportStatusPending,
	}

	if err := s.adminReportRepo.Create(ctx, report); err != nil {
		return nil, err
	}

	created, err := s.adminReportRepo.FindByIDWithRelations(ctx, report.ID)
	if err != nil {
		return dto.ToReportResponse(report), nil
	}
	return dto.ToReportResponse(created), nil
}

func (s *adminReportService) GetReportByID(ctx context.Context, id int64, requesterUserID int64, requesterRole enum.UserRole) (*dto.ReportResponse, error) {
	report, err := s.adminReportRepo.FindByIDWithRelations(ctx, id)
	if err != nil {
		return nil, err
	}

	if requesterRole != enum.UserRoleAdmin {
		if report.ReporterUserID == nil || *report.ReporterUserID != requesterUserID {
			return nil, exception.ErrForbiddenRole
		}
	}

	return dto.ToReportResponse(report), nil
}

func (s *adminReportService) ListMyReports(ctx context.Context, userID int64, page, limit int) (*dto.ReportListResponse, error) {
	page, limit = normalizePagination(page, limit)
	offset := (page - 1) * limit

	reports, total, err := s.adminReportRepo.FindByReporter(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]*dto.ReportResponse, 0, len(reports))
	for _, r := range reports {
		items = append(items, dto.ToReportResponse(r))
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &dto.ReportListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *adminReportService) UpdateMyPendingReport(ctx context.Context, id int64, userID int64, req dto.UpdateReportRequest) (*dto.ReportResponse, error) {
	report, err := s.adminReportRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if report.ReporterUserID == nil || *report.ReporterUserID != userID {
		return nil, exception.ErrForbiddenRole
	}

	if report.Status != enum.ReportStatusPending {
		return nil, exception.ErrReportAlreadyResolved
	}

	if req.Reason != nil {
		if !req.Reason.IsValid() {
			return nil, exception.ErrInvalidReportReason
		}
		report.Reason = *req.Reason
	}

	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		if desc != "" {
			report.Description = desc
		}
	}

	if err := s.adminReportRepo.Update(ctx, report); err != nil {
		return nil, err
	}

	updated, err := s.adminReportRepo.FindByIDWithRelations(ctx, id)
	if err != nil {
		return dto.ToReportResponse(report), nil
	}
	return dto.ToReportResponse(updated), nil
}

func (s *adminReportService) CancelMyReport(ctx context.Context, id int64, userID int64) error {
	report, err := s.adminReportRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if report.ReporterUserID == nil || *report.ReporterUserID != userID {
		return exception.ErrForbiddenRole
	}

	if report.Status != enum.ReportStatusPending {
		return exception.ErrReportAlreadyResolved
	}

	return s.adminReportRepo.DeleteByID(ctx, id)
}

func (s *adminReportService) ListReports(ctx context.Context, req dto.ReportFilterRequest) (*dto.ReportListResponse, error) {
	if req.Status != nil && *req.Status != "" && !req.Status.IsValid() {
		return nil, exception.ErrInvalidReportStatus
	}
	if req.Reason != nil && *req.Reason != "" && !req.Reason.IsValid() {
		return nil, exception.ErrInvalidReportReason
	}

	page, limit := normalizePagination(req.Page, req.Limit)
	offset := (page - 1) * limit

	orderDesc := true
	if req.OrderDesc != nil {
		orderDesc = *req.OrderDesc
	}

	filter := domain_repo.ReportFilter{
		Status:           req.Status,
		Reason:           req.Reason,
		ReporterUserID:   req.ReporterUserID,
		ReportedUserID:   req.ReportedUserID,
		ReportedGroupID:  req.ReportedGroupID,
		ReviewedByUserID: req.ReviewedByUserID,
		Search:           strings.TrimSpace(req.Search),
		OrderBy:          req.OrderBy,
		OrderDesc:        orderDesc,
	}

	reports, total, err := s.adminReportRepo.FindAll(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]*dto.ReportResponse, 0, len(reports))
	for _, r := range reports {
		items = append(items, dto.ToReportResponse(r))
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &dto.ReportListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *adminReportService) ReviewReport(ctx context.Context, id int64, reviewerUserID int64, req dto.ReviewReportRequest) (*dto.ReportResponse, error) {
	if !req.Status.IsValid() {
		return nil, exception.ErrInvalidReportStatus
	}

	exists, err := s.adminReportRepo.NotSolvedByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, exception.ErrReportNotFound
	}

	var resolvedAt time.Time
	if req.Status == enum.ReportStatusResolved || req.Status == enum.ReportStatusRejected || req.Status == enum.ReportStatusDismissed {
		resolvedAt = time.Now().UTC()
	}

	response := strings.TrimSpace(req.Response)
	if err := s.adminReportRepo.UpdateReview(ctx, id, req.Status, reviewerUserID, response, resolvedAt); err != nil {
		return nil, err
	}

	updated, err := s.adminReportRepo.FindByIDWithRelations(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.ToReportResponse(updated), nil
}

func (s *adminReportService) DeleteReport(ctx context.Context, id int64) error {
	return s.adminReportRepo.DeleteByID(ctx, id)
}

func (s *adminReportService) RestoreReport(ctx context.Context, id int64) error {
	return s.adminReportRepo.RestoreByID(ctx, id)
}

func (s *adminReportService) HardDeleteReport(ctx context.Context, id int64) error {
	return s.adminReportRepo.HardDeleteByID(ctx, id)
}

func (s *adminReportService) GetReportStats(ctx context.Context) (*dto.ReportStatsResponse, error) {
	byStatus, err := s.adminReportRepo.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}

	var total int64
	for _, count := range byStatus {
		total += count
	}

	return &dto.ReportStatsResponse{
		TotalReports: total,
		ByStatus:     byStatus,
	}, nil
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	return page, limit
}
