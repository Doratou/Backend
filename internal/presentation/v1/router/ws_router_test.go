package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	wsInfra "github.com/Doratou/Backend/internal/infrastructure/websocket"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type stubWSAuthTokenRepo struct{}

func (r *stubWSAuthTokenRepo) SaveRefreshSession(context.Context, string, string, string, time.Duration) error {
	return nil
}
func (r *stubWSAuthTokenRepo) GetRefreshSession(context.Context, string, string) (string, error) {
	return "", nil
}
func (r *stubWSAuthTokenRepo) DeleteRefreshSession(context.Context, string, string) (string, error) {
	return "", nil
}
func (r *stubWSAuthTokenRepo) DeleteUserRefreshSessions(context.Context, string) error {
	return nil
}
func (r *stubWSAuthTokenRepo) BlacklistAccessToken(context.Context, string, time.Duration) error {
	return nil
}
func (r *stubWSAuthTokenRepo) IsBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}

func TestWSRouter_MountsEndpoint(t *testing.T) {
	cfg := wsInfra.DefaultConfig()
	cfg.RequireAuth = false
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0

	hub := wsInfra.NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	wsSrv := wsInfra.NewServer(hub, cfg)
	engine := New(Dependencies{
		WSHandler: handler.NewWSHandler(wsSrv, wsSrv),
	})

	ts := httptest.NewServer(engine)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial /api/v1/ws via Gin router: %v", err)
	}
	defer conn.CloseNow()

	var env domain_ws.Envelope
	if err := wsjson.Read(ctx, conn, &env); err != nil {
		t.Fatalf("wsjson.Read failed: %v", err)
	}
	if env.Event != domain_ws.EventConnected {
		t.Fatalf("expected %s, got %s", domain_ws.EventConnected, env.Event)
	}

	_ = conn.Close(websocket.StatusNormalClosure, "done")
}

func TestWSRouter_AdminStatsAndSnapshot(t *testing.T) {
	maker, err := jwt.NewJWTMaker("super-secret-test-key-minimum-32-bytes!!")
	if err != nil {
		t.Fatalf("NewJWTMaker failed: %v", err)
	}
	tokenRepo := &stubWSAuthTokenRepo{}

	userToken, _, err := maker.CreateToken(7, enum.UserRoleUser, jwt.TokenTypeAccess, time.Hour)
	if err != nil {
		t.Fatalf("CreateToken user failed: %v", err)
	}
	adminToken, _, err := maker.CreateToken(1, enum.UserRoleAdmin, jwt.TokenTypeAccess, time.Hour)
	if err != nil {
		t.Fatalf("CreateToken admin failed: %v", err)
	}

	cfg := wsInfra.DefaultConfig()
	cfg.RequireAuth = true
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0

	hub := wsInfra.NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	wsAuth := wsInfra.NewJWTAuthenticator(maker, tokenRepo)
	wsSrv := wsInfra.NewServer(hub, cfg, wsInfra.WithAuthenticator(wsAuth))

	engine := New(Dependencies{
		WSHandler:  handler.NewWSHandler(wsSrv, wsSrv),
		TokenMaker: maker,
		TokenRepo:  tokenRepo,
	})

	ts := httptest.NewServer(engine)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Unauthenticated stats -> 401
	resp, err := http.Get(ts.URL + "/api/v1/admin/ws/stats")
	if err != nil {
		t.Fatalf("GET stats unauth failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.StatusCode)
	}

	// 2. Non-admin user -> 403
	reqUser, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/admin/ws/stats", nil)
	reqUser.Header.Set("Authorization", "Bearer "+userToken)
	respUser, err := http.DefaultClient.Do(reqUser)
	if err != nil {
		t.Fatalf("GET stats as user failed: %v", err)
	}
	_ = respUser.Body.Close()
	if respUser.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin user, got %d", respUser.StatusCode)
	}

	// Connect a real WS client as user 7 and subscribe to a channel
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/ws?token=" + userToken
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("WS Dial failed: %v", err)
	}
	defer conn.CloseNow()

	var welcome domain_ws.Envelope
	_ = wsjson.Read(ctx, conn, &welcome)
	subEnv, _ := domain_ws.NewEnvelope(domain_ws.EventSubscribe, nil)
	_ = wsjson.Write(ctx, conn, subEnv.WithChannel("room:ops"))
	var subAck domain_ws.Envelope
	_ = wsjson.Read(ctx, conn, &subAck)

	// 3. Admin stats -> 200 with active connection and channel
	reqAdminStats, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/admin/ws/stats", nil)
	reqAdminStats.Header.Set("Authorization", "Bearer "+adminToken)
	respStats, err := http.DefaultClient.Do(reqAdminStats)
	if err != nil {
		t.Fatalf("GET stats as admin failed: %v", err)
	}
	defer respStats.Body.Close()
	if respStats.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for admin stats, got %d", respStats.StatusCode)
	}
	var statsBody struct {
		Success bool               `json:"success"`
		Data    domain_ws.HubStats `json:"data"`
	}
	if err := json.NewDecoder(respStats.Body).Decode(&statsBody); err != nil {
		t.Fatalf("decode stats response failed: %v", err)
	}
	if !statsBody.Success || statsBody.Data.ActiveConnections != 1 || statsBody.Data.ActiveChannels != 1 || statsBody.Data.InstanceID == "" {
		t.Fatalf("unexpected stats body: %+v", statsBody)
	}

	// 4. Admin snapshot -> 200 with filtered connection details
	reqAdminSnap, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/admin/ws/snapshot?user_id=7&channel=room:ops&limit=10", nil)
	reqAdminSnap.Header.Set("Authorization", "Bearer "+adminToken)
	respSnap, err := http.DefaultClient.Do(reqAdminSnap)
	if err != nil {
		t.Fatalf("GET snapshot as admin failed: %v", err)
	}
	defer respSnap.Body.Close()
	if respSnap.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for admin snapshot, got %d", respSnap.StatusCode)
	}
	var snapBody struct {
		Success bool                  `json:"success"`
		Data    domain_ws.HubSnapshot `json:"data"`
	}
	if err := json.NewDecoder(respSnap.Body).Decode(&snapBody); err != nil {
		t.Fatalf("decode snapshot response failed: %v", err)
	}
	if !snapBody.Success || len(snapBody.Data.Connections) != 1 || snapBody.Data.Connections[0].UserID != 7 {
		t.Fatalf("unexpected snapshot body: %+v", snapBody)
	}

	// 5. Invalid limit -> 400
	reqBadLimit, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/admin/ws/snapshot?limit=-5", nil)
	reqBadLimit.Header.Set("Authorization", "Bearer "+adminToken)
	respBad, err := http.DefaultClient.Do(reqBadLimit)
	if err != nil {
		t.Fatalf("GET snapshot invalid limit failed: %v", err)
	}
	_ = respBad.Body.Close()
	if respBad.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid limit, got %d", respBad.StatusCode)
	}

	_ = conn.Close(websocket.StatusNormalClosure, "done")
}
