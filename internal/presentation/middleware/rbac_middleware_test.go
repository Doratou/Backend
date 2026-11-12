package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/Doratou/Backend/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRBACRouter(userRole *enum.UserRole, allowedRoles ...enum.UserRole) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if userRole != nil {
			tokencontext.SetClaims(c, &jwt.Claims{
				UserID: 1,
				Role:   *userRole,
			})
		}
	})
	r.GET("/admin-only", middleware.RequireRole(allowedRoles...), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "granted"})
	})
	return r
}

func TestRequireRole_Allowed(t *testing.T) {
	role := enum.UserRoleAdmin
	router := setupRBACRouter(&role, enum.UserRoleAdmin)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestRequireRole_ForbiddenRole(t *testing.T) {
	role := enum.UserRoleUser
	router := setupRBACRouter(&role, enum.UserRoleAdmin)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestRequireRole_MissingClaims(t *testing.T) {
	router := setupRBACRouter(nil, enum.UserRoleAdmin)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestRequireRole_MultipleAllowedRoles(t *testing.T) {
	userRole := enum.UserRoleUser
	router := setupRBACRouter(&userRole, enum.UserRoleAdmin, enum.UserRoleUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}
