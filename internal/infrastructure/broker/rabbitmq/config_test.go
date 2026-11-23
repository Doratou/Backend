package rabbitmq

import (
	"testing"
)

func TestConfig_URL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expected string
	}{
		{
			name: "default local config",
			cfg: Config{
				Host:     "127.0.0.1",
				Port:     "5672",
				User:     "guest",
				Password: "guest",
				VHost:    "/",
			},
			expected: "amqp://guest:guest@127.0.0.1:5672/",
		},
		{
			name: "custom vhost without leading slash",
			cfg: Config{
				Host:     "rabbitmq.local",
				Port:     "5672",
				User:     "app_user",
				Password: "secretPassword123",
				VHost:    "doratou_vhost",
			},
			expected: "amqp://app_user:secretPassword123@rabbitmq.local:5672/doratou_vhost",
		},
		{
			name: "escaped password with special characters",
			cfg: Config{
				Host:     "10.0.0.5",
				Port:     "5672",
				User:     "user",
				Password: "p@ssword!#:^",
				VHost:    "/test",
			},
			expected: "amqp://user:p%40ssword%21%23%3A%5E@10.0.0.5:5672/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.URL()
			if got != tt.expected {
				t.Errorf("URL() = %q, want %q", got, tt.expected)
			}
		})
	}
}
