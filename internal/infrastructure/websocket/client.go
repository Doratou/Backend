package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

type wsClient struct {
	id             string
	identity       *domain_ws.Identity
	remoteAddr     string
	connectedAt    time.Time
	lastActivityNs atomic.Int64

	conn           *websocket.Conn
	hub            *InMemoryHub
	cfg            Config
	inboundHandler domain_ws.InboundHandler

	sendMu     sync.RWMutex
	sendCh     chan domain_ws.Envelope
	sendClosed bool

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	pumpWG    sync.WaitGroup
}

func newWSClient(
	parentCtx context.Context,
	conn *websocket.Conn,
	hub *InMemoryHub,
	identity *domain_ws.Identity,
	remoteAddr string,
	cfg Config,
	inboundHandler domain_ws.InboundHandler,
) *wsClient {
	if identity == nil {
		identity = &domain_ws.Identity{}
	}
	now := time.Now().UTC()
	ctx, cancel := context.WithCancel(parentCtx)
	c := &wsClient{
		id:             uuid.NewString(),
		identity:       identity,
		remoteAddr:     remoteAddr,
		connectedAt:    now,
		conn:           conn,
		hub:            hub,
		cfg:            cfg,
		inboundHandler: inboundHandler,
		sendCh:         make(chan domain_ws.Envelope, cfg.SendBufferSize),
		ctx:            ctx,
		cancel:         cancel,
	}
	c.lastActivityNs.Store(now.UnixNano())
	return c
}

func (c *wsClient) touchActivity() {
	c.lastActivityNs.Store(time.Now().UTC().UnixNano())
}

func (c *wsClient) snapshot(channels []string) domain_ws.ConnectionSnapshot {
	if channels == nil {
		channels = []string{}
	}
	var role string
	if c.identity != nil {
		role = c.identity.Role
	}
	return domain_ws.ConnectionSnapshot{
		ConnID:         c.id,
		UserID:         c.UserID(),
		Role:           role,
		RemoteAddr:     c.remoteAddr,
		Channels:       channels,
		SendQueueLen:   len(c.sendCh),
		SendQueueCap:   cap(c.sendCh),
		ConnectedAt:    c.connectedAt,
		LastActivityAt: time.Unix(0, c.lastActivityNs.Load()).UTC(),
	}
}

func (c *wsClient) ID() string {
	return c.id
}

func (c *wsClient) UserID() int64 {
	if c.identity == nil {
		return 0
	}
	return c.identity.UserID
}

func (c *wsClient) Identity() *domain_ws.Identity {
	return c.identity
}

func (c *wsClient) Channels() []string {
	if c.hub == nil {
		return nil
	}
	return c.hub.clientChannels(c.id)
}

func (c *wsClient) Send(env domain_ws.Envelope) error {
	c.sendMu.RLock()
	if c.sendClosed {
		c.sendMu.RUnlock()
		return domain_ws.ErrClientClosed
	}

	select {
	case c.sendCh <- env:
		c.sendMu.RUnlock()
		if c.hub != nil {
			c.hub.messagesEnqueued.Add(1)
		}
		return nil
	default:
		c.sendMu.RUnlock()
	}

	if c.hub != nil {
		c.hub.messagesDropped.Add(1)
	}
	if c.cfg.SlowConsumerPolicy == SlowConsumerDisconnect {
		go func() {
			_ = c.Close(int(websocket.StatusPolicyViolation), "slow consumer: send buffer full")
		}()
	}

	return domain_ws.ErrSendBufferFull
}

func (c *wsClient) Close(code int, reason string) error {
	var closeErr error
	c.closeOnce.Do(func() {
		c.sendMu.Lock()
		if !c.sendClosed {
			c.sendClosed = true
			close(c.sendCh)
		}
		c.sendMu.Unlock()

		if c.cancel != nil {
			c.cancel()
		}

		if c.conn != nil {
			if code == 0 {
				closeErr = c.conn.CloseNow()
			} else {
				closeErr = c.conn.Close(websocket.StatusCode(code), reason)
				if closeErr != nil {
					_ = c.conn.CloseNow()
				}
			}
		}
	})
	return closeErr
}

// run starts the write and heartbeat pumps, blocks on the read pump until disconnection or context cancellation,
// and ensures all background goroutines exit cleanly before returning.
func (c *wsClient) run() {
	defer c.cancel()

	c.pumpWG.Add(2)
	go func() {
		defer c.pumpWG.Done()
		c.writePump(c.ctx)
	}()
	go func() {
		defer c.pumpWG.Done()
		c.heartbeatLoop(c.ctx)
	}()

	c.readPump(c.ctx)

	// Ensure write & heartbeat pumps stop after readPump exits.
	_ = c.Close(int(websocket.StatusNormalClosure), "")
	c.pumpWG.Wait()
}

func (c *wsClient) readPump(ctx context.Context) {
	for {
		msgType, payload, err := c.conn.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if status != websocket.StatusNormalClosure &&
				status != websocket.StatusGoingAway &&
				!errors.Is(err, context.Canceled) {
				log.Printf("[WS CLIENT] read closed conn=%s user=%d status=%d err=%v", c.id, c.UserID(), status, err)
			}
			return
		}

		if msgType != websocket.MessageText && msgType != websocket.MessageBinary {
			continue
		}
		c.touchActivity()

		var env domain_ws.Envelope
		if err := json.Unmarshal(payload, &env); err != nil {
			if c.hub != nil {
				c.hub.inboundErrors.Add(1)
			}
			c.sendError("", "INVALID_ENVELOPE", "message must be a valid JSON envelope")
			continue
		}
		if err := env.Validate(); err != nil {
			if c.hub != nil {
				c.hub.inboundErrors.Add(1)
			}
			c.sendError(env.ID, "INVALID_ENVELOPE", err.Error())
			continue
		}
		if c.hub != nil {
			c.hub.messagesReceived.Add(1)
		}

		if c.cfg.EnableControlEvents && c.handleControlEvent(ctx, env) {
			continue
		}

		if c.inboundHandler != nil {
			if err := c.inboundHandler.HandleMessage(ctx, c, env); err != nil {
				if c.hub != nil {
					c.hub.inboundErrors.Add(1)
				}
				c.sendError(env.ID, "INBOUND_HANDLER_ERROR", err.Error())
			}
		}
	}
}

func (c *wsClient) handleControlEvent(_ context.Context, env domain_ws.Envelope) bool {
	switch env.Event {
	case domain_ws.EventPing:
		pong, err := domain_ws.NewEnvelope(domain_ws.EventPong, nil)
		if err == nil {
			_ = c.Send(pong.WithID(env.ID))
		}
		return true

	case domain_ws.EventSubscribe:
		chName := extractChannelName(env)
		if err := c.hub.Subscribe(c.id, chName); err != nil {
			c.sendError(env.ID, "SUBSCRIBE_FAILED", err.Error())
			return true
		}
		ack, err := domain_ws.NewEnvelope(domain_ws.EventSubscribed, domain_ws.ChannelAckData{Channel: chName})
		if err == nil {
			_ = c.Send(ack.WithID(env.ID).WithChannel(chName))
		}
		return true

	case domain_ws.EventUnsubscribe:
		chName := extractChannelName(env)
		if err := c.hub.Unsubscribe(c.id, chName); err != nil {
			c.sendError(env.ID, "UNSUBSCRIBE_FAILED", err.Error())
			return true
		}
		ack, err := domain_ws.NewEnvelope(domain_ws.EventUnsubscribed, domain_ws.ChannelAckData{Channel: chName})
		if err == nil {
			_ = c.Send(ack.WithID(env.ID).WithChannel(chName))
		}
		return true

	default:
		return false
	}
}

func extractChannelName(env domain_ws.Envelope) string {
	if ch := strings.TrimSpace(env.Channel); ch != "" {
		return ch
	}
	var payload domain_ws.ChannelAckData
	if err := env.DecodeData(&payload); err == nil {
		return strings.TrimSpace(payload.Channel)
	}
	return ""
}

func (c *wsClient) sendError(replyID, code, message string) {
	errEnv, err := domain_ws.NewEnvelope(domain_ws.EventError, domain_ws.ErrorData{
		Code:    code,
		Message: message,
	})
	if err != nil {
		return
	}
	_ = c.Send(errEnv.WithID(replyID))
}

func (c *wsClient) writePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case env, ok := <-c.sendCh:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, c.cfg.WriteTimeout)
			err := wsjson.Write(writeCtx, c.conn, env)
			cancel()
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Printf("[WS CLIENT] write failed conn=%s user=%d err=%v", c.id, c.UserID(), err)
				}
				_ = c.Close(0, "")
				return
			}
			c.touchActivity()
			if c.hub != nil {
				c.hub.messagesSent.Add(1)
			}
		}
	}
}

func (c *wsClient) heartbeatLoop(ctx context.Context) {
	if c.cfg.PingInterval <= 0 {
		return
	}

	ticker := time.NewTicker(c.cfg.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, c.cfg.PongTimeout)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					if c.hub != nil {
						c.hub.heartbeatFailures.Add(1)
					}
					log.Printf("[WS CLIENT] heartbeat ping failed conn=%s user=%d err=%v", c.id, c.UserID(), err)
				}
				_ = c.Close(0, "")
				return
			}
			c.touchActivity()
		}
	}
}
