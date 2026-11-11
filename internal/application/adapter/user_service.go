package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type UserService interface {
	GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error)
	FindByFriendCode(ctx context.Context, userID int64, friendCode string) (*dto.UserPreviewResponse, error)
}
