package DB

import (
	"context"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

// ReportFilter defines comprehensive criteria for querying and filtering admin reports.
type ReportFilter struct {
	Status           *enum.ReportStatus
	Reason           *enum.ReportReason
	ReporterUserID   *int64
	ReportedUserID   *int64
	ReportedGroupID  *int64
	ReviewedByUserID *int64
	Search           string
	FromDate         *time.Time
	ToDate           *time.Time
	OrderBy          string
	OrderDesc        bool
}

// AdminReportRepository defines a comprehensive CRUD and query interface for AdminReport entities.
type AdminReportRepository interface {
	// Create persists a new admin report record.
	Create(ctx context.Context, report *entity.AdminReport) error

	// FindByID retrieves a report by primary key (without relations).
	FindByID(ctx context.Context, id int64) (*entity.AdminReport, error)

	// FindByIDWithRelations retrieves a report by primary key with ReporterUser, ReportedUser, and ReviewedByUser preloaded.
	FindByIDWithRelations(ctx context.Context, id int64) (*entity.AdminReport, error)

	// FindAll returns a slice of reports matching filter criteria with pagination and total count.
	FindAll(ctx context.Context, filter ReportFilter, limit, offset int) ([]*entity.AdminReport, int64, error)

	// FindByReporter lists reports submitted by a specific user with pagination.
	FindByReporter(ctx context.Context, reporterUserID int64, limit, offset int) ([]*entity.AdminReport, int64, error)

	// FindByReportedUser lists reports targeting a specific user with pagination.
	FindByReportedUser(ctx context.Context, reportedUserID int64, limit, offset int) ([]*entity.AdminReport, int64, error)

	// FindByReportedGroup lists reports targeting a specific group with pagination.
	FindByReportedGroup(ctx context.Context, groupID int64, limit, offset int) ([]*entity.AdminReport, int64, error)

	// Update updates all writable fields of an existing report.
	Update(ctx context.Context, report *entity.AdminReport) error

	// UpdateReview updates status, reviewer, response text, and resolved_at timestamp.
	UpdateReview(ctx context.Context, id int64, status enum.ReportStatus, reviewerID int64, response string, resolvedAt time.Time) error

	// DeleteByID performs a soft delete on a report record.
	DeleteByID(ctx context.Context, id int64) error

	// RestoreByID restores a soft-deleted report record.
	RestoreByID(ctx context.Context, id int64) error

	// HardDeleteByID permanently deletes a report record from the database.
	HardDeleteByID(ctx context.Context, id int64) error

	// CountByStatus aggregates report counts grouped by their status (for dashboard stats).
	CountByStatus(ctx context.Context) (map[enum.ReportStatus]int64, error)

	// ExistsByID checks if a report exists by ID.
	ExistsByID(ctx context.Context, id int64) (bool, error)
}
