package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type FriendshipHandler struct {
	friendshipService adapter.FriendshipService
}

func NewFriendshipHandler(friendshipService adapter.FriendshipService) *FriendshipHandler {
	return &FriendshipHandler{
		friendshipService: friendshipService,
	}
}

func (h *FriendshipHandler) SendRequest(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.CreateFriendshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.friendshipService.SendRequest(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusCreated, res)
}

func (h *FriendshipHandler) ListRequests(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.ListFriendRequestsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.friendshipService.ListRequests(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *FriendshipHandler) AcceptRequest(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	requestID, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	res, err := h.friendshipService.AcceptRequest(c.Request.Context(), userID, requestID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *FriendshipHandler) RejectRequest(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	requestID, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	res, err := h.friendshipService.RejectRequest(c.Request.Context(), userID, requestID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *FriendshipHandler) CancelRequest(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	requestID, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	if err := h.friendshipService.CancelRequest(c.Request.Context(), userID, requestID); err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{
		"message": "friend request canceled successfully",
	})
}

func (h *FriendshipHandler) ListFriends(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.ListFriendsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.friendshipService.ListFriends(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *FriendshipHandler) RemoveFriend(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	friendshipID, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	if err := h.friendshipService.RemoveFriend(c.Request.Context(), userID, friendshipID); err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{
		"message": "friend removed successfully",
	})
}
