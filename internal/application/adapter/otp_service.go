package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type OTPService interface {
	SendOTP(ctx context.Context, identifier string, otpType enum.OTPType) (*dto.SendOTPResponse, error)
	VerifyOTP(ctx context.Context, identifier string, otpType enum.OTPType, code string) (*dto.VerifyOTPResponse, error)
	IsVerified(ctx context.Context, identifier string, otpType enum.OTPType) (bool, error)
	ClearVerified(ctx context.Context, identifier string, otpType enum.OTPType) error
}
