package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService adapter.UserService
}

func NewUserHandler(userService adapter.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	response, err := h.userService.GetMe(c.Request.Context(), userID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

func (h *UserHandler) FindByFriendCode(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	response, err := h.userService.FindByFriendCode(c.Request.Context(), userID, c.Param("friendCode"))
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	targetUserID, err := parseIDParam(c, "userID")
	if err != nil {
		fail(c, err)
		return
	}
	response, err := h.userService.GetProfile(c.Request.Context(), userID, targetUserID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}
