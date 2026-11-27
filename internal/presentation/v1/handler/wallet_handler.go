package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	wallet adapter.WalletService
}

func NewWalletHandler(wallet adapter.WalletService) *WalletHandler {
	return &WalletHandler{wallet: wallet}
}

func (h *WalletHandler) Summary(c *gin.Context) {
	userID, ok := walletIdentity(c)
	if !ok {
		return
	}
	response, err := h.wallet.Summary(c.Request.Context(), userID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

func (h *WalletHandler) Offers(c *gin.Context) {
	userID, ok := walletIdentity(c)
	if !ok {
		return
	}
	response, err := h.wallet.Offers(c.Request.Context(), userID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

func (h *WalletHandler) Purchase(c *gin.Context) {
	userID, ok := walletIdentity(c)
	if !ok {
		return
	}
	var request dto.WalletPurchaseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	response, created, err := h.wallet.Purchase(c.Request.Context(), userID, c.GetHeader("Idempotency-Key"), request)
	if err != nil {
		fail(c, err)
		return
	}
	if created {
		success(c, http.StatusCreated, response)
		return
	}
	success(c, http.StatusOK, response)
}

func (h *WalletHandler) ListTransactions(c *gin.Context) {
	userID, ok := walletIdentity(c)
	if !ok {
		return
	}
	page, limit, offset, err := parsePagination(c)
	if err != nil {
		fail(c, err)
		return
	}
	kind := enum.WalletTransactionType(c.Query("type"))
	response, err := h.wallet.ListTransactions(c.Request.Context(), userID, kind, page, limit, offset)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

func (h *WalletHandler) Transaction(c *gin.Context) {
	userID, ok := walletIdentity(c)
	if !ok {
		return
	}
	transactionID, err := parseIDParam(c, "transactionID")
	if err != nil {
		fail(c, err)
		return
	}
	response, err := h.wallet.Transaction(c.Request.Context(), userID, transactionID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, response)
}

func walletIdentity(c *gin.Context) (int64, bool) {
	c.Header("Cache-Control", "no-store")
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return 0, false
	}
	if userID <= 0 {
		fail(c, exception.ErrTokenInvalid)
		return 0, false
	}
	return userID, true
}
