package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrTrxConflict = errors.New("redis transaction: watched key changed, retries exhausted")

type TrxManager interface {
	// WithTransaction runs fn inside MULTI/EXEC pipeline.
	WithTransaction(ctx context.Context, fn func(trxCtx context.Context) error) error

	// WithWatch performs optimistic locking using WATCH/MULTI/EXEC with retry support.
	WithWatch(ctx context.Context, keys []string, read, write func(trxCtx context.Context) error) error
}

type redisTrxManager struct {
	cache      Cache
	maxRetries int
	backoff    time.Duration
}

func NewTrxManager(cache Cache, maxRetries int, backoff time.Duration) TrxManager {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if backoff <= 0 {
		backoff = 20 * time.Millisecond
	}
	return &redisTrxManager{cache: cache, maxRetries: maxRetries, backoff: backoff}
}

func (m *redisTrxManager) WithTransaction(ctx context.Context, fn func(trxCtx context.Context) error) error {
	_, err := m.cache.GetClient().TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		return fn(InjectTrx(ctx, m.cache.WithCmdable(pipe)))
	})
	return err
}

func (m *redisTrxManager) WithWatch(ctx context.Context, keys []string, read, write func(trxCtx context.Context) error) error {
	client := m.cache.GetClient()

	for attempt := 0; attempt < m.maxRetries; attempt++ {
		err := client.Watch(ctx, func(tx *redis.Tx) error {
			if read != nil {
				if err := read(InjectTrx(ctx, m.cache.WithCmdable(tx))); err != nil {
					return err
				}
			}
			_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				return write(InjectTrx(ctx, m.cache.WithCmdable(pipe)))
			})
			return err
		}, keys...)

		if !errors.Is(err, redis.TxFailedErr) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.backoff * time.Duration(attempt+1)):
		}
	}
	return fmt.Errorf("%w (keys=%v)", ErrTrxConflict, keys)
}
