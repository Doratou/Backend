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
	ID                int64                 `json:"id"`
	Title             string                `json:"title"`
	Description       *string               `json:"description,omitempty"`
	CreatedByUserID   int64                 `json:"created_by_user_id"`
	CoverAttachmentID *int64                `json:"cover_attachment_id,omitempty"`
	Status            string                `json:"status"`
	IsPremium         bool                  `json:"is_premium"`
	ArchivedAt        *time.Time            `json:"archived_at,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	MemberCount       int64                 `json:"member_count,omitempty"`
	Members           []GroupMemberResponse `json:"members,omitempty"`
	Invites           []GroupInviteResponse `json:"invites,omitempty"`
}

type GroupMemberResponse struct {
	MemberID int64     `json:"member_id"`
	UserID   int64     `json:"user_id"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type TransferGroupOwnershipRequest struct {
	NewOwnerID int64 `json:"new_owner_id" binding:"required,gt=0"`
}

type CreateGroupInviteRequest struct {
	Password  string     `json:"password" binding:"required,min=8,max=72"`
	ExpiresAt *time.Time `json:"expires_at"`
	MaxUses   *int       `json:"max_uses" binding:"omitempty,gt=0"`
}

type ChangeGroupInvitePasswordRequest struct {
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type UpdateGroupInviteRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
	MaxUses   *int       `json:"max_uses" binding:"omitempty,gt=0"`
}

type JoinGroupByInviteRequest struct {
	Password string `json:"password" binding:"required,min=1,max=72"`
}

type GroupInviteResponse struct {
	ID        int64      `json:"id"`
	GroupID   int64      `json:"group_id"`
	Token     string     `json:"token,omitempty"`
	InviteURL string     `json:"invite_url,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	MaxUses   *int       `json:"max_uses,omitempty"`
	UsedCount int        `json:"used_count"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
