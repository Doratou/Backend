package cache

import (
	"context"
	"time"
)

// CaptchaRepository defines the cache operations for managing captcha challenge tokens.
type CaptchaRepository interface {
	// Save stores a captcha challenge answer indexed by captchaID with an expiration TTL.
	Save(ctx context.Context, captchaID, code string, ttl time.Duration) error

	// Get retrieves the expected captcha code for a given captchaID.
	Get(ctx context.Context, captchaID string) (string, error)

	// Delete removes the captcha challenge from cache to enforce single-use consumption.
	Delete(ctx context.Context, captchaID string) error

	// Exists checks if a captcha ID exists in cache without reading its value.
	Exists(ctx context.Context, captchaID string) (bool, error)
}
