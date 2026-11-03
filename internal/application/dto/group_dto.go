package dto

import "time"

type CreateGroupRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=150"`
	Description *string `json:"description" binding:"omitempty,max=5000"`
}

type UpdateGroupRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=150"`
	Description *string `json:"description" binding:"omitempty,max=5000"`
}

type CloneGroupRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=150"`
	Description *string `json:"description" binding:"omitempty,max=5000"`
}

type GroupResponse struct {
	ID                int64      `json:"id"`
	Title             string     `json:"title"`
	Description       *string    `json:"description,omitempty"`
	CreatedByUserID   int64      `json:"created_by_user_id"`
	CoverAttachmentID *int64     `json:"cover_attachment_id,omitempty"`
	Status            string     `json:"status"`
	IsPremium         bool       `json:"is_premium"`
	ArchivedAt        *time.Time `json:"archived_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	MemberCount       int64      `json:"member_count,omitempty"`
}
