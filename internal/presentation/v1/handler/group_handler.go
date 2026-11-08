package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type GroupHandler struct{ service adapter.GroupService }

func NewGroupHandler(service adapter.GroupService) *GroupHandler {
	return &GroupHandler{service: service}
}

func (h *GroupHandler) Create(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.Wrap(exception.ErrGroupInvalidPayload, err))
		return
	}
	result, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusCreated, result)
}

func (h *GroupHandler) ListMine(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	page, limit, offset, err := parsePagination(c)
	if err != nil {
		failLogged(c, exception.ErrGroupInvalidPagination)
		return
	}
	var status *string
	if value, exists := c.GetQuery("status"); exists {
		status = &value
	}
	groups, total, err := h.service.ListMine(c.Request.Context(), userID, status, limit, offset)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, dto.PaginatedResponse[dto.GroupResponse]{
		Items: groups, Pagination: dto.PaginationResponse{Page: page, Limit: limit, Total: total},
	})
}

func (h *GroupHandler) ListAll(c *gin.Context) {
	role, ok := tokencontext.GetUserRole(c)
	if !ok || role != enum.UserRoleAdmin {
		failLogged(c, exception.ErrGroupForbidden)
		return
	}
	page, limit, offset, err := parsePagination(c)
	if err != nil {
		failLogged(c, exception.ErrGroupInvalidPagination)
		return
	}
	var status *string
	if value, exists := c.GetQuery("status"); exists {
		status = &value
	}
	groups, total, err := h.service.ListAll(c.Request.Context(), status, limit, offset)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, dto.PaginatedResponse[dto.GroupResponse]{
		Items: groups, Pagination: dto.PaginationResponse{Page: page, Limit: limit, Total: total},
	})
}

func (h *GroupHandler) Update(c *gin.Context) {
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
	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.Wrap(exception.ErrGroupInvalidUpdatePayload, err))
		return
	}
	result, err := h.service.Update(c.Request.Context(), userID, role, groupID, req)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, result)
}

func (h *GroupHandler) ChangeCover(c *gin.Context) {
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
	if err := h.service.ChangeCover(c.Request.Context(), userID, role, groupID); err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "group cover update is not implemented yet"})
}

func (h *GroupHandler) Archive(c *gin.Context) {
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
	if err := h.service.Archive(c.Request.Context(), userID, role, groupID); err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "group archived successfully"})
}

func (h *GroupHandler) Delete(c *gin.Context) {
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
	if err := h.service.Delete(c.Request.Context(), userID, role, groupID); err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "group deleted successfully"})
}

func getUserIdentity(c *gin.Context) (int64, enum.UserRole, bool) {
	userID, userOK := tokencontext.GetUserID(c)
	role, roleOK := tokencontext.GetUserRole(c)
	return userID, role, userOK && roleOK
}

func (h *GroupHandler) Clone(c *gin.Context) {
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
	var req dto.CloneGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.Wrap(exception.ErrGroupInvalidPayload, err))
		return
	}
	result, err := h.service.Clone(c.Request.Context(), userID, role, groupID, req)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusCreated, result)
}

func (h *GroupHandler) Get(c *gin.Context) {
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
	result, err := h.service.Get(c.Request.Context(), userID, role, groupID)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, result)
}

func (h *GroupHandler) ListMembers(c *gin.Context) {
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
	members, err := h.service.ListMembers(c.Request.Context(), userID, role, groupID)
	if err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, members)
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
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
	memberID, err := parseIDParam(c, "memberID")
	if err != nil {
		failLogged(c, err)
		return
	}
	if err := h.service.RemoveMember(c.Request.Context(), userID, role, groupID, memberID); err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "group member removed successfully"})
}

func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		failLogged(c, exception.ErrMissingToken)
		return
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil {
		failLogged(c, err)
		return
	}
	if err := h.service.LeaveGroup(c.Request.Context(), userID, groupID); err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "you left the group successfully"})
}

func (h *GroupHandler) TransferOwnership(c *gin.Context) {
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
	var req dto.TransferGroupOwnershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failLogged(c, exception.Wrap(exception.ErrGroupInvalidTransferOwner, err))
		return
	}
	if err := h.service.TransferOwnership(c.Request.Context(), userID, role, groupID, req.NewOwnerID); err != nil {
		failLogged(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "group ownership transferred successfully"})
}
