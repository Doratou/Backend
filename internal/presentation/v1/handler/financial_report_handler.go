package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type FinancialReportHandler struct {
	reports adapter.FinancialReportService
}

func NewFinancialReportHandler(reports adapter.FinancialReportService) *FinancialReportHandler {
	return &FinancialReportHandler{reports: reports}
}

func (h *FinancialReportHandler) OwedByMe(c *gin.Context) {
	h.expenseShares(c, true)
}

func (h *FinancialReportHandler) OwedToMe(c *gin.Context) {
	h.expenseShares(c, false)
}

func (h *FinancialReportHandler) expenseShares(c *gin.Context, owedByMe bool) {
	userID, groupID, ok := financialReportIdentity(c)
	if !ok {
		return
	}
	var req dto.ExpensePaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	report, err := h.reports.ExpenseShares(c.Request.Context(), userID, groupID, owedByMe, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, report)
}

func (h *FinancialReportHandler) ExpenseGraph(c *gin.Context) {
	userID, groupID, ok := financialReportIdentity(c)
	if !ok {
		return
	}
	report, err := h.reports.ExpenseGraph(c.Request.Context(), userID, groupID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, report)
}

func (h *FinancialReportHandler) SettlementGraph(c *gin.Context) {
	userID, groupID, ok := financialReportIdentity(c)
	if !ok {
		return
	}
	report, err := h.reports.SettlementGraph(c.Request.Context(), userID, groupID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, report)
}

func (h *FinancialReportHandler) MySettlements(c *gin.Context) {
	userID, groupID, ok := financialReportIdentity(c)
	if !ok {
		return
	}
	report, err := h.reports.MySettlements(c.Request.Context(), userID, groupID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, report)
}

func financialReportIdentity(c *gin.Context) (int64, int64, bool) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return 0, 0, false
	}
	if userID <= 0 {
		fail(c, exception.ErrTokenInvalid)
		return 0, 0, false
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil || groupID <= 0 {
		fail(c, exception.ErrBadRequest)
		return 0, 0, false
	}
	return userID, groupID, true
}
