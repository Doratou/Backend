package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

// TokenPriceHandler handles HTTP endpoints for updating platform token price.
type TokenPriceHandler struct {
	service adapter.TokenPriceService
}

// NewTokenPriceHandler creates a new TokenPriceHandler.
func NewTokenPriceHandler(service adapter.TokenPriceService) *TokenPriceHandler {
	return &TokenPriceHandler{service: service}
}

// UpdatePrice handles PUT /api/v1/admin/token-price.
func (h *TokenPriceHandler) UpdatePrice(c *gin.Context) {
	adminUserID, ok := tokencontext.GetUserID(c)
	if !ok || adminUserID <= 0 {
		fail(c, exception.ErrMissingToken)
		return
	}

	var req dto.UpdateTokenPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}

	res, err := h.service.UpdatePrice(c.Request.Context(), adminUserID, req)
	if err != nil {
		fail(c, err)
		return
	}

	success(c, http.StatusOK, res)
}
