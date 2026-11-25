package router

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	wsInfra "github.com/Doratou/Backend/internal/infrastructure/websocket"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestWSRouter_MountsEndpoint(t *testing.T) {
	cfg := wsInfra.DefaultConfig()
	cfg.RequireAuth = false
	cfg.InsecureSkipVerify = true
	cfg.PingInterval = 0

	hub := wsInfra.NewInMemoryHub(cfg)
	defer func() { _ = hub.Close(context.Background()) }()

	wsSrv := wsInfra.NewServer(hub, cfg)
	engine := New(Dependencies{
		WSHandler: handler.NewWSHandler(wsSrv),
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
