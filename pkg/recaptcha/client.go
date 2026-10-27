package recaptcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// ErrMissingToken is returned when an empty token is passed for verification.
	ErrMissingToken = errors.New("recaptcha: token cannot be empty")

	// ErrMissingSecret is returned when the secret key is not configured.
	ErrMissingSecret = errors.New("recaptcha: secret key cannot be empty")

	// ErrVerificationFailed is returned when Google reCAPTCHA marks the response as unsuccessful.
	ErrVerificationFailed = errors.New("recaptcha: verification failed by provider")

	// ErrActionMismatch is returned when the response action does not match expected action.
	ErrActionMismatch = errors.New("recaptcha: action mismatch")

	// ErrScoreTooLow is returned when the assessed score is below the configured MinScore threshold.
	ErrScoreTooLow = errors.New("recaptcha: assessed score is below acceptable threshold")
)

// Response represents the verification payload returned by Google reCAPTCHA v3.
type Response struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

// Time parses ChallengeTS into a time.Time if available.
func (r *Response) Time() (time.Time, error) {
	if r.ChallengeTS == "" {
		return time.Time{}, errors.New("recaptcha: challenge_ts is empty")
	}
	return time.Parse(time.RFC3339, r.ChallengeTS)
}

// Verifier describes the capabilities for verifying reCAPTCHA v3 challenge tokens.
type Verifier interface {
	// Verify submits the token to the siteverify endpoint and returns the raw response.
	Verify(ctx context.Context, token string, remoteIP ...string) (*Response, error)

	// VerifyAction verifies the token and validates both the expected action and minimum score threshold.
	VerifyAction(ctx context.Context, token string, expectedAction string, remoteIP ...string) (bool, *Response, error)
}

// Client implements Verifier against Google reCAPTCHA v3.
type Client struct {
	secretKey  string
	minScore   float64
	verifyURL  string
	httpClient *http.Client
}

// New creates a new reCAPTCHA v3 client from the supplied Config.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.SecretKey) == "" {
		return nil, ErrMissingSecret
	}

	minScore := cfg.MinScore
	if minScore <= 0 {
		minScore = DefaultMinScore
	}

	verifyURL := strings.TrimSpace(cfg.VerifyURL)
	if verifyURL == "" {
		verifyURL = DefaultVerifyURL
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &Client{
		secretKey:  cfg.SecretKey,
		minScore:   minScore,
		verifyURL:  verifyURL,
		httpClient: httpClient,
	}, nil
}

// Verify sends the token to the reCAPTCHA siteverify API and unmarshals the response.
func (c *Client) Verify(ctx context.Context, token string, remoteIP ...string) (*Response, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrMissingToken
	}

	formData := url.Values{
		"secret":   {c.secretKey},
		"response": {token},
	}
	if len(remoteIP) > 0 && strings.TrimSpace(remoteIP[0]) != "" {
		formData.Set("remoteip", strings.TrimSpace(remoteIP[0]))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.verifyURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("recaptcha: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("recaptcha: execute request: %w", err)
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("recaptcha: read response body: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recaptcha: unexpected status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var res Response
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("recaptcha: parse response: %w", err)
	}

	if !res.Success {
		return &res, ErrVerificationFailed
	}

	return &res, nil
}

// VerifyAction verifies the token and enforces that the action matches expectedAction and the score >= MinScore.
func (c *Client) VerifyAction(ctx context.Context, token string, expectedAction string, remoteIP ...string) (bool, *Response, error) {
	res, err := c.Verify(ctx, token, remoteIP...)
	if err != nil {
		return false, res, err
	}

	if expectedAction != "" && res.Action != expectedAction {
		return false, res, ErrActionMismatch
	}

	if res.Score < c.minScore {
		return false, res, ErrScoreTooLow
	}

	return true, res, nil
}
