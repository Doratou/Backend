package handler

import (
	"net/http"
	"strings"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
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

// GetMe handles GET /api/v1/users/me
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

// UpdateMe handles PATCH /api/v1/users/me
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	response, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

// ChangeEmail handles PATCH /api/v1/users/me/email
func (h *UserHandler) ChangeEmail(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	response, err := h.userService.ChangeEmail(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

// ChangePhone handles PATCH /api/v1/users/me/phone
func (h *UserHandler) ChangePhone(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.ChangePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	response, err := h.userService.ChangePhone(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

// ActivateUserStatus handles POST /api/v1/admin/users/status/activate/:id
func (h *UserHandler) ActivateUserStatus(c *gin.Context) {
	targetUserID, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	response, err := h.userService.ActivateUserStatus(c.Request.Context(), targetUserID)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, response)
}

// DeleteMe handles DELETE /api/v1/users/me
func (h *UserHandler) DeleteMe(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	authHeader := c.GetHeader("Authorization")
	fields := strings.Fields(authHeader)
	var tokenString string
	if len(fields) == 2 && strings.EqualFold(fields[0], "bearer") {
		tokenString = fields[1]
	}

	if err := h.userService.DeleteAccount(c.Request.Context(), userID, tokenString); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{"message": "account deleted successfully"})
}

// GetWallet handles GET /api/v1/users/me/wallet
func (h *UserHandler) GetWallet(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	response, err := h.userService.GetWallet(c.Request.Context(), userID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

// GetUserByID handles GET /api/v1/users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	currentUserID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	targetUserID, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	response, err := h.userService.GetUserByID(c.Request.Context(), currentUserID, targetUserID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

// FindByFriendCode handles GET /api/v1/users/by-friend-code/:friendCode
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

// RegenerateFriendCode handles POST /api/v1/users/me/friend-code/regenerate
func (h *UserHandler) RegenerateFriendCode(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}
	response, err := h.userService.RegenerateFriendCode(c.Request.Context(), userID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

// ==========================================
// Admin Endpoints
// ==========================================

// AdminListUsers handles GET /api/v1/admin/users
func (h *UserHandler) AdminListUsers(c *gin.Context) {
	page, limit, offset, err := parsePagination(c)
	if err != nil {
		fail(c, err)
		return
	}

	var filter dto.AdminUserFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.userService.AdminListUsers(c.Request.Context(), filter, page, limit, offset)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

// AdminGetUserByID handles GET /api/v1/admin/users/:id
func (h *UserHandler) AdminGetUserByID(c *gin.Context) {
	targetUserID, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	res, err := h.userService.AdminGetUserByID(c.Request.Context(), targetUserID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

// AdminUpdateUserRole handles PATCH /api/v1/admin/users/:id/role
func (h *UserHandler) AdminUpdateUserRole(c *gin.Context) {
	adminUserID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	targetUserID, err := parseIDParam(c, "id")
	if err != nil {
		fail(c, err)
		return
	}

	var req dto.ChangeUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.userService.AdminUpdateUserRole(c.Request.Context(), adminUserID, targetUserID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}
