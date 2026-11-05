package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	pollService adapter.PollService
}

func NewPollHandler(pollService adapter.PollService) *PollHandler {
	return &PollHandler{pollService: pollService}
}

func (h *PollHandler) Create(c *gin.Context) {
	userID, groupID, ok := h.getPollGroupIdentity(c)
	if !ok {
		return
	}
	var req dto.CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.pollService.Create(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusCreated, res)
}

func (h *PollHandler) List(c *gin.Context) {
	userID, groupID, ok := h.getPollGroupIdentity(c)
	if !ok {
		return
	}
	var req dto.ListPollsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.pollService.List(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *PollHandler) Get(c *gin.Context) {
	userID, groupID, pollID, ok := h.getPollRequestIdentity(c)
	if !ok {
		return
	}
	res, err := h.pollService.Get(c.Request.Context(), userID, groupID, pollID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *PollHandler) Vote(c *gin.Context) {
	userID, groupID, pollID, ok := h.getPollRequestIdentity(c)
	if !ok {
		return
	}
	var req dto.VotePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.pollService.Vote(c.Request.Context(), userID, groupID, pollID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *PollHandler) WithdrawVote(c *gin.Context) {
	userID, groupID, pollID, ok := h.getPollRequestIdentity(c)
	if !ok {
		return
	}
	if err := h.pollService.WithdrawVote(c.Request.Context(), userID, groupID, pollID); err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "poll vote withdrawn successfully"})
}

func (h *PollHandler) ListVoters(c *gin.Context) {
	userID, groupID, pollID, ok := h.getPollRequestIdentity(c)
	if !ok {
		return
	}
	var req dto.ListPollVotersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.pollService.ListVoters(c.Request.Context(), userID, groupID, pollID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *PollHandler) Close(c *gin.Context) {
	userID, groupID, pollID, ok := h.getPollRequestIdentity(c)
	if !ok {
		return
	}
	res, err := h.pollService.Close(c.Request.Context(), userID, groupID, pollID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *PollHandler) Delete(c *gin.Context) {
	userID, groupID, pollID, ok := h.getPollRequestIdentity(c)
	if !ok {
		return
	}
	if err := h.pollService.Delete(c.Request.Context(), userID, groupID, pollID); err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "poll deleted successfully"})
}

func (h *PollHandler) getPollGroupIdentity(c *gin.Context) (int64, int64, bool) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return 0, 0, false
	}
	if userID <= 0 {
		fail(c, exception.ErrTokenInvalid)
		return 0, 0, false
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil || groupID <= 0 {
		fail(c, exception.ErrBadRequest)
		return 0, 0, false
	}
	return userID, groupID, true
}

func (h *PollHandler) getPollRequestIdentity(c *gin.Context) (int64, int64, int64, bool) {
	userID, groupID, ok := h.getPollGroupIdentity(c)
	if !ok {
		return 0, 0, 0, false
	}
	pollID, err := parseIDParam(c, "pollID")
	if err != nil || pollID <= 0 {
		fail(c, exception.ErrBadRequest)
		return 0, 0, 0, false
	}
	return userID, groupID, pollID, true
}
