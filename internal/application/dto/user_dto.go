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

type WalletResponse struct {
	Wallet  int64  `json:"wallet"`
	Message string `json:"message,omitempty"`
}

type ChangeUserRoleRequest struct {
	Role enum.UserRole `json:"role" binding:"required,oneof=USER ADMIN"`
}

type AdminUserFilterRequest struct {
	Status *enum.UserStatus `form:"status"`
	Role   *enum.UserRole   `form:"role"`
	Search *string          `form:"search"`
}
