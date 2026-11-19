package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type ChecklistHandler struct {
	service adapter.ChecklistService
}

func NewChecklistHandler(service adapter.ChecklistService) *ChecklistHandler {
	return &ChecklistHandler{service: service}
}

func (h *ChecklistHandler) Get(c *gin.Context) {
	userID, groupID, ok := checklistGroupIdentity(c)
	if !ok {
		return
	}
	var req dto.GetChecklistRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.service.Get(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *ChecklistHandler) CreateItem(c *gin.Context) {
	userID, groupID, ok := checklistGroupIdentity(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8<<10)
	var req dto.CreateChecklistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.service.CreateItem(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusCreated, res)
}

func (h *ChecklistHandler) ClaimItem(c *gin.Context) {
	h.setClaim(c, true)
}

func (h *ChecklistHandler) ReleaseItem(c *gin.Context) {
	h.setClaim(c, false)
}

func (h *ChecklistHandler) setClaim(c *gin.Context, claimed bool) {
	userID, groupID, itemID, ok := checklistItemIdentity(c)
	if !ok {
		return
	}
	res, err := h.service.SetClaim(c.Request.Context(), userID, groupID, itemID, claimed)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *ChecklistHandler) DeleteItem(c *gin.Context) {
	userID, groupID, itemID, ok := checklistItemIdentity(c)
	if !ok {
		return
	}
	if err := h.service.DeleteItem(c.Request.Context(), userID, groupID, itemID); err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "checklist item deleted successfully"})
}

func checklistGroupIdentity(c *gin.Context) (int64, int64, bool) {
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

func checklistItemIdentity(c *gin.Context) (int64, int64, int64, bool) {
	userID, groupID, ok := checklistGroupIdentity(c)
	if !ok {
		return 0, 0, 0, false
	}
	itemID, err := parseIDParam(c, "itemID")
	if err != nil || itemID <= 0 {
		fail(c, exception.ErrBadRequest)
		return 0, 0, 0, false
	}
	return userID, groupID, itemID, true
}
