package websocket

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/Doratou/Backend/internal/domain/repository/cache"
	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/redis/go-redis/v9"
)

// Runtime bundles the WebSocket hub, HTTP upgrade server, and optional Redis Pub/Sub relay
// behind a self-contained lifecycle manager. All methods are nil-safe when WebSocket is disabled.
type Runtime struct {
	cfg    Config
	Hub    *InMemoryHub
	Server *Server
	Relay  *RedisRelay

	mu          sync.Mutex
	relayCancel context.CancelFunc
	wg          sync.WaitGroup
}

// NewRuntime constructs a fully wired WebSocket Runtime from cfg and backing dependencies.
func NewRuntime(
	cfg Config,
	redisClient *redis.Client,
	tokenMaker jwt.TokenMaker,
	tokenRepo cache.AuthnTokenRepository,
	opts ...ServerOption,
) *Runtime {
	cfg = cfg.withDefaults()
	hub := NewInMemoryHub(cfg)

	var serverOpts []ServerOption
	if tokenMaker != nil {
		serverOpts = append(serverOpts, WithAuthenticator(NewJWTAuthenticator(tokenMaker, tokenRepo)))
	}
	serverOpts = append(serverOpts, opts...)

	srv := NewServer(hub, cfg, serverOpts...)

	var relay *RedisRelay
	if redisClient != nil {
		relay = NewRedisRelay(redisClient, cfg.PubSubChannel)
	}

	return &Runtime{
		cfg:    cfg,
		Hub:    hub,
		Server: srv,
		Relay:  relay,
	}
}

// Publisher returns the Pub/Sub publisher used by application services, or nil when disabled.
func (r *Runtime) Publisher() domain_ws.Publisher {
	if r == nil || r.Relay == nil {
		return nil
	}
	return r.Relay
}

// HTTPHandler returns the WebSocket HTTP upgrade handler, or nil when disabled.
func (r *Runtime) HTTPHandler() http.Handler {
	if r == nil || r.Server == nil {
		return nil
	}
	return r.Server
}

// Monitor returns the read-only stats/snapshot monitor, or nil when disabled.
func (r *Runtime) Monitor() domain_ws.Monitor {
	if r == nil || r.Server == nil {
		return nil
	}
	return r.Server
}

// Start launches the Redis Pub/Sub relay subscriber loop in the background when configured.
func (r *Runtime) Start(ctx context.Context) {
	if r == nil || r.Relay == nil || r.Hub == nil {
		return
	}

	relayCtx, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.relayCancel = cancel
	r.mu.Unlock()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := r.Relay.Run(relayCtx, r.Hub); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[WS REDIS RELAY ERROR] stopped with error: %v", err)
		}
	}()

	log.Printf("WebSocket runtime started (route: /api/v1/ws, auth_required: %v, pubsub_channel: %s)",
		r.cfg.RequireAuth, r.cfg.PubSubChannel)
}

// Shutdown stops the Redis relay loop and gracefully closes all active WebSocket connections.
func (r *Runtime) Shutdown() {
	if r == nil {
		return
	}

	r.mu.Lock()
	cancel := r.relayCancel
	r.relayCancel = nil
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}

	r.wg.Wait()

	if r.Hub != nil {
		ctx, cancelHub := context.WithTimeout(context.Background(), r.cfg.ShutdownTimeout)
		defer cancelHub()
		if err := r.Hub.Close(ctx); err != nil {
			log.Printf("error during WebSocket hub shutdown: %v", err)
		}
	}
}
