package cache

import (
	"context"
	"time"
)

// AuthnTokenRepository defines Redis cache operations for refresh token sessions and access token blacklist.
type AuthnTokenRepository interface {
	// SaveRefreshSession persists a refresh token mapped to userID and tokenID (jti) with TTL.
	SaveRefreshSession(ctx context.Context, userID, tokenID, refreshToken string, ttl time.Duration) error

	// GetRefreshSession retrieves the active refresh token for a given user and token ID.
	GetRefreshSession(ctx context.Context, userID, tokenID string) (string, error)

	// DeleteRefreshSession purges a specific refresh token session during rotation or logout.
	DeleteRefreshSession(ctx context.Context, userID, tokenID string) (string, error)

	// DeleteUserRefreshSessions purges all active refresh token sessions for a user (e.g. on full logout).
	DeleteUserRefreshSessions(ctx context.Context, userID string) error

	// BlacklistAccessToken stores revoked access token jti until its expiration.
	BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error

	// IsBlacklisted checks if an access token jti is present in the revocation blacklist.
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}
