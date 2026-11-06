package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// GoogleClaims holds the OIDC claims returned by Google.
// picture is intentionally omitted per requirements — avatar is managed via the attachment system.
type GoogleClaims struct {
	Sub           string `json:"sub"` // unique Google user ID (stable identifier)
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

// GoogleConfig holds OAuth2 client credentials and callback settings for Google.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// GoogleProvider defines the interface for Google OIDC interactions.
type GoogleProvider interface {
	AuthURL(state, nonce string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	VerifyIDToken(ctx context.Context, token *oauth2.Token, nonce string) (*GoogleClaims, error)
}

type googleProvider struct {
	provider *oidc.Provider
	config   oauth2.Config
	verifier *oidc.IDTokenVerifier
}

// NewGoogleProvider creates a GoogleProvider by fetching Google's OIDC discovery document.
func NewGoogleProvider(ctx context.Context, cfg GoogleConfig) (GoogleProvider, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("connect to google oidc provider: %w", err)
	}

	config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	return &googleProvider{provider: provider, config: config, verifier: verifier}, nil
}

// AuthURL returns the Google consent screen URL with state and nonce embedded.
func (g *googleProvider) AuthURL(state, nonce string) string {
	return g.config.AuthCodeURL(state, oidc.Nonce(nonce))
}

// Exchange swaps an authorization code for an OAuth2 token set (server-to-server).
func (g *googleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

// VerifyIDToken validates the id_token signature, checks the nonce, and returns claims.
func (g *googleProvider) VerifyIDToken(ctx context.Context, token *oauth2.Token, nonce string) (*GoogleClaims, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id_token missing from token response")
	}

	idToken, err := g.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify id_token: %w", err)
	}

	if idToken.Nonce != nonce {
		return nil, fmt.Errorf("nonce mismatch: possible replay attack")
	}

	var claims GoogleClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parse id_token claims: %w", err)
	}

	return &claims, nil
}

// RandString generates a cryptographically secure random base64url string of n bytes.
func RandString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
