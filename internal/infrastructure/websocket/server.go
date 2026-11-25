package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Doratou/Backend/internal/domain/exception"
	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/coder/websocket"
)

// ServerOption customizes Server behavior.
type ServerOption func(*Server)

// WithAuthenticator attaches a pluggable handshake authenticator to the WebSocket server.
func WithAuthenticator(auth domain_ws.Authenticator) ServerOption {
	return func(s *Server) {
		s.auth = auth
	}
}

// WithInboundHandler attaches a handler for non-control envelopes sent by connected clients.
func WithInboundHandler(handler domain_ws.InboundHandler) ServerOption {
	return func(s *Server) {
		s.inboundHandler = handler
	}
}

// Server handles HTTP-to-WebSocket upgrades, handshake authentication, and client lifecycle execution.
type Server struct {
	hub            *InMemoryHub
	cfg            Config
	auth           domain_ws.Authenticator
	inboundHandler domain_ws.InboundHandler
}

// NewServer constructs a WebSocket Server backed by hub and cfg.
func NewServer(hub *InMemoryHub, cfg Config, opts ...ServerOption) *Server {
	s := &Server{
		hub: hub,
		cfg: cfg.withDefaults(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
}

// ServeHTTP authenticates the handshake request (if configured), upgrades to WebSocket using coder/websocket,
// registers the client in the hub, and runs its read/write/heartbeat pumps until disconnection.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.hub == nil {
		writeJSONError(w, exception.ErrInternal)
		return
	}

	var identity *domain_ws.Identity
	if s.auth != nil {
		id, err := s.auth.Authenticate(r.Context(), r)
		if err != nil {
			if s.cfg.RequireAuth {
				writeAuthError(w, err)
				return
			}
			// Optional auth mode: only allow anonymous fallback when no token was provided.
			if !errors.Is(err, exception.ErrMissingToken) {
				writeAuthError(w, err)
				return
			}
		} else {
			identity = id
		}
	}

	acceptOpts := &websocket.AcceptOptions{
		Subprotocols:       s.cfg.Subprotocols,
		InsecureSkipVerify: s.cfg.InsecureSkipVerify,
		OriginPatterns:     s.cfg.OriginPatterns,
		CompressionMode:    s.cfg.CompressionMode,
	}

	conn, err := websocket.Accept(w, r, acceptOpts)
	if err != nil {
		// websocket.Accept already writes the HTTP error response on handshake failure.
		log.Printf("[WS SERVER] accept failed: %v", err)
		return
	}

	conn.SetReadLimit(s.cfg.ReadLimit)

	client := newWSClient(r.Context(), conn, s.hub, identity, s.cfg, s.inboundHandler)
	if err := s.hub.Register(client); err != nil {
		code := websocket.StatusInternalError
		if errors.Is(err, domain_ws.ErrTooManyConnections) {
			code = websocket.StatusPolicyViolation
		} else if errors.Is(err, domain_ws.ErrHubClosed) {
			code = websocket.StatusGoingAway
		}
		_ = client.Close(int(code), err.Error())
		return
	}

	s.hub.trackClientStart()
	defer func() {
		s.hub.Unregister(client.ID())
		_ = client.Close(0, "")
		s.hub.trackClientDone()
	}()

	if s.cfg.SendConnectedEvent {
		connectedEnv, err := domain_ws.NewEnvelope(domain_ws.EventConnected, domain_ws.ConnectedData{
			ConnID: client.ID(),
			UserID: client.UserID(),
		})
		if err == nil {
			_ = client.Send(connectedEnv)
		}
	}

	client.run()
}

func writeAuthError(w http.ResponseWriter, err error) {
	var appErr *exception.AppError
	if errors.As(err, &appErr) {
		writeJSONError(w, appErr)
		return
	}
	if errors.Is(err, domain_ws.ErrUnauthorized) {
		writeJSONError(w, exception.ErrMissingToken)
		return
	}
	writeJSONError(w, exception.ErrTokenInvalid)
}

func writeJSONError(w http.ResponseWriter, appErr *exception.AppError) {
	if appErr == nil {
		appErr = exception.ErrInternal
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(appErr.HTTPStatus)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"code":    appErr.Code,
		"error":   appErr.Message,
	})
}
