package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type adminReportRepository struct {
	db database.Database
}

func NewAdminReportRepository(db database.Database) domain_repo.AdminReportRepository {
	return &adminReportRepository{db: db}
}

func (r *adminReportRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *adminReportRepository) Create(ctx context.Context, report *entity.AdminReport) error {
	db := r.getDB(ctx)
	if err := db.Create(report).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *adminReportRepository) FindByID(ctx context.Context, id int64) (*entity.AdminReport, error) {
	var report entity.AdminReport
	db := r.getDB(ctx)

	err := db.Where("id = ?", id).First(&report).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrReportNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &report, nil
}

func (r *adminReportRepository) FindByIDWithRelations(ctx context.Context, id int64) (*entity.AdminReport, error) {
	var report entity.AdminReport
	db := r.getDB(ctx)

	err := db.
		Preload("ReporterUser").
		Preload("ReportedUser").
		Preload("ReviewedByUser").
		Where("id = ?", id).
		First(&report).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrReportNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &report, nil
}

func (r *adminReportRepository) FindAll(ctx context.Context, filter domain_repo.ReportFilter, limit, offset int) ([]*entity.AdminReport, int64, error) {
	var reports []*entity.AdminReport
	var total int64
	db := r.getDB(ctx).Model(&entity.AdminReport{})

	// Dynamic Filtering
	if filter.Status != nil && *filter.Status != "" {
		db = db.Where("status = ?", *filter.Status)
	}
	if filter.Reason != nil && *filter.Reason != "" {
		db = db.Where("reason = ?", *filter.Reason)
	}
	if filter.ReporterUserID != nil {
		db = db.Where("reporter_user_id = ?", *filter.ReporterUserID)
	}
	if filter.ReportedUserID != nil {
		db = db.Where("reported_user_id = ?", *filter.ReportedUserID)
	}
	if filter.ReportedGroupID != nil {
		db = db.Where("reported_group_id = ?", *filter.ReportedGroupID)
	}
	if filter.ReviewedByUserID != nil {
		db = db.Where("reviewed_by_user_id = ?", *filter.ReviewedByUserID)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		pattern := "%" + search + "%"
		db = db.Where("description ILIKE ? OR response ILIKE ?", pattern, pattern)
	}
	if filter.FromDate != nil {
		db = db.Where("created_at >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		db = db.Where("created_at <= ?", *filter.ToDate)
	}

	// Count matching records before pagination
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	// Sanitize OrderBy to prevent SQL injection
	orderCol := "created_at"
	switch strings.ToLower(filter.OrderBy) {
	case "id":
		orderCol = "id"
	case "status":
		orderCol = "status"
	case "reason":
		orderCol = "reason"
	case "updated_at":
		orderCol = "updated_at"
	case "resolved_at":
		orderCol = "resolved_at"
	default:
		orderCol = "created_at"
	}

	direction := "DESC"
	if !filter.OrderDesc && filter.OrderBy != "" {
		direction = "ASC"
	}
	db = db.Order(fmt.Sprintf("%s %s", orderCol, direction))

	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	err := db.
		Preload("ReporterUser").
		Preload("ReportedUser").
		Preload("ReviewedByUser").
		Find(&reports).Error

	if err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return reports, total, nil
}

func (r *adminReportRepository) FindByReporter(ctx context.Context, reporterUserID int64, limit, offset int) ([]*entity.AdminReport, int64, error) {
	var reports []*entity.AdminReport
	var total int64
	db := r.getDB(ctx).Model(&entity.AdminReport{}).Where("reporter_user_id = ?", reporterUserID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	db = db.Order("created_at DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	err := db.
		Preload("ReporterUser").
		Preload("ReportedUser").
		Preload("ReviewedByUser").
		Find(&reports).Error

	if err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return reports, total, nil
}

func (r *adminReportRepository) FindByReportedUser(ctx context.Context, reportedUserID int64, limit, offset int) ([]*entity.AdminReport, int64, error) {
	var reports []*entity.AdminReport
	var total int64
	db := r.getDB(ctx).Model(&entity.AdminReport{}).Where("reported_user_id = ?", reportedUserID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	db = db.Order("created_at DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	err := db.
		Preload("ReporterUser").
		Preload("ReportedUser").
		Preload("ReviewedByUser").
		Find(&reports).Error

	if err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return reports, total, nil
}

func (r *adminReportRepository) FindByReportedGroup(ctx context.Context, groupID int64, limit, offset int) ([]*entity.AdminReport, int64, error) {
	var reports []*entity.AdminReport
	var total int64
	db := r.getDB(ctx).Model(&entity.AdminReport{}).Where("reported_group_id = ?", groupID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	db = db.Order("created_at DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	err := db.
		Preload("ReporterUser").
		Preload("ReportedUser").
		Preload("ReviewedByUser").
		Find(&reports).Error

	if err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return reports, total, nil
}

func (r *adminReportRepository) Update(ctx context.Context, report *entity.AdminReport) error {
	db := r.getDB(ctx)
	if err := db.Save(report).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *adminReportRepository) UpdateReview(ctx context.Context, id int64, status enum.ReportStatus, reviewerID int64, response string, resolvedAt time.Time) error {
	db := r.getDB(ctx)
	updates := map[string]any{
		"status":              status,
		"reviewed_by_user_id": reviewerID,
		"response":            response,
		"resolved_at":         resolvedAt,
		"updated_at":          time.Now().UTC(),
	}

	res := db.Model(&entity.AdminReport{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrReportNotFound
	}
	return nil
}

func (r *adminReportRepository) DeleteByID(ctx context.Context, id int64) error {
	db := r.getDB(ctx)
	res := db.Where("id = ?", id).Delete(&entity.AdminReport{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrReportNotFound
	}
	return nil
}

func (r *adminReportRepository) RestoreByID(ctx context.Context, id int64) error {
	db := r.getDB(ctx)
	res := db.Unscoped().Model(&entity.AdminReport{}).Where("id = ?", id).Update("deleted_at", nil)
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrReportNotFound
	}
	return nil
}

func (r *adminReportRepository) HardDeleteByID(ctx context.Context, id int64) error {
	db := r.getDB(ctx)
	res := db.Unscoped().Where("id = ?", id).Delete(&entity.AdminReport{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrReportNotFound
	}
	return nil
}

func (r *adminReportRepository) CountByStatus(ctx context.Context) (map[enum.ReportStatus]int64, error) {
	db := r.getDB(ctx)

	type statusRow struct {
		Status enum.ReportStatus `gorm:"column:status"`
		Count  int64             `gorm:"column:count"`
	}

	var rows []statusRow
	err := db.Model(&entity.AdminReport{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&rows).Error

	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}

	result := make(map[enum.ReportStatus]int64)
	// Initialize default zeros for all statuses
	result[enum.ReportStatusPending] = 0
	result[enum.ReportStatusUnderReview] = 0
	result[enum.ReportStatusResolved] = 0
	result[enum.ReportStatusRejected] = 0
	result[enum.ReportStatusDismissed] = 0

	for _, row := range rows {
		result[row.Status] = row.Count
	}

	return result, nil
}

func (r *adminReportRepository) ExistsByID(ctx context.Context, id int64) (bool, error) {
	db := r.getDB(ctx)
	var count int64
	if err := db.Model(&entity.AdminReport{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}

func (r *adminReportRepository) NotSolvedByID(ctx context.Context, id int64) (bool, error) {
	db := r.getDB(ctx)
	var count int64
	notSolvedStatus := []enum.ReportStatus{enum.ReportStatusPending, enum.ReportStatusUnderReview}

	if err := db.Model(&entity.AdminReport{}).
		Where("id = ? AND status IN ?", id, notSolvedStatus).
		Count(&count).Error; err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}
