package handler

import (
	"net/http"
	"strconv"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service adapter.NotificationService
}

func NewNotificationHandler(service adapter.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	var query dto.GetNotificationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.service.GetNotifications(c.Request.Context(), userID, query)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	var groupID *int64
	if gidStr := c.Query("group_id"); gidStr != "" {
		if gid, err := strconv.ParseInt(gidStr, 10, 64); err == nil && gid > 0 {
			groupID = &gid
		}
	}

	res, err := h.service.GetUnreadCount(c.Request.Context(), userID, groupID)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

func (h *NotificationHandler) SyncNotifications(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	var query dto.SyncNotificationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.service.SyncNotifications(c.Request.Context(), userID, query)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil || id <= 0 {
		fail(c, exception.ErrBadRequest)
		return
	}

	res, err := h.service.GetNotificationByID(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.MarkAsReadRequest
	_ = c.ShouldBindJSON(&req)

	res, err := h.service.MarkAsRead(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

func (h *NotificationHandler) MarkSingleAsRead(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil || id <= 0 {
		fail(c, exception.ErrBadRequest)
		return
	}

	res, err := h.service.MarkSingleAsRead(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil || id <= 0 {
		fail(c, exception.ErrBadRequest)
		return
	}

	if err := h.service.DeleteNotification(c.Request.Context(), userID, id); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *NotificationHandler) ClearNotifications(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	var query dto.ClearNotificationsQuery
	_ = c.ShouldBindQuery(&query)

	if err := h.service.ClearNotifications(c.Request.Context(), userID, query.OnlyRead); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{"cleared": true})
}

func (h *NotificationHandler) CreateGroupAnnouncement(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok || userID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	groupID, err := parseIDParam(c, "groupID")
	if err != nil || groupID <= 0 {
		fail(c, exception.ErrBadRequest)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var req dto.GroupAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.service.CreateGroupAnnouncement(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusCreated, res)
}
