package websocket

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewEnvelope_AndDecodeData(t *testing.T) {
	payload := map[string]any{"hello": "world", "count": 3}
	env, err := NewEnvelope("test.event", payload)
	if err != nil {
		t.Fatalf("NewEnvelope returned unexpected error: %v", err)
	}
	env = env.WithID("msg-1").WithChannel("room:alpha")

	if err := env.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if env.ID != "msg-1" || env.Channel != "room:alpha" || env.Event != "test.event" {
		t.Fatalf("unexpected envelope metadata: %+v", env)
	}
	if env.Timestamp.IsZero() {
		t.Fatal("expected non-zero UTC timestamp")
	}

	var decoded struct {
		Hello string `json:"hello"`
		Count int    `json:"count"`
	}
	if err := env.DecodeData(&decoded); err != nil {
		t.Fatalf("DecodeData failed: %v", err)
	}
	if decoded.Hello != "world" || decoded.Count != 3 {
		t.Fatalf("unexpected decoded data: %+v", decoded)
	}
}

func TestNewEnvelope_ValidationErrors(t *testing.T) {
	if _, err := NewEnvelope("   ", nil); !errors.Is(err, ErrInvalidEnvelope) {
		t.Fatalf("expected ErrInvalidEnvelope for blank event, got %v", err)
	}

	if _, err := NewEnvelope("ok", []byte("{invalid")); !errors.Is(err, ErrInvalidEnvelope) {
		t.Fatalf("expected ErrInvalidEnvelope for invalid raw JSON bytes, got %v", err)
	}

	raw := json.RawMessage(`{"k":"v"}`)
	env, err := NewEnvelope("raw.event", raw)
	if err != nil {
		t.Fatalf("unexpected error for json.RawMessage: %v", err)
	}
	if string(env.Data) != string(raw) {
		t.Fatalf("expected raw data %s, got %s", raw, env.Data)
	}
}
