package websocket

import (
	"context"
	"errors"
	"sync"
	"testing"

	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
)

type stubClient struct {
	id     string
	userID int64

	mu       sync.Mutex
	sent     []domain_ws.Envelope
	closed   bool
	closeCod int
	closeRsn string
	sendErr  error
}

func newStubClient(id string, userID int64) *stubClient {
	return &stubClient{id: id, userID: userID}
}

func (c *stubClient) ID() string                    { return c.id }
func (c *stubClient) UserID() int64                 { return c.userID }
func (c *stubClient) Identity() *domain_ws.Identity { return &domain_ws.Identity{UserID: c.userID} }
func (c *stubClient) Channels() []string            { return nil }
func (c *stubClient) Send(env domain_ws.Envelope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return domain_ws.ErrClientClosed
	}
	if c.sendErr != nil {
		return c.sendErr
	}
	c.sent = append(c.sent, env)
	return nil
}

func (c *stubClient) Close(code int, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	c.closeCod = code
	c.closeRsn = reason
	return nil
}

func (c *stubClient) sentCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.sent)
}

func TestInMemoryHub_RegisterUnregisterAndStats(t *testing.T) {
	hub := NewInMemoryHub(DefaultConfig())

	c1 := newStubClient("c1", 10)
	c2 := newStubClient("c2", 10) // second device for same user
	c3 := newStubClient("c3", 0)  // anonymous connection

	for _, c := range []*stubClient{c1, c2, c3} {
		if err := hub.Register(c); err != nil {
			t.Fatalf("Register(%s) failed: %v", c.ID(), err)
		}
	}

	if err := hub.Subscribe("c1", "room:1"); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	if err := hub.Subscribe("c2", "room:1"); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	stats := hub.Stats()
	if stats.ActiveConnections != 3 || stats.UniqueUsers != 1 || stats.ActiveChannels != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	hub.Unregister("c1")
	stats = hub.Stats()
	if stats.ActiveConnections != 2 || stats.UniqueUsers != 1 || stats.ActiveChannels != 1 {
		t.Fatalf("unexpected stats after unregister c1: %+v", stats)
	}

	hub.Unregister("c2")
	stats = hub.Stats()
	if stats.ActiveConnections != 1 || stats.UniqueUsers != 0 || stats.ActiveChannels != 0 {
		t.Fatalf("unexpected stats after unregister c2: %+v", stats)
	}
}

func TestInMemoryHub_MaxConnsPerUser(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxConnsPerUser = 2
	hub := NewInMemoryHub(cfg)

	if err := hub.Register(newStubClient("c1", 42)); err != nil {
		t.Fatalf("unexpected error on conn 1: %v", err)
	}
	if err := hub.Register(newStubClient("c2", 42)); err != nil {
		t.Fatalf("unexpected error on conn 2: %v", err)
	}
	if err := hub.Register(newStubClient("c3", 42)); !errors.Is(err, domain_ws.ErrTooManyConnections) {
		t.Fatalf("expected ErrTooManyConnections on conn 3, got %v", err)
	}
}

func TestInMemoryHub_Routing(t *testing.T) {
	ctx := context.Background()
	hub := NewInMemoryHub(DefaultConfig())

	c1 := newStubClient("c1", 1)
	c2 := newStubClient("c2", 1)
	c3 := newStubClient("c3", 2)
	for _, c := range []*stubClient{c1, c2, c3} {
		if err := hub.Register(c); err != nil {
			t.Fatalf("Register(%s) failed: %v", c.ID(), err)
		}
	}

	if err := hub.Subscribe("c1", "room:a"); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	if err := hub.Subscribe("c3", "room:a"); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	env, err := domain_ws.NewEnvelope("infra.test", map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("NewEnvelope failed: %v", err)
	}

	if err := hub.SendToConn(ctx, "c2", env); err != nil {
		t.Fatalf("SendToConn failed: %v", err)
	}
	if c2.sentCount() != 1 || c1.sentCount() != 0 {
		t.Fatalf("unexpected counts after SendToConn: c1=%d c2=%d", c1.sentCount(), c2.sentCount())
	}

	if err := hub.SendToUser(ctx, 1, env); err != nil {
		t.Fatalf("SendToUser failed: %v", err)
	}
	if c1.sentCount() != 1 || c2.sentCount() != 2 || c3.sentCount() != 0 {
		t.Fatalf("unexpected counts after SendToUser: c1=%d c2=%d c3=%d", c1.sentCount(), c2.sentCount(), c3.sentCount())
	}

	if err := hub.SendToUsers(ctx, []int64{1, 2, 1}, env); err != nil {
		t.Fatalf("SendToUsers failed: %v", err)
	}
	if c1.sentCount() != 2 || c2.sentCount() != 3 || c3.sentCount() != 1 {
		t.Fatalf("unexpected counts after SendToUsers: c1=%d c2=%d c3=%d", c1.sentCount(), c2.sentCount(), c3.sentCount())
	}

	if err := hub.SendToChannel(ctx, "room:a", env); err != nil {
		t.Fatalf("SendToChannel failed: %v", err)
	}
	if c1.sentCount() != 3 || c2.sentCount() != 3 || c3.sentCount() != 2 {
		t.Fatalf("unexpected counts after SendToChannel: c1=%d c2=%d c3=%d", c1.sentCount(), c2.sentCount(), c3.sentCount())
	}

	if err := hub.Unsubscribe("c3", "room:a"); err != nil {
		t.Fatalf("Unsubscribe failed: %v", err)
	}
	if err := hub.SendToChannel(ctx, "room:a", env); err != nil {
		t.Fatalf("SendToChannel after unsubscribe failed: %v", err)
	}
	if c1.sentCount() != 4 || c3.sentCount() != 2 {
		t.Fatalf("unexpected counts after unsubscribe: c1=%d c3=%d", c1.sentCount(), c3.sentCount())
	}

	if err := hub.Broadcast(ctx, env); err != nil {
		t.Fatalf("Broadcast failed: %v", err)
	}
	if c1.sentCount() != 5 || c2.sentCount() != 4 || c3.sentCount() != 3 {
		t.Fatalf("unexpected counts after Broadcast: c1=%d c2=%d c3=%d", c1.sentCount(), c2.sentCount(), c3.sentCount())
	}

	if err := hub.SendToConn(ctx, "missing", env); !errors.Is(err, domain_ws.ErrConnectionNotFound) {
		t.Fatalf("expected ErrConnectionNotFound, got %v", err)
	}
}

func TestInMemoryHub_Close(t *testing.T) {
	hub := NewInMemoryHub(DefaultConfig())
	c1 := newStubClient("c1", 1)
	if err := hub.Register(c1); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if err := hub.Close(context.Background()); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !c1.closed {
		t.Fatal("expected client to be closed when hub closes")
	}

	if err := hub.Register(newStubClient("c2", 2)); !errors.Is(err, domain_ws.ErrHubClosed) {
		t.Fatalf("expected ErrHubClosed on Register after Close, got %v", err)
	}
}
