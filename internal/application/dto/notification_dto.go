package dto

import (
	"encoding/json"
	"time"

	"github.com/Doratou/Backend/internal/domain/broker/payload"
)

// GetNotificationsQuery defines query parameters for fetching paginated notifications.
type GetNotificationsQuery struct {
	CursorID int64  `form:"cursor_id"`
	Limit    int    `form:"limit"`
	Status   string `form:"status"` // "all", "unread", "read"
	GroupID  *int64 `form:"group_id"`
}

// SyncNotificationsQuery defines query parameters for catching up missed notifications.
type SyncNotificationsQuery struct {
	SinceID int64 `form:"since_id"`
	Limit   int   `form:"limit"`
}

// ClearNotificationsQuery specifies whether to purge only read notifications or all.
type ClearNotificationsQuery struct {
	OnlyRead bool `form:"only_read"`
}

// NotificationResponse represents a single notification record in HTTP responses.
type NotificationResponse struct {
	ID          int64           `json:"id"`
	RecipientID int64           `json:"recipient_id"`
	ActorID     *int64          `json:"actor_id,omitempty"`
	GroupID     *int64          `json:"group_id,omitempty"`
	Event       payload.WSEvent `json:"event"`
	Data        json.RawMessage `json:"data"`
	IsRead      bool            `json:"is_read"`
	CreatedAt   time.Time       `json:"created_at"`
}

// NotificationListResponse represents the cursor-paginated timeline response.
type NotificationListResponse struct {
	Items       []NotificationResponse `json:"items"`
	UnreadCount int64                  `json:"unread_count"`
	NextCursor  int64                  `json:"next_cursor"`
	HasMore     bool                   `json:"has_more"`
}

// NotificationSyncResponse represents the response when catching up missed notifications.
type NotificationSyncResponse struct {
	Items    []NotificationResponse `json:"items"`
	LatestID int64                  `json:"latest_id"`
	HasMore  bool                   `json:"has_more"`
}

// UnreadCountResponse represents the badge count endpoint response.
type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
	LastReadID  int64 `json:"last_read_id"`
}

// MarkAsReadRequest defines the target notification up to which items should be marked as read.
type MarkAsReadRequest struct {
	UpToID int64 `json:"up_to_id"`
}

// MarkAsReadResponse contains the updated bookmark position and remaining unread count.
type MarkAsReadResponse struct {
	LastReadID  int64 `json:"last_read_id"`
	UnreadCount int64 `json:"unread_count"`
}

// GroupAnnouncementRequest represents payload for Case 9 (Owner manual announcement).
type GroupAnnouncementRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Message string `json:"message" binding:"required,min=1,max=2000"`
}

// GroupAnnouncementResponse represents the summary of an announcement dispatch.
type GroupAnnouncementResponse struct {
	DispatchedCount int   `json:"dispatched_count"`
	GroupID         int64 `json:"group_id"`
}
