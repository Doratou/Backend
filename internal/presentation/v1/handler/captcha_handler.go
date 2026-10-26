package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/gin-gonic/gin"
)

// CaptchaHandler handles HTTP requests for generating and verifying captcha challenges.
type CaptchaHandler struct {
	captchaService adapter.CaptchaService
}

// NewCaptchaHandler initializes a CaptchaHandler instance.
func NewCaptchaHandler(captchaService adapter.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{
		captchaService: captchaService,
	}
}

// Generate produces a new captcha challenge and returns its base64 image along with captcha_id.
// POST /api/v1/captcha/generate or GET /api/v1/captcha/generate
func (h *CaptchaHandler) Generate(c *gin.Context) {
	resp, err := h.captchaService.Generate(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, resp)
}

// Verify validates a submitted captcha code against the expected cached answer.
// POST /api/v1/captcha/verify
func (h *CaptchaHandler) Verify(c *gin.Context) {
	var req dto.VerifyCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	valid, err := h.captchaService.Verify(c.Request.Context(), req.CaptchaID, req.Code)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, dto.VerifyCaptchaResponse{Valid: valid})
}
