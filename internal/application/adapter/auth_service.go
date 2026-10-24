package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	LoginWithEmail(ctx context.Context, req dto.LoginWithEmailRequest) (*dto.AuthResponse, error)
	LoginWithPhone(ctx context.Context, req dto.LoginWithPhoneRequest) (*dto.AuthResponse, error)
	ChangePassword(ctx context.Context, userID int64, req dto.ChangePasswordRequest) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
	Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.TokenPairResponse, error)
	Logout(ctx context.Context, accessToken string) error
	SendOTP(ctx context.Context, req dto.SendOTPRequest) (*dto.SendOTPResponse, error)
	VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error)
}
