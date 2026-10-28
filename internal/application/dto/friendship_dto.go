package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type CreateFriendshipRequest struct {
	ReceiverUserID int64 `json:"receiver_user_id" binding:"required,gt=0"`
}

type ListFriendRequestsRequest struct {
	Direction string `form:"direction" binding:"required,oneof=incoming outgoing"`
	Page      int    `form:"page,default=1" binding:"min=1"`
	Limit     int    `form:"limit,default=20" binding:"min=1,max=100"`
}

type ListFriendsRequest struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}

type FriendRequestResponse struct {
	ID              int64                 `json:"id"`
	RequesterUserID int64                 `json:"requester_user_id"`
	ReceiverUserID  int64                 `json:"receiver_user_id"`
	Status          enum.FriendshipStatus `json:"status"`
	CreatedAt       time.Time             `json:"created_at"`
	RespondedAt     *time.Time            `json:"responded_at"`
}

type FriendRequestStatusResponse struct {
	ID          int64                 `json:"id"`
	Status      enum.FriendshipStatus `json:"status"`
	RespondedAt *time.Time            `json:"responded_at"`
}

type FriendshipUserResponse struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	AvatarAttachmentID *int64 `json:"avatar_attachment_id"`
}

type FriendRequestListItem struct {
	ID        int64                  `json:"id"`
	Requester FriendshipUserResponse `json:"requester"`
	Receiver  FriendshipUserResponse `json:"receiver"`
	Status    enum.FriendshipStatus  `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
}

type FriendshipPaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type FriendRequestListResponse struct {
	Items      []FriendRequestListItem      `json:"items"`
	Pagination FriendshipPaginationResponse `json:"pagination"`
}

type FriendResponse struct {
	FriendshipID int64                  `json:"friendship_id"`
	User         FriendshipUserResponse `json:"user"`
	FriendsSince *time.Time             `json:"friends_since"`
}

type FriendListResponse struct {
	Items      []FriendResponse             `json:"items"`
	Pagination FriendshipPaginationResponse `json:"pagination"`
}

func ToFriendRequestResponse(friendship *entity.Friendship) FriendRequestResponse {
	return FriendRequestResponse{
		ID:              friendship.ID,
		RequesterUserID: friendship.RequesterUserID,
		ReceiverUserID:  friendship.ReceiverUserID,
		Status:          friendship.Status,
		CreatedAt:       friendship.CreatedAt,
		RespondedAt:     friendship.RespondedAt,
	}
}

func ToFriendRequestStatusResponse(friendship *entity.Friendship) FriendRequestStatusResponse {
	return FriendRequestStatusResponse{
		ID:          friendship.ID,
		Status:      friendship.Status,
		RespondedAt: friendship.RespondedAt,
	}
}

func ToFriendRequestListItem(friendship *entity.Friendship) FriendRequestListItem {
	return FriendRequestListItem{
		ID:        friendship.ID,
		Requester: toFriendshipUserResponse(friendship.RequesterUserID, friendship.RequesterUser),
		Receiver:  toFriendshipUserResponse(friendship.ReceiverUserID, friendship.ReceiverUser),
		Status:    friendship.Status,
		CreatedAt: friendship.CreatedAt,
	}
}

func ToFriendResponse(friendship *entity.Friendship, userID int64) FriendResponse {
	friendID, friend := friendship.RequesterUserID, friendship.RequesterUser
	if friendship.RequesterUserID == userID {
		friendID, friend = friendship.ReceiverUserID, friendship.ReceiverUser
	}
	return FriendResponse{
		FriendshipID: friendship.ID,
		User:         toFriendshipUserResponse(friendID, friend),
		FriendsSince: friendship.RespondedAt,
	}
}

func toFriendshipUserResponse(userID int64, user *entity.User) FriendshipUserResponse {
	response := FriendshipUserResponse{ID: userID}
	if user != nil {
		response.Name = user.Name
		response.AvatarAttachmentID = user.AvatarAttachmentID
	}
	return response
}
