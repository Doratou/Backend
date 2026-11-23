package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConnectionManager defines the interface for interacting with the resilient RabbitMQ connection.
type ConnectionManager interface {
	Connection() *amqp.Connection
	Channel() (*amqp.Channel, error)
	SubscribeReconnect() <-chan struct{}
	IsConnected() bool
	Ping(ctx context.Context) error
	Close() error
}

type connectionManager struct {
	cfg        Config
	conn       *amqp.Connection
	connMu     sync.RWMutex
	isClosed   atomic.Bool
	closedChan chan struct{}

	listenersMu sync.Mutex
	listeners   []chan struct{}
}

// NewConnectionManager establishes a resilient RabbitMQ connection and starts the auto-reconnect supervisor.
func NewConnectionManager(ctx context.Context, cfg Config) (ConnectionManager, error) {
	cm := &connectionManager{
		cfg:        cfg,
		closedChan: make(chan struct{}),
	}

	if err := cm.dial(ctx); err != nil {
		return nil, fmt.Errorf("rabbitmq: initial dial failed: %w", err)
	}

	go cm.supervisor()

	return cm, nil
}

func (cm *connectionManager) dial(ctx context.Context) error {
	dialConfig := amqp.Config{
		Heartbeat: cm.cfg.Heartbeat,
		Dial:      amqp.DefaultDial(cm.cfg.ConnectionTimeout),
	}

	type dialResult struct {
		conn *amqp.Connection
		err  error
	}

	ch := make(chan dialResult, 1)
	go func() {
		conn, err := amqp.DialConfig(cm.cfg.URL(), dialConfig)
		ch <- dialResult{conn: conn, err: err}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return res.err
		}
		cm.connMu.Lock()
		cm.conn = res.conn
		cm.connMu.Unlock()
		return nil
	}
}

func (cm *connectionManager) supervisor() {
	for {
		if cm.isClosed.Load() {
			return
		}

		cm.connMu.RLock()
		conn := cm.conn
		cm.connMu.RUnlock()

		if conn == nil || conn.IsClosed() {
			cm.reconnect()
			continue
		}

		closeErrChan := make(chan *amqp.Error, 1)
		conn.NotifyClose(closeErrChan)

		select {
		case <-cm.closedChan:
			return
		case err, ok := <-closeErrChan:
			if !ok || cm.isClosed.Load() {
				return
			}
			log.Printf("[RABBITMQ WARNING] connection closed: %v. Initiating reconnect...", err)
			cm.reconnect()
		}
	}
}

func (cm *connectionManager) reconnect() {
	delay := cm.cfg.InitialReconnectDelay
	if delay <= 0 {
		delay = 1 * time.Second
	}
	maxDelay := cm.cfg.MaxReconnectDelay
	if maxDelay <= 0 {
		maxDelay = 30 * time.Second
	}

	for {
		if cm.isClosed.Load() {
			return
		}

		// Add jitter to avoid thundering herd
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))
		sleepDuration := delay + jitter

		log.Printf("[RABBITMQ NOTICE] attempting reconnection in %v...", sleepDuration)
		time.Sleep(sleepDuration)

		if cm.isClosed.Load() {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), cm.cfg.ConnectionTimeout)
		err := cm.dial(ctx)
		cancel()

		if err == nil {
			log.Print("[RABBITMQ INFO] successfully reconnected to broker")
			cm.notifyListeners()
			return
		}

		log.Printf("[RABBITMQ WARNING] reconnection attempt failed: %v", err)
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}

func (cm *connectionManager) notifyListeners() {
	cm.listenersMu.Lock()
	defer cm.listenersMu.Unlock()

	for _, ch := range cm.listeners {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (cm *connectionManager) SubscribeReconnect() <-chan struct{} {
	cm.listenersMu.Lock()
	defer cm.listenersMu.Unlock()

	ch := make(chan struct{}, 1)
	cm.listeners = append(cm.listeners, ch)
	return ch
}

func (cm *connectionManager) Connection() *amqp.Connection {
	cm.connMu.RLock()
	defer cm.connMu.RUnlock()
	return cm.conn
}

func (cm *connectionManager) Channel() (*amqp.Channel, error) {
	cm.connMu.RLock()
	defer cm.connMu.RUnlock()

	if cm.conn == nil || cm.conn.IsClosed() {
		return nil, amqp.ErrClosed
	}

	return cm.conn.Channel()
}

func (cm *connectionManager) IsConnected() bool {
	cm.connMu.RLock()
	defer cm.connMu.RUnlock()
	return cm.conn != nil && !cm.conn.IsClosed()
}

func (cm *connectionManager) Ping(ctx context.Context) error {
	ch, err := cm.Channel()
	if err != nil {
		return fmt.Errorf("ping: failed to open channel: %w", err)
	}
	defer ch.Close()

	done := make(chan error, 1)
	go func() {
		// Attempt a light passive check (flow control query or declaring a temporary test channel operation)
		done <- ch.Flow(true)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("ping failed: %w", err)
		}
		return nil
	}
}

func (cm *connectionManager) Close() error {
	if !cm.isClosed.CompareAndSwap(false, true) {
		return nil
	}

	close(cm.closedChan)

	cm.connMu.Lock()
	defer cm.connMu.Unlock()

	if cm.conn != nil && !cm.conn.IsClosed() {
		return cm.conn.Close()
	}
	return nil
}
