package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type FriendshipService interface {
	SendRequest(ctx context.Context, userID int64, req dto.CreateFriendshipRequest) (*dto.FriendRequestResponse, error)
	ListRequests(ctx context.Context, userID int64, req dto.ListFriendRequestsRequest) (*dto.FriendRequestListResponse, error)
	AcceptRequest(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error)
	RejectRequest(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error)
	CancelRequest(ctx context.Context, userID, requestID int64) error
	ListFriends(ctx context.Context, userID int64, req dto.ListFriendsRequest) (*dto.FriendListResponse, error)
	RemoveFriend(ctx context.Context, userID, friendshipID int64) error
}
