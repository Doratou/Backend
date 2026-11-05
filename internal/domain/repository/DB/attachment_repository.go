package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

// AttachmentFilter holds query parameters for filtering attachments.
type AttachmentFilter struct {
	UploadedByUserID *int64
	GroupID          *int64
	ExpenseID        *int64
	Type             *enum.AttachmentType
	IsPublic         *bool
	Search           string
}

// Pagination defines page and limit parameters.
type Pagination struct {
	Page     int
	PageSize int
}

// AttachmentRepository defines a comprehensive CRUD and query interface for Attachment entities.
type AttachmentRepository interface {
	// Create persists a new attachment record.
	Create(ctx context.Context, attachment *entity.Attachment) error

	// CreateBatch persists multiple attachment records in a single database operation.
	CreateBatch(ctx context.Context, attachments []*entity.Attachment) error

	// FindByID retrieves an attachment by primary key.
	FindByID(ctx context.Context, id int64) (*entity.Attachment, error)

	// FindByIDs retrieves multiple attachments by primary keys.
	FindByIDs(ctx context.Context, ids []int64) ([]*entity.Attachment, error)

	// FindByObjectKey retrieves an attachment by its object storage key.
	FindByObjectKey(ctx context.Context, objectKey string) (*entity.Attachment, error)

	// FindByHash retrieves an attachment by its file content hash (Deduplication).
	FindByHash(ctx context.Context, hash string) (*entity.Attachment, error)

	// FindByUploadedByUser lists attachments uploaded by a specific user.
	FindByUploadedByUser(ctx context.Context, userID int64, limit, offset int) ([]*entity.Attachment, error)

	// FindByGroup lists attachments associated with a specific group.
	FindByGroup(ctx context.Context, groupID int64, limit, offset int) ([]*entity.Attachment, error)

	// FindByExpense lists attachments tied to an expense receipt.
	FindByExpense(ctx context.Context, expenseID int64) ([]*entity.Attachment, error)

	// FindByType lists attachments filtered by type (e.g. PROFILE_IMAGE, GALLERY_IMAGE).
	FindByType(ctx context.Context, attachmentType enum.AttachmentType, limit, offset int) ([]*entity.Attachment, error)

	// List retrieves attachments based on dynamic filters and pagination, returning items and total count.
	List(ctx context.Context, filter AttachmentFilter, pagination Pagination) ([]*entity.Attachment, int64, error)

	// Update updates an existing attachment's metadata.
	Update(ctx context.Context, attachment *entity.Attachment) error

	// UpdateBatch updates common fields for multiple attachments by their IDs.
	UpdateBatch(ctx context.Context, ids []int64, updates map[string]any) error

	// DeleteByID soft-deletes an attachment record.
	DeleteByID(ctx context.Context, id int64) error

	// DeleteByIDs soft-deletes multiple attachments.
	DeleteByIDs(ctx context.Context, ids []int64) error

	// HardDeleteByID permanently deletes an attachment record.
	HardDeleteByID(ctx context.Context, id int64) error

	// HardDeleteByIDs permanently deletes multiple attachments.
	HardDeleteByIDs(ctx context.Context, ids []int64) error

	// ExistsByID checks if an attachment exists.
	ExistsByID(ctx context.Context, id int64) (bool, error)

	// Count returns the total number of non-deleted attachments.
	Count(ctx context.Context) (int64, error)
}
