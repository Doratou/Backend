package rabbitmq

import (
	"testing"
	"time"

	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
)

func TestConsumer_CalculateBackoffDelay(t *testing.T) {
	c := &Consumer{}

	policy := &domain_broker.RetryPolicy{
		Enabled:           true,
		MaxRetries:        5,
		InitialDelay:      2 * time.Second,
		BackoffMultiplier: 2.0,
		MaxDelay:          20 * time.Second,
	}

	tests := []struct {
		retryCount int
		expected   time.Duration
	}{
		{retryCount: 0, expected: 2 * time.Second},  // 2 * 2^0 = 2s
		{retryCount: 1, expected: 4 * time.Second},  // 2 * 2^1 = 4s
		{retryCount: 2, expected: 8 * time.Second},  // 2 * 2^2 = 8s
		{retryCount: 3, expected: 16 * time.Second}, // 2 * 2^3 = 16s
		{retryCount: 4, expected: 20 * time.Second}, // 2 * 2^4 = 32s -> capped at MaxDelay (20s)
	}

	for _, tt := range tests {
		got := c.calculateBackoffDelay(policy, tt.retryCount)
		if got != tt.expected {
			t.Errorf("calculateBackoffDelay(attempt=%d) = %v, want %v", tt.retryCount, got, tt.expected)
		}
	}
}
