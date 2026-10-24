package middleware

import (
	"errors"
	"strings"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/gin-gonic/gin"
)

const (
	HeaderCaptchaID   = "X-Captcha-ID"
	HeaderCaptchaCode = "X-Captcha-Code"
)

// CaptchaGuard verifies that incoming requests on sensitive endpoints contain valid single-use captcha headers.
// If enabled is false (bypass mode for dev/test/bruno), the verification is skipped.
func CaptchaGuard(captchaService adapter.CaptchaService, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		captchaID := strings.TrimSpace(c.GetHeader(HeaderCaptchaID))
		captchaCode := strings.TrimSpace(c.GetHeader(HeaderCaptchaCode))

		if captchaID == "" || captchaCode == "" {
			respondError(c, exception.ErrCaptchaRequired)
			return
		}

		valid, err := captchaService.Verify(c.Request.Context(), captchaID, captchaCode)
		if err != nil {
			var appErr *exception.AppError
			if errors.As(err, &appErr) {
				respondError(c, appErr)
			} else {
				respondError(c, exception.ErrCaptchaInvalid)
			}
			return
		}

		if !valid {
			respondError(c, exception.ErrCaptchaInvalid)
			return
		}

		c.Next()
	}
}
