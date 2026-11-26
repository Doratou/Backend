package rabbitmq

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDeliveryWrapper_Message(t *testing.T) {
	now := time.Now().UTC()
	amqpDelivery := amqp.Delivery{
		MessageId:     "msg-123",
		ContentType:   "application/json",
		Body:          []byte(`{"key":"value"}`),
		CorrelationId: "corr-456",
		ReplyTo:       "reply-queue",
		Timestamp:     now,
		Headers: amqp.Table{
			"x-test-header": "test-val",
		},
	}

	d := NewDelivery(amqpDelivery)
	msg := d.Message()

	if msg.ID != "msg-123" {
		t.Errorf("expected msg.ID = 'msg-123', got %q", msg.ID)
	}
	if msg.ContentType != "application/json" {
		t.Errorf("expected msg.ContentType = 'application/json', got %q", msg.ContentType)
	}
	if string(msg.Body) != `{"key":"value"}` {
		t.Errorf("expected msg.Body = '{\"key\":\"value\"}', got %q", string(msg.Body))
	}
	if msg.CorrelationID != "corr-456" {
		t.Errorf("expected msg.CorrelationID = 'corr-456', got %q", msg.CorrelationID)
	}
	if msg.Headers["x-test-header"] != "test-val" {
		t.Errorf("expected header 'test-val', got %v", msg.Headers["x-test-header"])
	}
}

func TestDeliveryWrapper_RetryCount(t *testing.T) {
	t.Run("without retry header", func(t *testing.T) {
		d := NewDelivery(amqp.Delivery{})
		if count := d.RetryCount(); count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})

	t.Run("with x-retry-count int", func(t *testing.T) {
		d := NewDelivery(amqp.Delivery{
			Headers: amqp.Table{"x-retry-count": 3},
		})
		if count := d.RetryCount(); count != 3 {
			t.Errorf("expected 3, got %d", count)
		}
	})

	t.Run("with x-death fallback", func(t *testing.T) {
		d := NewDelivery(amqp.Delivery{
			Headers: amqp.Table{
				"x-death": []any{
					amqp.Table{
						"count": int64(2),
						"queue": "test-queue",
					},
				},
			},
		})
		if count := d.RetryCount(); count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	})
}

func TestDeliveryWrapper_Urgency(t *testing.T) {
	t.Run("default when missing", func(t *testing.T) {
		d := NewDelivery(amqp.Delivery{})
		if u := d.Urgency(); u != 1 {
			t.Errorf("expected default UrgencyNormal (1), got %d", u)
		}
	})

	t.Run("with x-urgency header", func(t *testing.T) {
		d := NewDelivery(amqp.Delivery{
			Headers: amqp.Table{"x-urgency": uint8(3)},
		})
		if u := d.Urgency(); u != 3 {
			t.Errorf("expected UrgencyRealtime (3), got %d", u)
		}
	})
}
