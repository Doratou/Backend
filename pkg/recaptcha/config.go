package recaptcha

import (
	"net/http"
	"time"
)

const (
	// DefaultVerifyURL is the official Google reCAPTCHA v3 siteverify endpoint.
	DefaultVerifyURL = "https://www.google.com/recaptcha/api/siteverify"

	// FallbackVerifyURL is the alternative global endpoint (recaptcha.net) for regions where google.com is restricted.
	FallbackVerifyURL = "https://recaptcha.net/recaptcha/api/siteverify"

	// DefaultMinScore is Google's recommended baseline human threshold for v3 (ranges from 0.0 to 1.0).
	DefaultMinScore = 0.5

	// DefaultTimeout is the HTTP request timeout for verification calls.
	DefaultTimeout = 5 * time.Second
)

// Config holds runtime configuration options for the reCAPTCHA v3 client.
type Config struct {
	// SecretKey is the shared secret key between your site and reCAPTCHA.
	SecretKey string

	// MinScore is the acceptable score threshold (0.0 to 1.0). Defaults to 0.5.
	MinScore float64

	// Timeout specifies the HTTP request timeout. Defaults to 5s.
	Timeout time.Duration

	// VerifyURL allows customizing the verification endpoint. Defaults to DefaultVerifyURL.
	VerifyURL string

	// HTTPClient allows injecting a custom *http.Client for proxy or custom transport.
	HTTPClient *http.Client
}

// DefaultConfig provides a pre-configured Config with recommended production defaults.
func DefaultConfig(secretKey string) Config {
	return Config{
		SecretKey: secretKey,
		MinScore:  DefaultMinScore,
		Timeout:   DefaultTimeout,
		VerifyURL: DefaultVerifyURL,
	}
}
