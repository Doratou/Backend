package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidSecretKey = errors.New("jwt secret key must be at least 32 bytes")
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type TokenMaker interface {
	CreateToken(userID int64, role enum.UserRole, tokenType TokenType, duration time.Duration) (string, *Claims, error)
	VerifyToken(tokenString string) (*Claims, error)
}

type Claims struct {
	UserID    int64         `json:"user_id"`
	Role      enum.UserRole `json:"role"`
	TokenType TokenType     `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTMaker struct {
	secretKey []byte
}

func NewJWTMaker(secretKey string) (TokenMaker, error) {
	if len(secretKey) < 32 {
		return nil, ErrInvalidSecretKey
	}
	return &JWTMaker{
		secretKey: []byte(secretKey),
	}, nil
}

func (maker *JWTMaker) CreateToken(userID int64, role enum.UserRole, tokenType TokenType, duration time.Duration) (string, *Claims, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return "", nil, exception.Wrap(exception.ErrInternal, err)
	}

	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(maker.secretKey)
	if err != nil {
		return "", nil, exception.Wrap(exception.ErrInternal, err)
	}

	return tokenString, claims, nil
}

func (maker *JWTMaker) VerifyToken(tokenString string) (*Claims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, exception.ErrTokenInvalid
		}
		return maker.secretKey, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, exception.ErrTokenExpired
		}
		return nil, exception.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, exception.ErrTokenInvalid
	}

	return claims, nil
}
