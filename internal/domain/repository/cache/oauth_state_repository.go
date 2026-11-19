package cache

import (
	"context"
	"time"
)

// OAuthStateRepository manages short-lived OAuth state and nonce tokens in Redis to mitigate CSRF and replay attacks.
type OAuthStateRepository interface {
	// SaveState stores the state-to-nonce mapping with the given TTL.
	SaveState(ctx context.Context, state, nonce string, ttl time.Duration) error

	// ConsumeState retrieves the nonce associated with state and atomically removes the key (single-use).
	// Returns ErrOAuthStateInvalid if the state does not exist or has expired.
	ConsumeState(ctx context.Context, state string) (string, error)
}
