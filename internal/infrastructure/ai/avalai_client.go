package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
)

type AvalAIConfig struct {
	APIKey  string
	BaseURL string // مثلا https://api.avalai.ir/v1
	Model   string // مثلا gpt-4o-mini
}

type avalaiClient struct {
	cfg  AvalAIConfig
	http *http.Client
}

func NewAvalAIClient(cfg AvalAIConfig) adapter.LLMClient {
	return &avalaiClient{
		cfg:  cfg,
		http: &http.Client{Timeout: 60 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string            `json:"model"`
	Messages       []chatMessage     `json:"messages"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_tokens,omitempty"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

var errRetryable = errors.New("retryable status")

func (c *avalaiClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature:    0.4,
		MaxTokens:      2000,
		ResponseFormat: map[string]string{"type": "json_object"},
	})
	if err != nil {
		return "", err
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(1500 * time.Millisecond):
			}
		}
		out, err := c.do(ctx, body)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !errors.Is(err, errRetryable) {
			break
		}
	}
	return "", lastErr
}

func (c *avalaiClient) do(ctx context.Context, body []byte) (string, error) {
	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errRetryable, err)
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))

	if res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500 {
		return "", fmt.Errorf("%w: status %d", errRetryable, res.StatusCode)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("avalai status %d: %s", res.StatusCode, string(raw))
	}

	var cr chatResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return "", fmt.Errorf("decode avalai response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", errors.New("avalai: empty choices")
	}
	return cr.Choices[0].Message.Content, nil
}
