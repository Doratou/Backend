package handler

import (
	"log"
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type InviteHandler struct {
	service adapter.InviteService
}

func NewInviteHandler(service adapter.InviteService) *InviteHandler {
	return &InviteHandler{service: service}
}

func (h *InviteHandler) SendDirect(c *gin.Context) {
	userID, role, ok := getUserIdentity(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	var req dto.CreateDirectInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.ErrDirectInviteInvalidPayload)
		return
	}
	items, err := h.service.SendDirect(c.Request.Context(), userID, role, req)
	if err != nil {
		failLogged(c, err)
		return
	}
	for _, item := range items {
		if item.Error != nil {
			log.Printf("[INVITE RESULT] route=%s user_id=%d group_id=%d recipient_user_id=%d code=%s message=%q",
				c.FullPath(), userID, req.GroupID, item.UserID, item.Error.Code, item.Error.Message)
		}
	}
	success(c, http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *InviteHandler) RespondToDirect(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	inviteID, err := parseIDParam(c, "inviteID")
	if err != nil {
		failLogged(c, exception.ErrInviteInvalidID)
		return
	}
	var req dto.RespondDirectInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.ErrDirectInviteInvalidResponse)
		return
	}
	invite, err := h.service.RespondToDirect(c.Request.Context(), userID, inviteID, req.Response)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, invite)
}

func (h *InviteHandler) ListMyDirectInvites(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	items, err := h.service.ListMyDirectInvites(c.Request.Context(), userID)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *InviteHandler) ListGroupDirectInvites(c *gin.Context) {
	userID, role, ok := getUserIdentity(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil {
		failLogged(c, exception.ErrGroupInvalidID)
		return
	}
	items, err := h.service.ListGroupDirectInvites(c.Request.Context(), userID, role, groupID)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *InviteHandler) CreateLink(c *gin.Context) {
	userID, role, ok := getUserIdentity(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil {
		failLogged(c, err)
		return
	}
	var req dto.CreateLinkInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.Wrap(exception.ErrInviteInvalidPayload, err))
		return
	}
	invite, err := h.service.CreateLink(c.Request.Context(), userID, role, groupID, req)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusCreated, invite)
}

func (h *InviteHandler) ListGroupLinks(c *gin.Context) {
	userID, role, ok := getUserIdentity(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil {
		failLogged(c, err)
		return
	}
	invites, err := h.service.ListGroupLinks(c.Request.Context(), userID, role, groupID)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"items": invites, "total": len(invites)})
}

func (h *InviteHandler) DeactivateLink(c *gin.Context) {
	userID, role, groupID, inviteID, ok := h.getLinkRequestIdentity(c)
	if !ok {
		return
	}
	if err := h.service.DeactivateLink(c.Request.Context(), userID, role, groupID, inviteID); err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "group invite deactivated successfully"})
}

func (h *InviteHandler) ChangeLinkPassword(c *gin.Context) {
	userID, role, groupID, inviteID, ok := h.getLinkRequestIdentity(c)
	if !ok {
		return
	}
	var req dto.ChangeLinkInvitePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.Wrap(exception.ErrInviteInvalidPasswordFormat, err))
		return
	}
	if err := h.service.ChangeLinkPassword(c.Request.Context(), userID, role, groupID, inviteID, req.Password); err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "group invite password changed successfully"})
}

func (h *InviteHandler) UpdateLink(c *gin.Context) {
	userID, role, groupID, inviteID, ok := h.getLinkRequestIdentity(c)
	if !ok {
		return
	}
	var req dto.UpdateLinkInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.Wrap(exception.ErrInviteInvalidPayload, err))
		return
	}
	invite, err := h.service.UpdateLink(c.Request.Context(), userID, role, groupID, inviteID, req)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, invite)
}

func (h *InviteHandler) JoinByLink(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	var req dto.JoinByLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.Wrap(exception.ErrInviteJoinPayload, err))
		return
	}
	group, err := h.service.JoinByLink(c.Request.Context(), userID, c.Param("token"), req.Password)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, group)
}

func (h *InviteHandler) getLinkRequestIdentity(c *gin.Context) (int64, enum.UserRole, int64, int64, bool) {
	userID, role, ok := getUserIdentity(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return 0, "", 0, 0, false
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil {
		failLogged(c, err)
		return 0, "", 0, 0, false
	}
	inviteID, err := parseIDParam(c, "inviteID")
	if err != nil {
		failLogged(c, err)
		return 0, "", 0, 0, false
	}
	return userID, role, groupID, inviteID, true
}
