package websocket

import (
	"context"
	"errors"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

const (
	defaultSnapshotLimit = 100
	maxSnapshotLimit     = 1000
)

// InMemoryHub is a thread-safe, in-memory WebSocket connection registry and message router.
type InMemoryHub struct {
	cfg        Config
	instanceID string
	startedAt  time.Time

	mu           sync.RWMutex
	closed       bool
	conns        map[string]domain_ws.Client
	users        map[int64]map[string]domain_ws.Client
	anonConns    int
	channels     map[string]map[string]domain_ws.Client
	connChannels map[string]map[string]struct{}

	totalAccepted      atomic.Uint64
	totalClosed        atomic.Uint64
	totalRejectedLimit atomic.Uint64
	totalRejectedAuth  atomic.Uint64
	messagesEnqueued   atomic.Uint64
	messagesSent       atomic.Uint64
	messagesDropped    atomic.Uint64
	messagesReceived   atomic.Uint64
	inboundErrors      atomic.Uint64
	heartbeatFailures  atomic.Uint64

	activeWG sync.WaitGroup
}

// NewInMemoryHub creates a new InMemoryHub configured with cfg.
func NewInMemoryHub(cfg Config) *InMemoryHub {
	cfg = cfg.withDefaults()
	host, err := os.Hostname()
	if err != nil || strings.TrimSpace(host) == "" {
		host = "ws-node"
	}
	return &InMemoryHub{
		cfg:          cfg,
		instanceID:   host + "-" + uuid.NewString()[:8],
		startedAt:    time.Now().UTC(),
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
			h.totalRejectedLimit.Add(1)
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
	} else {
		h.anonConns++
	}
	h.totalAccepted.Add(1)

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
	h.totalClosed.Add(1)

	if userID := client.UserID(); userID > 0 {
		if userConns, ok := h.users[userID]; ok {
			delete(userConns, connID)
			if len(userConns) == 0 {
				delete(h.users, userID)
			}
		}
	} else if h.anonConns > 0 {
		h.anonConns--
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

// Stats returns a point-in-time snapshot of hub connection, user, channel, and cumulative counters.
func (h *InMemoryHub) Stats() domain_ws.HubStats {
	h.mu.RLock()
	stats := h.buildStatsLocked()
	h.mu.RUnlock()
	return stats
}

// Snapshot returns a filtered, bounded point-in-time view of hub metrics, active channels, and connections.
func (h *InMemoryHub) Snapshot(filter domain_ws.SnapshotFilter) domain_ws.HubSnapshot {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultSnapshotLimit
	} else if limit > maxSnapshotLimit {
		limit = maxSnapshotLimit
	}
	channelFilter := strings.TrimSpace(filter.Channel)

	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := h.buildStatsLocked()

	channels := make([]domain_ws.ChannelSnapshot, 0, len(h.channels))
	for ch, members := range h.channels {
		if channelFilter != "" && ch != channelFilter {
			continue
		}
		channels = append(channels, domain_ws.ChannelSnapshot{
			Channel:     ch,
			Subscribers: len(members),
		})
	}
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Channel < channels[j].Channel
	})

	// Choose the narrowest index when filters are provided.
	var candidates map[string]domain_ws.Client
	switch {
	case channelFilter != "":
		candidates = h.channels[channelFilter]
	case filter.UserID > 0:
		candidates = h.users[filter.UserID]
	default:
		candidates = h.conns
	}

	conns := make([]domain_ws.ConnectionSnapshot, 0, min(len(candidates), limit))
	for connID, client := range candidates {
		if filter.UserID > 0 && client.UserID() != filter.UserID {
			continue
		}
		conns = append(conns, h.buildConnSnapshotLocked(connID, client))
	}

	sort.Slice(conns, func(i, j int) bool {
		if !conns[i].ConnectedAt.Equal(conns[j].ConnectedAt) {
			return conns[i].ConnectedAt.Before(conns[j].ConnectedAt)
		}
		return conns[i].ConnID < conns[j].ConnID
	})

	truncated := false
	if len(conns) > limit {
		conns = conns[:limit]
		truncated = true
	}

	return domain_ws.HubSnapshot{
		Stats:       stats,
		Channels:    channels,
		Connections: conns,
		Truncated:   truncated,
	}
}

func (h *InMemoryHub) buildStatsLocked() domain_ws.HubStats {
	status := "running"
	if h.closed {
		status = "closed"
	}
	return domain_ws.HubStats{
		InstanceID:           h.instanceID,
		Status:               status,
		StartedAt:            h.startedAt,
		UptimeSeconds:        int64(time.Since(h.startedAt).Seconds()),
		ActiveConnections:    len(h.conns),
		AnonymousConnections: h.anonConns,
		UniqueUsers:          len(h.users),
		ActiveChannels:       len(h.channels),
		TotalAccepted:        h.totalAccepted.Load(),
		TotalClosed:          h.totalClosed.Load(),
		TotalRejectedLimit:   h.totalRejectedLimit.Load(),
		TotalRejectedAuth:    h.totalRejectedAuth.Load(),
		MessagesEnqueued:     h.messagesEnqueued.Load(),
		MessagesSent:         h.messagesSent.Load(),
		MessagesDropped:      h.messagesDropped.Load(),
		MessagesReceived:     h.messagesReceived.Load(),
		InboundErrors:        h.inboundErrors.Load(),
		HeartbeatFailures:    h.heartbeatFailures.Load(),
	}
}

func (h *InMemoryHub) buildConnSnapshotLocked(connID string, client domain_ws.Client) domain_ws.ConnectionSnapshot {
	subs := h.connChannels[connID]
	channels := make([]string, 0, len(subs))
	for ch := range subs {
		channels = append(channels, ch)
	}
	sort.Strings(channels)

	if wc, ok := client.(*wsClient); ok {
		return wc.snapshot(channels)
	}

	var role string
	if id := client.Identity(); id != nil {
		role = id.Role
	}
	return domain_ws.ConnectionSnapshot{
		ConnID:   client.ID(),
		UserID:   client.UserID(),
		Role:     role,
		Channels: channels,
	}
}

func (h *InMemoryHub) recordAuthRejected() {
	h.totalRejectedAuth.Add(1)
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
