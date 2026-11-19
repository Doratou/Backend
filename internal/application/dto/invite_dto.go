package dto

import "time"

type CreateDirectInviteRequest struct {
	GroupID int64   `json:"group_id" binding:"required,gt=0"`
	UserIDs []int64 `json:"user_ids" binding:"required,min=1,max=100,dive,gt=0"`
}

type RespondDirectInviteRequest struct {
	Response string `json:"response" binding:"required,oneof=ACCEPTED REJECTED"`
}

type DirectInviteSendResult struct {
	UserID  int64                 `json:"user_id"`
	Success bool                  `json:"success"`
	Invite  *DirectInviteResponse `json:"invite,omitempty"`
	Error   *DirectInviteError    `json:"error,omitempty"`
}

type DirectInviteError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DirectInviteResponse struct {
	ID          int64              `json:"id"`
	Status      string             `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	RespondedAt *time.Time         `json:"responded_at,omitempty"`
	Group       *DirectInviteGroup `json:"group,omitempty"`
	Sender      *DirectInviteUser  `json:"sender,omitempty"`
	Recipient   *DirectInviteUser  `json:"recipient,omitempty"`
}

type DirectInviteGroup struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type DirectInviteUser struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	AvatarAttachmentID *int64 `json:"avatar_attachment_id,omitempty"`
}

type CreateLinkInviteRequest struct {
	Password  string     `json:"password" binding:"required,min=8,max=72"`
	ExpiresAt *time.Time `json:"expires_at"`
	MaxUses   *int       `json:"max_uses" binding:"omitempty,gt=0"`
}

type ChangeLinkInvitePasswordRequest struct {
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type UpdateLinkInviteRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
	MaxUses   *int       `json:"max_uses" binding:"omitempty,gt=0"`
}

type JoinByLinkRequest struct {
	Password string `json:"password" binding:"required,min=1,max=72"`
}

type LinkInviteResponse struct {
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
