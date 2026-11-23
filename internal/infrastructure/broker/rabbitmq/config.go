package rabbitmq

import (
	"net"
	"net/url"
	"strings"
	"time"
)

// Config encapsulates configuration needed to connect and interact with RabbitMQ.
type Config struct {
	Host                  string
	Port                  string
	User                  string
	Password              string
	VHost                 string
	ConnectionTimeout     time.Duration
	Heartbeat             time.Duration
	InitialReconnectDelay time.Duration
	MaxReconnectDelay     time.Duration
}

// URL constructs a valid AMQP connection URL with proper credentials and vhost escaping.
func (c Config) URL() string {
	u := &url.URL{
		Scheme: "amqp",
		Host:   net.JoinHostPort(c.Host, c.Port),
	}

	if c.User != "" || c.Password != "" {
		u.User = url.UserPassword(c.User, c.Password)
	}

	vhost := c.VHost
	if vhost == "" {
		vhost = "/"
	}
	if !strings.HasPrefix(vhost, "/") {
		vhost = "/" + vhost
	}
	u.Path = vhost

	return u.String()
}

// DefaultConfig provides recommended production-grade defaults.
func DefaultConfig(host, port, user, password, vhost string) Config {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "5672"
	}
	if user == "" {
		user = "guest"
	}
	if password == "" {
		password = "guest"
	}
	if vhost == "" {
		vhost = "/"
	}

	return Config{
		Host:                  host,
		Port:                  port,
		User:                  user,
		Password:              password,
		VHost:                 vhost,
		ConnectionTimeout:     10 * time.Second,
		Heartbeat:             10 * time.Second,
		InitialReconnectDelay: 1 * time.Second,
		MaxReconnectDelay:     30 * time.Second,
	}
}
