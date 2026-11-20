package exception

import (
	"errors"
	"net/http"
)

var (
	ErrFinancialReportForbidden = &AppError{
		Code: "FINANCIAL_REPORT_FORBIDDEN", Message: "only active group members can access financial reports",
		HTTPStatus: http.StatusForbidden, Err: errors.New("financial report access is forbidden"),
	}
	ErrFinancialAmountOutOfRange = &AppError{
		Code: "FINANCIAL_AMOUNT_OUT_OF_RANGE", Message: "financial balance exceeds the supported range",
		HTTPStatus: http.StatusUnprocessableEntity, Err: errors.New("financial amount out of range"),
	}
	ErrSettlementPlanTimeout = &AppError{
		Code: "SETTLEMENT_PLAN_TIMEOUT", Message: "the minimum settlement plan could not be calculated in time",
		HTTPStatus: http.StatusServiceUnavailable, Err: errors.New("minimum settlement plan timed out"),
	}
)
