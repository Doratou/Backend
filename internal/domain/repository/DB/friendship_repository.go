package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type FriendshipRepository interface {
	Create(ctx context.Context, friendship *entity.Friendship) error
	ReopenRequest(ctx context.Context, friendship *entity.Friendship) error
	FindByID(ctx context.Context, id int64) (*entity.Friendship, error)
	FindByUserPair(ctx context.Context, firstUserID, secondUserID int64) (*entity.Friendship, error)
	FindIncomingRequests(ctx context.Context, userID int64, limit, offset int) ([]*entity.Friendship, int64, error)
	FindOutgoingRequests(ctx context.Context, userID int64, limit, offset int) ([]*entity.Friendship, int64, error)
	FindFriends(ctx context.Context, userID int64, limit, offset int) ([]*entity.Friendship, int64, error)
	Update(ctx context.Context, friendship *entity.Friendship, expectedStatus enum.FriendshipStatus) error
	DeleteByID(ctx context.Context, id int64, expectedStatus enum.FriendshipStatus) error
}
