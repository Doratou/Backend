package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type UserService interface {
	GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
	ChangeEmail(ctx context.Context, userID int64, req dto.ChangeEmailRequest) (*dto.UserResponse, error)
	ChangePhone(ctx context.Context, userID int64, req dto.ChangePhoneRequest) (*dto.UserResponse, error)
	DeleteAccount(ctx context.Context, userID int64, tokenString string) error
	GetWallet(ctx context.Context, userID int64) (*dto.WalletResponse, error)
	TopUpWallet(ctx context.Context, userID int64, req dto.TopUpWalletRequest) (*dto.WalletResponse, error)
	GetUserByID(ctx context.Context, currentUserID, targetUserID int64) (*dto.UserPreviewResponse, error)
	FindByFriendCode(ctx context.Context, userID int64, friendCode string) (*dto.UserPreviewResponse, error)
	RegenerateFriendCode(ctx context.Context, userID int64) (*dto.UserResponse, error)

	// Admin operations
	AdminListUsers(ctx context.Context, filter dto.AdminUserFilterRequest, page, limit, offset int) (*dto.PaginatedResponse[dto.UserResponse], error)
	AdminGetUserByID(ctx context.Context, targetUserID int64) (*dto.UserResponse, error)
	AdminUpdateUserRole(ctx context.Context, adminUserID, targetUserID int64, req dto.ChangeUserRoleRequest) (*dto.UserResponse, error)
}
