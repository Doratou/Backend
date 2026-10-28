package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

// CaptchaService defines the application boundary for captcha challenge generation and verification.
type CaptchaService interface {
	// Generate creates a new captcha challenge image and stores its solution in cache.
	Generate(ctx context.Context) (*dto.CaptchaResponse, error)

	// Verify validates the user provided code against the cached challenge and consumes it on success.
	Verify(ctx context.Context, captchaID, code string) (bool, error)
}
