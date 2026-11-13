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
	IsBlocked(ctx context.Context, key string) (bool, error)
	Block(ctx context.Context, key string) error
	Unblock(ctx context.Context, key string) error
}
