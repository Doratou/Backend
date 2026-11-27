package dto

import "github.com/Doratou/Backend/internal/domain/enum"

type UserPreviewResponse struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	AvatarAttachmentID *int64  `json:"avatar_attachment_id"`
	AvatarURL          *string `json:"avatar_url"`
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type ChangeEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ChangePhoneRequest struct {
	Phone string `json:"phone" binding:"required,min=10,max=20"`
}

type ProfileCompletionSteps struct {
	FullName          bool `json:"full_name"`
	ProfilePhoto      bool `json:"profile_photo"`
	VerifiedEmail     bool `json:"verified_email"`
	VerifiedPhone     bool `json:"verified_phone"`
	TelegramConnected bool `json:"telegram_connected"`
}

type ProfileCompletionResponse struct {
	Steps          ProfileCompletionSteps `json:"steps"`
	CompletedSteps int                    `json:"completed_steps"`
	TotalSteps     int                    `json:"total_steps"`
	Progress       int                    `json:"progress_percent"`
	RewardTokens   int64                  `json:"reward_tokens"`
	RewardGranted  bool                   `json:"reward_granted"`
}

type ChangeUserRoleRequest struct {
	Role enum.UserRole `json:"role" binding:"required,oneof=USER ADMIN"`
}

type AdminUserFilterRequest struct {
	Status *enum.UserStatus `form:"status"`
	Role   *enum.UserRole   `form:"role"`
	Search *string          `form:"search"`
}
