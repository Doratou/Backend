package websocket

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
	tokenQueryParamKey      = "token"
)

// JWTAuthenticator validates WebSocket upgrade requests using the existing JWT TokenMaker
// and Redis access-token revocation blacklist.
type JWTAuthenticator struct {
	tokenMaker jwt.TokenMaker
	tokenRepo  cache.AuthnTokenRepository
}

// NewJWTAuthenticator creates a new JWTAuthenticator for WebSocket handshakes.
func NewJWTAuthenticator(tokenMaker jwt.TokenMaker, tokenRepo cache.AuthnTokenRepository) *JWTAuthenticator {
	return &JWTAuthenticator{
		tokenMaker: tokenMaker,
		tokenRepo:  tokenRepo,
	}
}

// Authenticate extracts a Bearer JWT from the Authorization header or `?token=` query parameter,
// verifies signature and expiration, ensures it is an access token, and checks the Redis blacklist.
func (a *JWTAuthenticator) Authenticate(ctx context.Context, r *http.Request) (*domain_ws.Identity, error) {
	if a.tokenMaker == nil {
		return nil, exception.ErrInternal
	}

	tokenString := extractBearerToken(r)
	if tokenString == "" {
		return nil, exception.ErrMissingToken
	}

	claims, err := a.tokenMaker.VerifyToken(tokenString)
	if err != nil {
		if errors.Is(err, exception.ErrTokenExpired) {
			return nil, exception.ErrTokenExpired
		}
		return nil, exception.ErrTokenInvalid
	}

	if claims.TokenType != jwt.TokenTypeAccess {
		return nil, exception.ErrTokenInvalid
	}

	if a.tokenRepo != nil {
		isBlacklisted, err := a.tokenRepo.IsBlacklisted(ctx, claims.ID)
		if err != nil {
			return nil, exception.ErrInternal
		}
		if isBlacklisted {
			return nil, exception.ErrTokenRevoked
		}
	}

	return &domain_ws.Identity{
		UserID:  claims.UserID,
		Role:    string(claims.Role),
		TokenID: claims.ID,
	}, nil
}

func extractBearerToken(r *http.Request) string {
	header := r.Header.Get(authorizationHeaderKey)
	if header != "" {
		fields := strings.Fields(header)
		if len(fields) == 2 && strings.EqualFold(fields[0], authorizationTypeBearer) {
			return strings.TrimSpace(fields[1])
		}
	}

	if r.URL != nil {
		if q := strings.TrimSpace(r.URL.Query().Get(tokenQueryParamKey)); q != "" {
			return q
		}
	}

	return ""
}
