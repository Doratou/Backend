package cache

import "context"

// PoWBlockRepository manages temporary client blocks that are set when a client
// exceeds the L2 rate limit.  Blocks live in Redis with a configurable TTL and
// are lifted as soon as the client successfully solves a PoW challenge.
//
// Keys passed to these methods should uniquely identify the client in the
// context of a specific route (e.g. "POST:/api/v1/auth/login:ip:1.2.3.4").
// The repository implementation is responsible for adding its own namespace
// prefix so keys do not collide with other Redis data.
type PoWBlockRepository interface {
	// IsBlocked reports whether the given client key currently has an active
	// PoW block.  It returns (false, nil) on a Redis miss and propagates
	// Redis errors to the caller.
	IsBlocked(ctx context.Context, key string) (bool, error)

	// Block records a PoW block for the given client key.  The block expires
	// automatically after the repository's configured BlockTTL.
	Block(ctx context.Context, key string) error

	// Unblock removes an active PoW block before its TTL expires.  It is
	// called when a client presents a valid PoW solution.  Calling Unblock
	// on a key that is not blocked is a no-op.
	Unblock(ctx context.Context, key string) error
}
