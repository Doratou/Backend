package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type SettlementHandler struct{ settlementService adapter.SettlementService }

func NewSettlementHandler(settlementService adapter.SettlementService) *SettlementHandler {
	return &SettlementHandler{settlementService: settlementService}
}

func (h *SettlementHandler) Create(c *gin.Context) {
	userID, groupID, ok := h.groupIdentity(c)
	if !ok {
		return
	}
	var req dto.CreateSettlementRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.settlementService.Create(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusCreated, res)
}

func (h *SettlementHandler) List(c *gin.Context) {
	userID, groupID, ok := h.groupIdentity(c)
	if !ok {
		return
	}
	var req dto.ListSettlementsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.settlementService.List(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *SettlementHandler) Get(c *gin.Context) {
	userID, groupID, settlementID, ok := h.settlementIdentity(c)
	if !ok {
		return
	}
	res, err := h.settlementService.Get(c.Request.Context(), userID, groupID, settlementID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *SettlementHandler) Delete(c *gin.Context) {
	userID, groupID, settlementID, ok := h.settlementIdentity(c)
	if !ok {
		return
	}
	if err := h.settlementService.Delete(c.Request.Context(), userID, groupID, settlementID); err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "settlement deleted successfully"})
}

func (h *SettlementHandler) Confirm(c *gin.Context) {
	userID, groupID, id, ok := h.settlementIdentity(c)
	if !ok {
		return
	}
	res, err := h.settlementService.Confirm(c.Request.Context(), userID, groupID, id)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *SettlementHandler) Reject(c *gin.Context) {
	userID, groupID, id, ok := h.settlementIdentity(c)
	if !ok {
		return
	}
	res, err := h.settlementService.Reject(c.Request.Context(), userID, groupID, id)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *SettlementHandler) groupIdentity(c *gin.Context) (int64, int64, bool) {
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

func (h *SettlementHandler) settlementIdentity(c *gin.Context) (int64, int64, int64, bool) {
	userID, groupID, ok := h.groupIdentity(c)
	if !ok {
		return 0, 0, 0, false
	}
	settlementID, err := parseIDParam(c, "settlementID")
	if err != nil || settlementID <= 0 {
		fail(c, exception.ErrBadRequest)
		return 0, 0, 0, false
	}
	return userID, groupID, settlementID, true
}
