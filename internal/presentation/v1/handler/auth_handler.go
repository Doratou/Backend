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

type AuthHandler struct {
	authService adapter.AuthService
}

func NewAuthHandler(authService adapter.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SendOTP handles POST /api/v1/auth/otp/send
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req dto.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.authService.SendOTP(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// VerifyOTP handles POST /api/v1/auth/otp/verify
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.authService.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusCreated, res)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// LoginWithEmail handles POST /api/v1/auth/login/email
func (h *AuthHandler) LoginWithEmail(c *gin.Context) {
	var req dto.LoginWithEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.authService.LoginWithEmail(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// LoginWithPhone handles POST /api/v1/auth/login/phone
func (h *AuthHandler) LoginWithPhone(c *gin.Context) {
	var req dto.LoginWithPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.authService.LoginWithPhone(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// ChangePassword handles POST /api/v1/auth/password/change (protected)
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{
		"message": "password changed successfully",
	})
}

// ResetPassword handles POST /api/v1/auth/password/reset (public, requires prior OTP verification)
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), req); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{
		"message": "password reset successfully",
	})
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.authService.Refresh(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	fields := strings.Fields(header)
	if len(fields) != 2 {
		fail(c, exception.ErrMissingToken)
		return
	}

	if err := h.authService.Logout(c.Request.Context(), fields[1]); err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}
