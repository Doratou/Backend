package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
)

// NotificationService defines comprehensive business operations for reading, managing,
// and dispatching notifications across the application.
type NotificationService interface {
	// GetNotifications retrieves a cursor-paginated list of notifications for the authenticated user.
	GetNotifications(ctx context.Context, userID int64, req dto.GetNotificationsQuery) (*dto.NotificationListResponse, error)

	// GetUnreadCount returns the count of unread notifications, optionally filtered by group.
	GetUnreadCount(ctx context.Context, userID int64, groupID *int64) (*dto.UnreadCountResponse, error)

	// SyncNotifications retrieves missed notifications since a given ID (catch-up protocol).
	SyncNotifications(ctx context.Context, userID int64, req dto.SyncNotificationsQuery) (*dto.NotificationSyncResponse, error)

	// GetNotificationByID retrieves a single notification ensuring ownership.
	GetNotificationByID(ctx context.Context, userID, id int64) (*dto.NotificationResponse, error)

	// MarkAsRead advances the user's High-Water Mark to a specific ID (or all notifications if up_to_id <= 0).
	MarkAsRead(ctx context.Context, userID int64, req dto.MarkAsReadRequest) (*dto.MarkAsReadResponse, error)

	// MarkSingleAsRead marks a specific notification as read.
	MarkSingleAsRead(ctx context.Context, userID, id int64) (*dto.MarkAsReadResponse, error)

	// DeleteNotification soft-deletes a single notification owned by the user.
	DeleteNotification(ctx context.Context, userID, id int64) error

	// ClearNotifications clears all notifications or only read notifications for the user.
	ClearNotifications(ctx context.Context, userID int64, onlyRead bool) error

	// CreateGroupAnnouncement allows a group owner to broadcast a custom announcement to all group members.
	CreateGroupAnnouncement(ctx context.Context, ownerUserID, groupID int64, req dto.GroupAnnouncementRequest) (*dto.GroupAnnouncementResponse, error)

	// DispatchToUser saves and publishes a notification targeted to a single recipient.
	DispatchToUser(ctx context.Context, recipientID int64, actorID *int64, groupID *int64, event payload.WSEvent, data any) error

	// DispatchToUsers saves and publishes notifications targeted to multiple recipients.
	DispatchToUsers(ctx context.Context, recipientIDs []int64, actorID *int64, groupID *int64, event payload.WSEvent, data any) error

	// DispatchToGroup saves and publishes notifications targeted to active members of a group.
	DispatchToGroup(ctx context.Context, groupID int64, actorID *int64, recipientMemberIDs []int64, event payload.WSEvent, data any) error

	// Broadcast publishes a system-wide broadcast event.
	Broadcast(ctx context.Context, actorID *int64, event payload.WSEvent, data any) error
}
