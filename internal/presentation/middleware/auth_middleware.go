package middleware

import (
	"errors"
	"strings"

	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
)

// Auth creates a Gin middleware that verifies JWT access tokens on protected routes.
func Auth(tokenMaker jwt.TokenMaker, tokenRepo cache.AuthnTokenRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authorizationHeaderKey)
		if header == "" {
			respondError(c, exception.ErrMissingToken)
			return
		}

		fields := strings.Fields(header)
		if len(fields) != 2 || !strings.EqualFold(fields[0], authorizationTypeBearer) {
			respondError(c, exception.ErrMissingToken)
			return
		}

		tokenString := fields[1]
		claims, err := tokenMaker.VerifyToken(tokenString)
		if err != nil {
			if errors.Is(err, exception.ErrTokenExpired) {
				respondError(c, exception.ErrTokenExpired)
				return
			}
			respondError(c, exception.ErrTokenInvalid)
			return
		}

		if claims.TokenType != jwt.TokenTypeAccess {
			respondError(c, exception.ErrTokenInvalid)
			return
		}

		isBlacklisted, err := tokenRepo.IsBlacklisted(c.Request.Context(), claims.ID)
		if err != nil {
			respondError(c, exception.ErrInternal)
			return
		}
		if isBlacklisted {
			respondError(c, exception.ErrTokenRevoked)
			return
		}

		tokencontext.SetClaims(c, claims)
		c.Next()
	}
}

func respondError(c *gin.Context, appErr *exception.AppError) {
	c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{
		"success": false,
		"code":    appErr.Code,
		"error":   appErr.Message,
	})
}
