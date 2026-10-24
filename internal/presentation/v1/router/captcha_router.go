package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCaptchaRoutes registers public captcha endpoints.
func RegisterCaptchaRoutes(rg *gin.RouterGroup, handler *handler.CaptchaHandler) {
	captchaGroup := rg.Group("/captcha")
	{
		captchaGroup.POST("/generate", handler.Generate)
		captchaGroup.POST("/verify", handler.Verify)
	}
}
