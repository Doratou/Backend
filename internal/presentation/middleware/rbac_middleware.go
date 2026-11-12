package middleware

import (
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

// RequireRole creates a Gin middleware that ensures the authenticated user possesses
// at least one of the specified allowed system roles (e.g. enum.UserRoleAdmin).
// It must be placed after the Auth() middleware in the handler chain.
func RequireRole(allowedRoles ...enum.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := tokencontext.GetUserRole(c)
		if !ok {
			respondError(c, exception.ErrMissingToken)
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		respondError(c, exception.ErrForbiddenRole)
	}
}

// RBAC is a descriptive alias for RequireRole.
func RBAC(allowedRoles ...enum.UserRole) gin.HandlerFunc {
	return RequireRole(allowedRoles...)
}
