package DB

import (
	"context"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/finance"
)

type ExpenseReportShare struct {
	finance.Transfer
	ShareID      int64
	ExpenseID    int64
	ExpenseTitle string
	OccurredAt   *time.Time
}

type FinancialBalanceRepository interface {
	LoadBalanceData(ctx context.Context, groupID int64) (finance.BalanceData, error)
}

type FinancialReportRepository interface {
	FinancialBalanceRepository
	FindGroup(ctx context.Context, groupID int64) (*entity.Group, error)
	FindViewerMemberID(ctx context.Context, groupID, userID int64) (int64, error)
	ListMembers(ctx context.Context, groupID int64, ids []int64) ([]entity.GroupMember, error)
	ListExpenseShares(ctx context.Context, groupID, fromMemberID, toMemberID int64, limit, offset int) ([]ExpenseReportShare, int64, error)
}
