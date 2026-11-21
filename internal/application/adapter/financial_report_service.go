package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type FinancialReportService interface {
	ExpenseShares(ctx context.Context, userID, groupID int64, owedByMe bool, pagination dto.ExpensePaginationRequest) (*dto.ExpenseShareReport, error)
	ExpenseGraph(ctx context.Context, userID, groupID int64) (*dto.ExpenseGraphReport, error)
	SettlementGraph(ctx context.Context, userID, groupID int64) (*dto.SettlementGraphReport, error)
	MySettlements(ctx context.Context, userID, groupID int64) (*dto.MySettlementReport, error)
}
