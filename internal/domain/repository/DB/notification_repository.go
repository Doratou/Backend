package DB

import (
	"context"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
)

// NotificationFilter defines query criteria and pagination for user notifications.
type NotificationFilter struct {
	RecipientID int64
	CursorID    int64 // For keyset pagination (id < CursorID when CursorID > 0)
	Limit       int   // Max items to return (default 20, max 100)
}

// NotificationRepository defines full CRUD operations, High-Water Mark bookmarking,
// and soft/hard delete capabilities for notifications without foreign key coupling.
type NotificationRepository interface {
	// Create persists a single notification record.
	Create(ctx context.Context, notif *entity.Notification) error

	// CreateBatch inserts multiple notifications in bulk within a single transaction.
	CreateBatch(ctx context.Context, notifs []*entity.Notification) error

	// FindByID retrieves a single active notification by ID.
	FindByID(ctx context.Context, id int64) (*entity.Notification, error)

	// ListByRecipient returns active notifications for a user ordered by id DESC with computed IsRead status.
	ListByRecipient(ctx context.Context, filter NotificationFilter) ([]*entity.Notification, error)

	// GetUnreadCount returns the count of active notifications with id > user's last_read_id.
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)

	// GetLastReadID retrieves the user's current High-Water Mark last_read_id.
	GetLastReadID(ctx context.Context, userID int64) (int64, error)

	// Update updates editable notification metadata (data payload or error).
	Update(ctx context.Context, notif *entity.Notification) error

	// UpdateLastError updates or clears the last_error field on a notification.
	UpdateLastError(ctx context.Context, id int64, lastErr *string) error

	// MarkAsRead advances the user's High-Water Mark last_read_id to max(current, upToID) in O(1).
	MarkAsRead(ctx context.Context, userID int64, upToID int64) error

	// SoftDelete marks a notification as deleted by setting deleted_at = CURRENT_TIMESTAMP.
	SoftDelete(ctx context.Context, id int64) error

	// SoftDeleteAllByRecipient soft-deletes all active notifications for a specific user.
	SoftDeleteAllByRecipient(ctx context.Context, recipientID int64) error

	// HardDelete permanently removes a notification record by ID.
	HardDelete(ctx context.Context, id int64) error

	// PurgeSoftDeletedBefore permanently removes soft-deleted notifications older than a retention threshold.
	PurgeSoftDeletedBefore(ctx context.Context, before time.Time) (int64, error)
}
