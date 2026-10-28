// Package tokencontext provides typed accessors for storing and retrieving
// verified JWT token claims on *gin.Context and context.Context.
package tokencontext

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

const claimsKey = "auth_claims"

type contextKey struct{}

// SetClaims stores verified JWT claims on the Gin context (used in auth middleware).
func SetClaims(c *gin.Context, claims *jwt.Claims) {
	c.Set(claimsKey, claims)
}

// GetClaims retrieves verified JWT claims from the Gin context.
func GetClaims(c *gin.Context) (*jwt.Claims, bool) {
	val, exists := c.Get(claimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*jwt.Claims)
	return claims, ok
}

// GetUserID extracts the authenticated user's ID from the Gin context.
func GetUserID(c *gin.Context) (int64, bool) {
	claims, ok := GetClaims(c)
	if !ok || claims == nil {
		return 0, false
	}
	return claims.UserID, true
}

// GetUserRole extracts the authenticated user's role from the Gin context.
func GetUserRole(c *gin.Context) (enum.UserRole, bool) {
	claims, ok := GetClaims(c)
	if !ok || claims == nil {
		return "", false
	}
	return claims.Role, true
}

// InjectClaims attaches verified JWT claims to a standard Go context.Context.
func InjectClaims(ctx context.Context, claims *jwt.Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, claims)
}

// ExtractClaims retrieves verified JWT claims from a standard Go context.Context.
func ExtractClaims(ctx context.Context) (*jwt.Claims, bool) {
	val := ctx.Value(contextKey{})
	if val == nil {
		return nil, false
	}
	claims, ok := val.(*jwt.Claims)
	return claims, ok
}

// ExtractUserID retrieves the authenticated user's ID from a context.Context.
func ExtractUserID(ctx context.Context) (int64, bool) {
	claims, ok := ExtractClaims(ctx)
	if !ok || claims == nil {
		return 0, false
	}
	return claims.UserID, true
}

// ExtractUserRole retrieves the authenticated user's role from a context.Context.
func ExtractUserRole(ctx context.Context) (enum.UserRole, bool) {
	claims, ok := ExtractClaims(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.Role, true
}
