package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type stubTokenRepo struct {
	mu          sync.RWMutex
	blacklisted map[string]bool
	err         error
}

func newStubTokenRepo() *stubTokenRepo {
	return &stubTokenRepo{blacklisted: make(map[string]bool)}
}

func (r *stubTokenRepo) SaveRefreshSession(context.Context, string, string, string, time.Duration) error {
	return nil
}
func (r *stubTokenRepo) GetRefreshSession(context.Context, string, string) (string, error) {
	return "", nil
}
func (r *stubTokenRepo) DeleteRefreshSession(context.Context, string, string) (string, error) {
	return "", nil
}
func (r *stubTokenRepo) DeleteUserRefreshSessions(context.Context, string) error {
	return nil
}
func (r *stubTokenRepo) BlacklistAccessToken(_ context.Context, jti string, _ time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blacklisted[jti] = true
	return nil
}
func (r *stubTokenRepo) IsBlacklisted(_ context.Context, jti string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.err != nil {
		return false, r.err
	}
	return r.blacklisted[jti], nil
}

func wsURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http")
}

func readEnvelopeWithTimeout(t *testing.T, conn *websocket.Conn, timeout time.Duration) domain_ws.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var env domain_ws.Envelope
	if err := wsjson.Read(ctx, conn, &env); err != nil {
		t.Fatalf("wsjson.Read failed: %v", err)
	}
	return env
}

func TestServer_ConnectAndControlEvents(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RequireAuth = false
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0

	hub := NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	srv := NewServer(hub, cfg)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(ts.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial failed: %v", err)
	}
	defer conn.CloseNow()

	// 1. Expect ws.connected welcome frame
	welcome := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	if welcome.Event != domain_ws.EventConnected {
		t.Fatalf("expected %s, got %s", domain_ws.EventConnected, welcome.Event)
	}
	var connData domain_ws.ConnectedData
	if err := welcome.DecodeData(&connData); err != nil {
		t.Fatalf("DecodeData failed: %v", err)
	}
	if connData.ConnID == "" {
		t.Fatal("expected non-empty conn_id")
	}

	// 2. Send ws.ping -> expect ws.pong with matching ID
	pingEnv, _ := domain_ws.NewEnvelope(domain_ws.EventPing, nil)
	if err := wsjson.Write(ctx, conn, pingEnv.WithID("p1")); err != nil {
		t.Fatalf("write ping failed: %v", err)
	}
	pong := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	if pong.Event != domain_ws.EventPong || pong.ID != "p1" {
		t.Fatalf("unexpected pong response: %+v", pong)
	}

	// 3. Subscribe to channel -> expect ws.subscribed, then receive channel broadcast
	subEnv, _ := domain_ws.NewEnvelope(domain_ws.EventSubscribe, nil)
	if err := wsjson.Write(ctx, conn, subEnv.WithID("s1").WithChannel("room:lobby")); err != nil {
		t.Fatalf("write subscribe failed: %v", err)
	}
	subAck := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	if subAck.Event != domain_ws.EventSubscribed || subAck.Channel != "room:lobby" {
		t.Fatalf("unexpected subscribe ack: %+v", subAck)
	}

	msgEnv, _ := domain_ws.NewEnvelope("infra.notice", map[string]string{"msg": "hello"})
	if err := hub.SendToChannel(ctx, "room:lobby", msgEnv.WithChannel("room:lobby")); err != nil {
		t.Fatalf("SendToChannel failed: %v", err)
	}
	received := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	if received.Event != "infra.notice" || received.Channel != "room:lobby" {
		t.Fatalf("unexpected channel message: %+v", received)
	}

	// 4. Unsubscribe -> expect ws.unsubscribed
	unsubEnv, _ := domain_ws.NewEnvelope(domain_ws.EventUnsubscribe, nil)
	if err := wsjson.Write(ctx, conn, unsubEnv.WithID("u1").WithChannel("room:lobby")); err != nil {
		t.Fatalf("write unsubscribe failed: %v", err)
	}
	unsubAck := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	if unsubAck.Event != domain_ws.EventUnsubscribed || unsubAck.Channel != "room:lobby" {
		t.Fatalf("unexpected unsubscribe ack: %+v", unsubAck)
	}

	_ = conn.Close(websocket.StatusNormalClosure, "bye")
}

func TestServer_JWTAuthentication(t *testing.T) {
	maker, err := jwt.NewJWTMaker("super-secret-test-key-minimum-32-bytes!!")
	if err != nil {
		t.Fatalf("NewJWTMaker failed: %v", err)
	}
	tokenRepo := newStubTokenRepo()
	auth := NewJWTAuthenticator(maker, tokenRepo)

	accessToken, accessClaims, err := maker.CreateToken(99, enum.UserRoleUser, jwt.TokenTypeAccess, time.Hour)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}
	refreshToken, _, err := maker.CreateToken(99, enum.UserRoleUser, jwt.TokenTypeRefresh, time.Hour)
	if err != nil {
		t.Fatalf("CreateToken refresh failed: %v", err)
	}
	expiredToken, _, err := maker.CreateToken(99, enum.UserRoleUser, jwt.TokenTypeAccess, -time.Minute)
	if err != nil {
		t.Fatalf("CreateToken expired failed: %v", err)
	}

	cfg := DefaultConfig()
	cfg.RequireAuth = true
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0

	hub := NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	srv := NewServer(hub, cfg, WithAuthenticator(auth))
	ts := httptest.NewServer(srv)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Missing token -> 401
	_, resp, err := websocket.Dial(ctx, wsURL(ts.URL), nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got err=%v resp=%v", err, resp)
	}

	// 2. Refresh token instead of access token -> 401
	_, resp, err = websocket.Dial(ctx, wsURL(ts.URL)+"?token="+refreshToken, nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for refresh token, got err=%v resp=%v", err, resp)
	}

	// 3. Expired token -> 401
	_, resp, err = websocket.Dial(ctx, wsURL(ts.URL)+"?token="+expiredToken, nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got err=%v resp=%v", err, resp)
	}

	// 4. Valid token via query param -> 101 + user_id=99
	connQuery, _, err := websocket.Dial(ctx, wsURL(ts.URL)+"?token="+accessToken, nil)
	if err != nil {
		t.Fatalf("expected successful dial with query token, got %v", err)
	}
	welcome := readEnvelopeWithTimeout(t, connQuery, 2*time.Second)
	var data domain_ws.ConnectedData
	_ = welcome.DecodeData(&data)
	if data.UserID != 99 {
		t.Fatalf("expected UserID 99, got %d", data.UserID)
	}
	_ = connQuery.Close(websocket.StatusNormalClosure, "")

	// 5. Valid token via Authorization header -> 101
	connHeader, _, err := websocket.Dial(ctx, wsURL(ts.URL), &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer " + accessToken}},
	})
	if err != nil {
		t.Fatalf("expected successful dial with header token, got %v", err)
	}
	welcome2 := readEnvelopeWithTimeout(t, connHeader, 2*time.Second)
	var data2 domain_ws.ConnectedData
	_ = welcome2.DecodeData(&data2)
	if data2.UserID != 99 {
		t.Fatalf("expected UserID 99 from header auth, got %d", data2.UserID)
	}
	_ = connHeader.Close(websocket.StatusNormalClosure, "")

	// 6. Blacklisted token -> 401
	_ = tokenRepo.BlacklistAccessToken(ctx, accessClaims.ID, time.Hour)
	_, resp, err = websocket.Dial(ctx, wsURL(ts.URL)+"?token="+accessToken, nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for blacklisted token, got err=%v resp=%v", err, resp)
	}
}

func TestServer_InboundHandlerAndInvalidJSON(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RequireAuth = false
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0

	hub := NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	echoHandler := domain_ws.InboundHandlerFunc(func(_ context.Context, client domain_ws.Client, env domain_ws.Envelope) error {
		reply, err := domain_ws.NewEnvelope("echo."+env.Event, env.Data)
		if err != nil {
			return err
		}
		return client.Send(reply.WithID(env.ID))
	})

	srv := NewServer(hub, cfg, WithInboundHandler(echoHandler))
	ts := httptest.NewServer(srv)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(ts.URL), nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.CloseNow()

	_ = readEnvelopeWithTimeout(t, conn, 2*time.Second) // ws.connected

	// Send malformed JSON -> expect ws.error without dropping connection
	if err := conn.Write(ctx, websocket.MessageText, []byte("{not-json")); err != nil {
		t.Fatalf("Write invalid JSON failed: %v", err)
	}
	errEnv := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	if errEnv.Event != domain_ws.EventError {
		t.Fatalf("expected %s, got %s", domain_ws.EventError, errEnv.Event)
	}

	// Send custom envelope -> expect echo from InboundHandler
	customEnv, _ := domain_ws.NewEnvelope("custom", json.RawMessage(`{"ok":true}`))
	if err := wsjson.Write(ctx, conn, customEnv.WithID("e1")); err != nil {
		t.Fatalf("Write custom envelope failed: %v", err)
	}
	echo := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	if echo.Event != "echo.custom" || echo.ID != "e1" {
		t.Fatalf("unexpected echo response: %+v", echo)
	}
}

func TestServer_BackpressureSlowConsumer(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RequireAuth = false
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0
	cfg.SendBufferSize = 1
	cfg.SlowConsumerPolicy = SlowConsumerDisconnect

	hub := NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	srv := NewServer(hub, cfg)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(ts.URL), nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.CloseNow()

	deadline := time.Now().Add(2 * time.Second)
	for hub.Stats().ActiveConnections == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if hub.Stats().ActiveConnections != 1 {
		t.Fatal("expected 1 active connection")
	}

	// Send large payloads in a tight burst so the single-slot sendCh fills immediately.
	env, _ := domain_ws.NewEnvelope("flood", map[string]string{"pad": strings.Repeat("x", 8192)})
	var gotBufferFull bool
	for i := 0; i < 500; i++ {
		if err := hub.Broadcast(ctx, env); errors.Is(err, domain_ws.ErrSendBufferFull) {
			gotBufferFull = true
			break
		}
	}
	if !gotBufferFull {
		t.Fatal("expected ErrSendBufferFull for stalled slow consumer")
	}

	// SlowConsumerDisconnect should evict the client promptly.
	evictDeadline := time.Now().Add(2 * time.Second)
	for hub.Stats().ActiveConnections != 0 && time.Now().Before(evictDeadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if hub.Stats().ActiveConnections != 0 {
		t.Fatalf("expected slow consumer to be disconnected, active=%d", hub.Stats().ActiveConnections)
	}
}

func TestServer_OptionalAuthAndOriginCheck(t *testing.T) {
	maker, err := jwt.NewJWTMaker("super-secret-test-key-minimum-32-bytes!!")
	if err != nil {
		t.Fatalf("NewJWTMaker failed: %v", err)
	}
	auth := NewJWTAuthenticator(maker, newStubTokenRepo())

	cfg := DefaultConfig()
	cfg.RequireAuth = false
	cfg.OriginPatterns = []string{"allowed.example.com"}
	cfg.InsecureSkipVerify = false
	cfg.PingInterval = 0

	hub := NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	srv := NewServer(hub, cfg, WithAuthenticator(auth))
	ts := httptest.NewServer(srv)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Disallowed Origin -> 403
	_, resp, err := websocket.Dial(ctx, wsURL(ts.URL), &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"https://evil.example.com"}},
	})
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for disallowed origin, got err=%v resp=%v", err, resp)
	}

	// 2. Allowed Origin without token (RequireAuth=false) -> anonymous UserID=0
	conn, _, err := websocket.Dial(ctx, wsURL(ts.URL), &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"https://allowed.example.com"}},
	})
	if err != nil {
		t.Fatalf("expected anonymous dial with allowed origin to succeed, got %v", err)
	}
	welcome := readEnvelopeWithTimeout(t, conn, 2*time.Second)
	var data domain_ws.ConnectedData
	_ = welcome.DecodeData(&data)
	if data.UserID != 0 {
		t.Fatalf("expected anonymous UserID 0, got %d", data.UserID)
	}
	_ = conn.Close(websocket.StatusNormalClosure, "")

	// 3. Allowed Origin with invalid token even when RequireAuth=false -> 401
	_, resp, err = websocket.Dial(ctx, wsURL(ts.URL)+"?token=bad.jwt.token", &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"https://allowed.example.com"}},
	})
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 when invalid token is provided in optional auth mode, got err=%v resp=%v", err, resp)
	}
}

func TestServer_HeartbeatPingPong(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RequireAuth = false
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 40 * time.Millisecond
	cfg.PongTimeout = 500 * time.Millisecond

	hub := NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	srv := NewServer(hub, cfg)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(ts.URL), nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.CloseNow()

	// Continuously read so client responds to server-initiated WebSocket Ping frames.
	readCtx, readCancel := context.WithCancel(ctx)
	defer readCancel()
	go func() {
		for {
			if _, _, err := conn.Read(readCtx); err != nil {
				return
			}
		}
	}()

	time.Sleep(150 * time.Millisecond)
	if hub.Stats().ActiveConnections != 1 {
		t.Fatalf("expected connection to stay alive across heartbeats, active=%d", hub.Stats().ActiveConnections)
	}
}

func TestServer_GracefulShutdown(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RequireAuth = false
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0

	hub := NewInMemoryHub(cfg)
	srv := NewServer(hub, cfg)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(ts.URL), nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.CloseNow()

	_ = readEnvelopeWithTimeout(t, conn, 2*time.Second) // ws.connected

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	if err := hub.Close(shutdownCtx); err != nil {
		t.Fatalf("hub.Close failed: %v", err)
	}

	_, _, err = conn.Read(ctx)
	if status := websocket.CloseStatus(err); status != websocket.StatusGoingAway {
		t.Fatalf("expected StatusGoingAway (1001), got status=%d err=%v", status, err)
	}
}
