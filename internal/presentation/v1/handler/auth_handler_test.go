package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/Doratou/Backend/internal/presentation/v1/router"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) SendOTP(ctx context.Context, req dto.SendOTPRequest) (*dto.SendOTPResponse, error) {
	args := m.Called(ctx, req)
	if res := args.Get(0); res != nil {
		return res.(*dto.SendOTPResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error) {
	args := m.Called(ctx, req)
	if res := args.Get(0); res != nil {
		return res.(*dto.VerifyOTPResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if res := args.Get(0); res != nil {
		return res.(*dto.AuthResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if res := args.Get(0); res != nil {
		return res.(*dto.AuthResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) LoginWithEmail(ctx context.Context, req dto.LoginWithEmailRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if res := args.Get(0); res != nil {
		return res.(*dto.AuthResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) LoginWithPhone(ctx context.Context, req dto.LoginWithPhoneRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if res := args.Get(0); res != nil {
		return res.(*dto.AuthResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) ChangePassword(ctx context.Context, userID int64, req dto.ChangePasswordRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *mockAuthService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockAuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.TokenPairResponse, error) {
	args := m.Called(ctx, req)
	if res := args.Get(0); res != nil {
		return res.(*dto.TokenPairResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func setupTestServer(authSvc *mockAuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handler.NewAuthHandler(authSvc)
	return router.New(router.Dependencies{
		AuthHandler: h,
	})
}

func setupProtectedTestServer(authSvc *mockAuthService, mockUserID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if mockUserID > 0 {
			tokencontext.SetClaims(c, &jwt.Claims{
				UserID:    mockUserID,
				Role:      enum.UserRoleUser,
				TokenType: jwt.TokenTypeAccess,
			})
		}
		c.Next()
	})
	h := handler.NewAuthHandler(authSvc)
	router.RegisterAuthRoutes(r.Group("/api/v1"), h, nil)
	return r
}

func TestAuthHandler_SendOTP(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.SendOTPRequest{
			Identifier: "test@example.com",
			Type:       enum.OTPTypeEmail,
		}
		expectedRes := &dto.SendOTPResponse{
			ExpiresInSeconds: 180,
			Message:          "verification code sent to test@example.com",
		}

		mockSvc.On("SendOTP", mock.Anything, reqPayload).Return(expectedRes, nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/otp/send", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		mockSvc.AssertExpectations(t)
	})

	t.Run("bad request validation failure", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		body := []byte(`{"identifier": ""}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/otp/send", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp["success"].(bool))
		assert.Equal(t, "BAD_REQUEST", resp["code"])
	})
}

func TestAuthHandler_VerifyOTP(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.VerifyOTPRequest{
			Identifier: "test@example.com",
			Type:       enum.OTPTypeEmail,
			Code:       "12345",
		}
		expectedRes := &dto.VerifyOTPResponse{
			Status:  "verified",
			Message: "verification successful",
		}

		mockSvc.On("VerifyOTP", mock.Anything, reqPayload).Return(expectedRes, nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/otp/verify", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid code returns 401 unauthorized", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.VerifyOTPRequest{
			Identifier: "test@example.com",
			Type:       enum.OTPTypeEmail,
			Code:       "00000",
		}

		mockSvc.On("VerifyOTP", mock.Anything, reqPayload).Return(nil, exception.ErrOTPInvalid)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/otp/verify", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp["success"].(bool))
		assert.Equal(t, "OTP_INVALID", resp["code"])
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		email := "user@example.com"
		reqPayload := dto.RegisterRequest{
			Name:     "Test User",
			Password: "securePassword123",
			Email:    &email,
		}
		expectedUser := &entity.User{
			BaseEntity:         entity.BaseEntity{ID: 1, CreatedAt: time.Now()},
			Name:               reqPayload.Name,
			Email:              &email,
			IsEmailVerified:    true,
			AvatarAttachmentID: nil,
			SystemRole:         enum.UserRoleUser,
			Status:             enum.UserStatusActive,
		}
		expectedRes := &dto.AuthResponse{
			User:                  dto.ToUserResponse(expectedUser),
			AccessToken:           "mock.access.token",
			AccessTokenExpiresAt:  time.Now().Add(24 * time.Hour),
			RefreshToken:          "mock.refresh.token",
			RefreshTokenExpiresAt: time.Now().Add(168 * time.Hour),
			ExpiresAt:             time.Now().Add(24 * time.Hour),
		}

		mockSvc.On("Register", mock.Anything, reqPayload).Return(expectedRes, nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		mockSvc.AssertExpectations(t)
	})

	t.Run("unverified credentials returns 403 forbidden", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		email := "unverified@example.com"
		reqPayload := dto.RegisterRequest{
			Name:     "Unverified User",
			Password: "securePassword123",
			Email:    &email,
		}

		mockSvc.On("Register", mock.Anything, reqPayload).Return(nil, exception.ErrUnverifiedCredentials)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp["success"].(bool))
		assert.Equal(t, "NOT_VERIFIED_CREDENTIALS", resp["code"])
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_LoginWithEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.LoginWithEmailRequest{
			Email: "test@example.com",
		}
		expectedRes := &dto.AuthResponse{
			AccessToken:           "mock.access.token",
			AccessTokenExpiresAt:  time.Now().Add(24 * time.Hour),
			RefreshToken:          "mock.refresh.token",
			RefreshTokenExpiresAt: time.Now().Add(168 * time.Hour),
		}

		mockSvc.On("LoginWithEmail", mock.Anything, reqPayload).Return(expectedRes, nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login/email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		mockSvc.AssertExpectations(t)
	})

	t.Run("unverified credentials returns 403", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.LoginWithEmailRequest{
			Email: "unverified@example.com",
		}

		mockSvc.On("LoginWithEmail", mock.Anything, reqPayload).Return(nil, exception.ErrUnverifiedCredentials)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login/email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("user not found returns 404", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.LoginWithEmailRequest{
			Email: "notfound@example.com",
		}

		mockSvc.On("LoginWithEmail", mock.Anything, reqPayload).Return(nil, exception.ErrUserNotFound)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login/email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_LoginWithPhone(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.LoginWithPhoneRequest{
			Phone: "09123456789",
		}
		expectedRes := &dto.AuthResponse{
			AccessToken:           "mock.access.token",
			AccessTokenExpiresAt:  time.Now().Add(24 * time.Hour),
			RefreshToken:          "mock.refresh.token",
			RefreshTokenExpiresAt: time.Now().Add(168 * time.Hour),
		}

		mockSvc.On("LoginWithPhone", mock.Anything, reqPayload).Return(expectedRes, nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login/phone", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		mockSvc.AssertExpectations(t)
	})

	t.Run("unverified credentials returns 403", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.LoginWithPhoneRequest{
			Phone: "09120000000",
		}

		mockSvc.On("LoginWithPhone", mock.Anything, reqPayload).Return(nil, exception.ErrUnverifiedCredentials)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login/phone", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupProtectedTestServer(mockSvc, 100)

		reqPayload := dto.ChangePasswordRequest{
			OldPassword: "oldpassword123",
			NewPassword: "newpassword123",
		}

		mockSvc.On("ChangePassword", mock.Anything, int64(100), reqPayload).Return(nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password/change", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		mockSvc.AssertExpectations(t)
	})

	t.Run("old password invalid returns 401", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupProtectedTestServer(mockSvc, 100)

		reqPayload := dto.ChangePasswordRequest{
			OldPassword: "wrongpassword",
			NewPassword: "newpassword123",
		}

		mockSvc.On("ChangePassword", mock.Anything, int64(100), reqPayload).Return(exception.ErrInvalidCredentials)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password/change", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_ResetPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.ResetPasswordRequest{
			Identifier:  "test@example.com",
			Type:        enum.OTPTypeEmail,
			NewPassword: "brandNewPassword123",
		}

		mockSvc.On("ResetPassword", mock.Anything, reqPayload).Return(nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		mockSvc.AssertExpectations(t)
	})

	t.Run("unverified credentials returns 403", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.ResetPasswordRequest{
			Identifier:  "unverified@example.com",
			Type:        enum.OTPTypeEmail,
			NewPassword: "brandNewPassword123",
		}

		mockSvc.On("ResetPassword", mock.Anything, reqPayload).Return(exception.ErrUnverifiedCredentials)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.LoginRequest{
			Identifier: "user@example.com",
			Password:   "password123",
		}
		expectedRes := &dto.AuthResponse{
			AccessToken:           "mock.jwt.token",
			AccessTokenExpiresAt:  time.Now().Add(24 * time.Hour),
			RefreshToken:          "mock.refresh.token",
			RefreshTokenExpiresAt: time.Now().Add(168 * time.Hour),
		}

		mockSvc.On("Login", mock.Anything, reqPayload).Return(expectedRes, nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid credentials returns 401 unauthorized", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.LoginRequest{
			Identifier: "user@example.com",
			Password:   "wrongpassword",
		}

		mockSvc.On("Login", mock.Anything, reqPayload).Return(nil, exception.ErrInvalidCredentials)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp["success"].(bool))
		assert.Equal(t, "INVALID_CREDENTIALS", resp["code"])
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.RefreshRequest{
			RefreshToken: "valid.refresh.token",
		}
		expectedRes := &dto.TokenPairResponse{
			AccessToken:           "new.access.token",
			AccessTokenExpiresAt:  time.Now().Add(24 * time.Hour),
			RefreshToken:          "new.refresh.token",
			RefreshTokenExpiresAt: time.Now().Add(168 * time.Hour),
		}

		mockSvc.On("Refresh", mock.Anything, reqPayload).Return(expectedRes, nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid refresh token returns 401", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		reqPayload := dto.RefreshRequest{
			RefreshToken: "bad.refresh.token",
		}

		mockSvc.On("Refresh", mock.Anything, reqPayload).Return(nil, exception.ErrRefreshTokenInvalid)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		token := "valid.access.token"
		mockSvc.On("Logout", mock.Anything, token).Return(nil)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("missing token returns 401", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		engine := setupTestServer(mockSvc)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
