package router

import (
	"io/fs"
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/Doratou/Backend/internal/presentation/middleware"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// Dependencies bundles all handlers and presentation dependencies needed to wire routes.
type Dependencies struct {
	AttachmentHandler  *handler.AttachmentHandler
	AuthHandler        *handler.AuthHandler
	UserHandler        *handler.UserHandler
	CaptchaHandler     *handler.CaptchaHandler
	AdminReportHandler *handler.AdminReportHandler
	FriendshipHandler  *handler.FriendshipHandler
	GroupHandler       *handler.GroupHandler
	PollHandler        *handler.PollHandler
	LuckyDrawHandler   *handler.LuckyDrawHandler
	InviteHandler      *handler.InviteHandler
	CaptchaService     adapter.CaptchaService
	CaptchaEnabled     bool
	TokenMaker         jwt.TokenMaker
	TokenRepo          cache.AuthnTokenRepository
	EmailFontFiles     fs.FS
}

// New builds the fully configured Gin engine with global middlewares and versioned routes.
func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middleware.CORS())
	if deps.EmailFontFiles != nil {
		RegisterEmailFontRoutes(engine, deps.EmailFontFiles)
	}

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	var authMiddleware gin.HandlerFunc
	var optionalAuthMiddleware gin.HandlerFunc
	if deps.TokenMaker != nil && deps.TokenRepo != nil {
		authMiddleware = middleware.Auth(deps.TokenMaker, deps.TokenRepo)
		optionalAuthMiddleware = middleware.OptionalAuth(deps.TokenMaker, deps.TokenRepo)
	}

	var captchaMiddleware gin.HandlerFunc
	if deps.CaptchaService != nil {
		captchaMiddleware = middleware.CaptchaGuard(deps.CaptchaService, deps.CaptchaEnabled)
	}

	v1 := engine.Group("/api/v1")
	RegisterAuthRoutes(v1, deps.AuthHandler, authMiddleware, captchaMiddleware)
	if deps.UserHandler != nil {
		RegisterUserRoutes(v1, deps.UserHandler, authMiddleware)
	}
	RegisterCaptchaRoutes(v1, deps.CaptchaHandler)
	RegisterAttachmentRoutes(v1, deps.AttachmentHandler, authMiddleware, optionalAuthMiddleware)
	RegisterAdminReportRoutes(v1, deps.AdminReportHandler, authMiddleware)

	if deps.FriendshipHandler != nil {
		RegisterFriendshipRoutes(v1, deps.FriendshipHandler, authMiddleware)
	}
	if deps.GroupHandler != nil {
		RegisterGroupRoutes(v1, deps.GroupHandler, authMiddleware)
		if deps.PollHandler != nil {
			RegisterPollRoutes(v1, deps.PollHandler, authMiddleware)
		}
	}
	if deps.InviteHandler != nil {
		RegisterInviteRoutes(v1, deps.InviteHandler, authMiddleware)
	}
	if deps.LuckyDrawHandler != nil {
		RegisterLuckyDrawRoutes(v1, deps.LuckyDrawHandler, authMiddleware)
	}

	return engine
}
