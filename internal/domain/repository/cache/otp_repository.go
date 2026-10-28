package cache

import (
	"context"
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

// OTPRepository defines the caching, verification, and rate limiting interface for one-time passwords.
type OTPRepository interface {
	// SetOTP stores an OTP code for an identifier with an expiration TTL.
	SetOTP(ctx context.Context, otpType enum.OTPType, identifier, code string, ttl time.Duration) error

	// GetOTP retrieves the active OTP code for an identifier.
	GetOTP(ctx context.Context, otpType enum.OTPType, identifier string) (string, error)

	// DeleteOTP removes the OTP code from cache.
	DeleteOTP(ctx context.Context, otpType enum.OTPType, identifier string) error

	// GetOTPTTL returns the remaining time to live for an active OTP code.
	GetOTPTTL(ctx context.Context, otpType enum.OTPType, identifier string) (time.Duration, error)

	// SetVerified marks an identifier as verified for a given TTL session window.
	SetVerified(ctx context.Context, otpType enum.OTPType, identifier string, ttl time.Duration) error

	// IsVerified checks whether an identifier has a valid verified session.
	IsVerified(ctx context.Context, otpType enum.OTPType, identifier string) (bool, error)

	// ClearVerified purges the verified status after successful consumption.
	ClearVerified(ctx context.Context, otpType enum.OTPType, identifier string) error
}
