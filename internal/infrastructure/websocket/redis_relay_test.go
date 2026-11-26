package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type relayTestTokenRepo struct{}

func (r *relayTestTokenRepo) SaveRefreshSession(context.Context, string, string, string, time.Duration) error {
	return nil
}
func (r *relayTestTokenRepo) GetRefreshSession(context.Context, string, string) (string, error) {
	return "", nil
}
func (r *relayTestTokenRepo) DeleteRefreshSession(context.Context, string, string) (string, error) {
	return "", nil
}
func (r *relayTestTokenRepo) DeleteUserRefreshSessions(context.Context, string) error {
	return nil
}
func (r *relayTestTokenRepo) BlacklistAccessToken(context.Context, string, time.Duration) error {
	return nil
}
func (r *relayTestTokenRepo) IsBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}

func TestDispatchRelayPayload_RoutesUsersChannelAndBroadcast(t *testing.T) {
	maker, err := jwt.NewJWTMaker("super-secret-test-key-minimum-32-bytes!!")
	if err != nil {
		t.Fatalf("NewJWTMaker failed: %v", err)
	}
	token1, _, _ := maker.CreateToken(101, enum.UserRoleUser, jwt.TokenTypeAccess, time.Hour)
	token2, _, _ := maker.CreateToken(202, enum.UserRoleUser, jwt.TokenTypeAccess, time.Hour)

	cfg := DefaultConfig()
	cfg.RequireAuth = true
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0

	hub := NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	srv := NewServer(hub, cfg, WithAuthenticator(NewJWTAuthenticator(maker, &relayTestTokenRepo{})))
	ts := httptest.NewServer(srv)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsBase := "ws" + strings.TrimPrefix(ts.URL, "http")

	c1, _, err := websocket.Dial(ctx, wsBase+"?token="+token1, nil)
	if err != nil {
		t.Fatalf("Dial user 101 failed: %v", err)
	}
	defer c1.CloseNow()

	c2, _, err := websocket.Dial(ctx, wsBase+"?token="+token2, nil)
	if err != nil {
		t.Fatalf("Dial user 202 failed: %v", err)
	}
	defer c2.CloseNow()

	// Drain initial ws.connected envelopes
	var env domain_ws.Envelope
	if err := wsjson.Read(ctx, c1, &env); err != nil {
		t.Fatalf("read c1 welcome failed: %v", err)
	}
	if err := wsjson.Read(ctx, c2, &env); err != nil {
		t.Fatalf("read c2 welcome failed: %v", err)
	}

	// 1. TargetUsers with Channel metadata (group notification style without manual ws.subscribe)
	userMsg := domain_ws.RelayMessage{
		ID:         "evt-1",
		Event:      "group.announcement",
		TargetType: domain_ws.RelayTargetUsers,
		TargetIDs:  []int64{101},
		Channel:    "group:5",
		Data:       json.RawMessage(`{"title":"Trip","group_id":5}`),
		Timestamp:  time.Now().UTC(),
	}
	rawUserMsg, _ := json.Marshal(userMsg)
	if err := DispatchRelayPayload(ctx, hub, rawUserMsg); err != nil {
		t.Fatalf("DispatchRelayPayload users failed: %v", err)
	}

	var got1 domain_ws.Envelope
	if err := wsjson.Read(ctx, c1, &got1); err != nil {
		t.Fatalf("read c1 userMsg failed: %v", err)
	}
	if got1.Event != "group.announcement" || got1.Channel != "group:5" || got1.ID != "evt-1" {
		t.Fatalf("unexpected envelope on c1: %+v", got1)
	}

	// 2. Subscribe c2 to room:ops; the very next frame c2 reads must be ws.subscribed (not group.announcement)
	subEnv, _ := domain_ws.NewEnvelope(domain_ws.EventSubscribe, nil)
	if err := wsjson.Write(ctx, c2, subEnv.WithChannel("room:ops")); err != nil {
		t.Fatalf("subscribe c2 failed: %v", err)
	}
	if err := wsjson.Read(ctx, c2, &env); err != nil {
		t.Fatalf("read c2 subscribe ack failed: %v", err)
	}
	if env.Event != domain_ws.EventSubscribed || env.Channel != "room:ops" {
		t.Fatalf("expected ws.subscribed on c2 (proving user 101 event was not sent to 202), got %+v", env)
	}

	chMsg := domain_ws.RelayMessage{
		Event:      "room.ping",
		TargetType: domain_ws.RelayTargetChannel,
		Channel:    "room:ops",
		Data:       json.RawMessage(`{"ok":true}`),
	}
	rawChMsg, _ := json.Marshal(chMsg)
	if err := DispatchRelayPayload(ctx, hub, rawChMsg); err != nil {
		t.Fatalf("DispatchRelayPayload channel failed: %v", err)
	}
	if err := wsjson.Read(ctx, c2, &env); err != nil {
		t.Fatalf("read c2 channel msg failed: %v", err)
	}
	if env.Event != "room.ping" || env.Channel != "room:ops" {
		t.Fatalf("unexpected channel envelope on c2: %+v", env)
	}

	// 3. Broadcast reaches both
	bcMsg := domain_ws.RelayMessage{
		Event:      "token.price_updated",
		TargetType: domain_ws.RelayTargetBroadcast,
		Data:       json.RawMessage(`{"new_price_toman":50000}`),
	}
	rawBcMsg, _ := json.Marshal(bcMsg)
	if err := DispatchRelayPayload(ctx, hub, rawBcMsg); err != nil {
		t.Fatalf("DispatchRelayPayload broadcast failed: %v", err)
	}
	if err := wsjson.Read(ctx, c1, &env); err != nil || env.Event != "token.price_updated" {
		t.Fatalf("c1 failed to receive broadcast: env=%+v err=%v", env, err)
	}
	if err := wsjson.Read(ctx, c2, &env); err != nil || env.Event != "token.price_updated" {
		t.Fatalf("c2 failed to receive broadcast: env=%+v err=%v", env, err)
	}

	_ = c1.Close(websocket.StatusNormalClosure, "")
	_ = c2.Close(websocket.StatusNormalClosure, "")
}

func TestDispatchRelayPayload_ValidationErrors(t *testing.T) {
	hub := NewInMemoryHub(DefaultConfig())
	ctx := context.Background()

	if err := DispatchRelayPayload(ctx, hub, []byte("{bad")); !errors.Is(err, domain_ws.ErrInvalidEnvelope) {
		t.Fatalf("expected ErrInvalidEnvelope for malformed json, got %v", err)
	}

	invalidTarget, _ := json.Marshal(domain_ws.RelayMessage{
		Event:      "test",
		TargetType: "unknown",
	})
	if err := DispatchRelayPayload(ctx, hub, invalidTarget); !errors.Is(err, domain_ws.ErrInvalidEnvelope) {
		t.Fatalf("expected ErrInvalidEnvelope for unknown target, got %v", err)
	}

	missingChannel, _ := json.Marshal(domain_ws.RelayMessage{
		Event:      "test",
		TargetType: domain_ws.RelayTargetChannel,
	})
	if err := DispatchRelayPayload(ctx, hub, missingChannel); !errors.Is(err, domain_ws.ErrInvalidChannel) {
		t.Fatalf("expected ErrInvalidChannel for empty channel target, got %v", err)
	}
}

func TestRedisRelay_NilClientGuard(t *testing.T) {
	relay := NewRedisRelay(nil, "")
	if relay.Channel() != defaultRedisRelayChannel {
		t.Fatalf("expected default channel %q, got %q", defaultRedisRelayChannel, relay.Channel())
	}
	err := relay.Publish(context.Background(), domain_ws.RelayMessage{
		Event:      "test",
		TargetType: domain_ws.RelayTargetBroadcast,
	})
	if err == nil {
		t.Fatal("expected error when publishing with nil redis client")
	}

	// Ensure HTTP handler check doesn't fail compile/import sanity
	_ = http.StatusOK
}

func TestRuntime_NilSafeAndLifecycle(t *testing.T) {
	var nilRT *Runtime
	if nilRT.Publisher() != nil || nilRT.HTTPHandler() != nil || nilRT.Monitor() != nil {
		t.Fatal("expected nil Runtime accessors to return nil")
	}
	nilRT.Start(context.Background())
	nilRT.Shutdown()

	rt := NewRuntime(DefaultConfig(), nil, nil, nil)
	if rt.Publisher() != nil {
		t.Fatal("expected nil publisher when redisClient is nil")
	}
	if rt.HTTPHandler() == nil || rt.Monitor() == nil {
		t.Fatal("expected non-nil server and monitor")
	}
	rt.Start(context.Background())
	rt.Shutdown()
	if stats := rt.Monitor().Stats(); stats.Status != "closed" {
		t.Fatalf("expected hub status closed after Shutdown, got %q", stats.Status)
	}
}
