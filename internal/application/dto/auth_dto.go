package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type SendOTPRequest struct {
	Identifier string       `json:"identifier" binding:"required"`
	Type       enum.OTPType `json:"type" binding:"required,oneof=email phone"`
}

type SendOTPResponse struct {
	ExpiresInSeconds int64  `json:"expires_in_seconds"`
	Message          string `json:"message"`
}

type VerifyOTPRequest struct {
	Identifier string       `json:"identifier" binding:"required"`
	Type       enum.OTPType `json:"type" binding:"required,oneof=email phone"`
	Code       string       `json:"code" binding:"required,min=4,max=8"`
}

type VerifyOTPResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type RegisterRequest struct {
	Name     string  `json:"name" binding:"omitempty,min=2,max=100"`
	Password string  `json:"password" binding:"required,min=6,max=100"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type LoginWithEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LoginWithPhoneRequest struct {
	Phone string `json:"phone" binding:"required,min=10,max=20"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=100"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

type ResetPasswordRequest struct {
	Identifier  string       `json:"identifier" binding:"required"`
	Type        enum.OTPType `json:"type" binding:"required,oneof=email phone"`
	NewPassword string       `json:"new_password" binding:"required,min=6,max=100"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenPairResponse struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type UserResponse struct {
	ID                 int64           `json:"id"`
	Name               string          `json:"name"`
	Email              *string         `json:"email,omitempty"`
	Phone              *string         `json:"phone,omitempty"`
	IsEmailVerified    bool            `json:"is_email_verified"`
	IsPhoneVerified    bool            `json:"is_phone_verified"`
	AvatarAttachmentID *int64          `json:"avatar_attachment_id"`
	SystemRole         enum.UserRole   `json:"system_role"`
	Status             enum.UserStatus `json:"status"`
	Wallet             int64           `json:"wallet"`
	TelegramChatID     *int64          `json:"telegram_chat_id,omitempty"`
	TelegramUserID     *int64          `json:"telegram_user_id,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
}

type AuthResponse struct {
	User                  UserResponse `json:"user"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	ExpiresAt             time.Time    `json:"expires_at"`
}

// ToUserResponse converts a domain User entity into a UserResponse DTO.
func ToUserResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:                 user.ID,
		Name:               user.Name,
		Email:              user.Email,
		Phone:              user.Phone,
		IsEmailVerified:    user.IsEmailVerified,
		IsPhoneVerified:    user.IsPhoneVerified,
		AvatarAttachmentID: user.AvatarAttachmentID,
		SystemRole:         user.SystemRole,
		Status:             user.Status,
		Wallet:             user.Wallet,
		TelegramChatID:     user.TelegramChatID,
		TelegramUserID:     user.TelegramUserID,
		CreatedAt:          user.CreatedAt,
	}
}
