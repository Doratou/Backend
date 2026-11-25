package websocket

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/coder/websocket"
)

// InMemoryHub is a thread-safe, in-memory WebSocket connection registry and message router.
type InMemoryHub struct {
	cfg Config

	mu           sync.RWMutex
	closed       bool
	conns        map[string]domain_ws.Client
	users        map[int64]map[string]domain_ws.Client
	channels     map[string]map[string]domain_ws.Client
	connChannels map[string]map[string]struct{}

	activeWG sync.WaitGroup
}

// NewInMemoryHub creates a new InMemoryHub configured with cfg.
func NewInMemoryHub(cfg Config) *InMemoryHub {
	cfg = cfg.withDefaults()
	return &InMemoryHub{
		cfg:          cfg,
		conns:        make(map[string]domain_ws.Client),
		users:        make(map[int64]map[string]domain_ws.Client),
		channels:     make(map[string]map[string]domain_ws.Client),
		connChannels: make(map[string]map[string]struct{}),
	}
}

// Register adds a client connection to the hub, enforcing MaxConnsPerUser if configured.
func (h *InMemoryHub) Register(client domain_ws.Client) error {
	if client == nil {
		return domain_ws.ErrConnectionNotFound
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return domain_ws.ErrHubClosed
	}

	userID := client.UserID()
	if userID > 0 && h.cfg.MaxConnsPerUser > 0 {
		if len(h.users[userID]) >= h.cfg.MaxConnsPerUser {
			return domain_ws.ErrTooManyConnections
		}
	}

	connID := client.ID()
	h.conns[connID] = client
	h.connChannels[connID] = make(map[string]struct{})

	if userID > 0 {
		if h.users[userID] == nil {
			h.users[userID] = make(map[string]domain_ws.Client)
		}
		h.users[userID][connID] = client
	}

	return nil
}

// Unregister removes a connection from all hub indexes (connections, users, and channels).
func (h *InMemoryHub) Unregister(connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, exists := h.conns[connID]
	if !exists {
		return
	}

	delete(h.conns, connID)

	if userID := client.UserID(); userID > 0 {
		if userConns, ok := h.users[userID]; ok {
			delete(userConns, connID)
			if len(userConns) == 0 {
				delete(h.users, userID)
			}
		}
	}

	if subs, ok := h.connChannels[connID]; ok {
		for ch := range subs {
			if members, ok := h.channels[ch]; ok {
				delete(members, connID)
				if len(members) == 0 {
					delete(h.channels, ch)
				}
			}
		}
		delete(h.connChannels, connID)
	}
}

// Subscribe attaches an active connection to a generic channel/room.
func (h *InMemoryHub) Subscribe(connID string, channel string) error {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return domain_ws.ErrInvalidChannel
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return domain_ws.ErrHubClosed
	}

	client, exists := h.conns[connID]
	if !exists {
		return domain_ws.ErrConnectionNotFound
	}

	if h.channels[channel] == nil {
		h.channels[channel] = make(map[string]domain_ws.Client)
	}
	h.channels[channel][connID] = client

	if h.connChannels[connID] == nil {
		h.connChannels[connID] = make(map[string]struct{})
	}
	h.connChannels[connID][channel] = struct{}{}

	return nil
}

// Unsubscribe detaches a connection from a generic channel/room.
func (h *InMemoryHub) Unsubscribe(connID string, channel string) error {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return domain_ws.ErrInvalidChannel
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.conns[connID]; !exists {
		return domain_ws.ErrConnectionNotFound
	}

	if members, ok := h.channels[channel]; ok {
		delete(members, connID)
		if len(members) == 0 {
			delete(h.channels, channel)
		}
	}

	if subs, ok := h.connChannels[connID]; ok {
		delete(subs, channel)
	}

	return nil
}

func (h *InMemoryHub) clientChannels(connID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs := h.connChannels[connID]
	if len(subs) == 0 {
		return nil
	}

	out := make([]string, 0, len(subs))
	for ch := range subs {
		out = append(out, ch)
	}
	sort.Strings(out)
	return out
}

// SendToConn enqueues an envelope to a specific connection ID.
func (h *InMemoryHub) SendToConn(ctx context.Context, connID string, env domain_ws.Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := env.Validate(); err != nil {
		return err
	}

	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return domain_ws.ErrHubClosed
	}
	client, exists := h.conns[connID]
	h.mu.RUnlock()

	if !exists {
		return domain_ws.ErrConnectionNotFound
	}

	return client.Send(env)
}

// SendToUser enqueues an envelope to all active connections belonging to userID.
func (h *InMemoryHub) SendToUser(ctx context.Context, userID int64, env domain_ws.Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := env.Validate(); err != nil {
		return err
	}

	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return domain_ws.ErrHubClosed
	}
	targets := copyClients(h.users[userID])
	h.mu.RUnlock()

	return dispatchToClients(ctx, targets, env)
}

// SendToUsers enqueues an envelope to all active connections belonging to any of userIDs.
func (h *InMemoryHub) SendToUsers(ctx context.Context, userIDs []int64, env domain_ws.Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := env.Validate(); err != nil {
		return err
	}

	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return domain_ws.ErrHubClosed
	}
	seen := make(map[string]struct{})
	var targets []domain_ws.Client
	for _, uid := range userIDs {
		for connID, client := range h.users[uid] {
			if _, ok := seen[connID]; ok {
				continue
			}
			seen[connID] = struct{}{}
			targets = append(targets, client)
		}
	}
	h.mu.RUnlock()

	return dispatchToClients(ctx, targets, env)
}

// SendToChannel enqueues an envelope to all connections subscribed to channel.
func (h *InMemoryHub) SendToChannel(ctx context.Context, channel string, env domain_ws.Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return domain_ws.ErrInvalidChannel
	}
	if err := env.Validate(); err != nil {
		return err
	}

	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return domain_ws.ErrHubClosed
	}
	targets := copyClients(h.channels[channel])
	h.mu.RUnlock()

	return dispatchToClients(ctx, targets, env)
}

// Broadcast enqueues an envelope to every active connection in the hub.
func (h *InMemoryHub) Broadcast(ctx context.Context, env domain_ws.Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := env.Validate(); err != nil {
		return err
	}

	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return domain_ws.ErrHubClosed
	}
	targets := copyClients(h.conns)
	h.mu.RUnlock()

	return dispatchToClients(ctx, targets, env)
}

// Stats returns a point-in-time snapshot of hub connection, user, and channel counts.
func (h *InMemoryHub) Stats() domain_ws.HubStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return domain_ws.HubStats{
		ActiveConnections: len(h.conns),
		UniqueUsers:       len(h.users),
		ActiveChannels:    len(h.channels),
	}
}

// Close marks the hub closed, closes all active connections with StatusGoingAway,
// and waits for active client loops to finish up to ctx deadline.
func (h *InMemoryHub) Close(ctx context.Context) error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	clients := copyClients(h.conns)
	h.mu.Unlock()

	var closeErrs []error
	for _, c := range clients {
		if err := c.Close(int(websocket.StatusGoingAway), "server shutting down"); err != nil {
			closeErrs = append(closeErrs, err)
		}
	}

	done := make(chan struct{})
	go func() {
		h.activeWG.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		for _, c := range clients {
			_ = c.Close(0, "")
		}
		return errors.Join(append(closeErrs, ctx.Err())...)
	}

	return errors.Join(closeErrs...)
}

func (h *InMemoryHub) trackClientStart() {
	h.activeWG.Add(1)
}

func (h *InMemoryHub) trackClientDone() {
	h.activeWG.Done()
}

func copyClients(m map[string]domain_ws.Client) []domain_ws.Client {
	if len(m) == 0 {
		return nil
	}
	out := make([]domain_ws.Client, 0, len(m))
	for _, c := range m {
		out = append(out, c)
	}
	return out
}

func dispatchToClients(ctx context.Context, targets []domain_ws.Client, env domain_ws.Envelope) error {
	var errs []error
	for _, c := range targets {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		if err := c.Send(env); err != nil && !errors.Is(err, domain_ws.ErrClientClosed) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
