package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type WalletTotals struct {
	PurchasedTokens int64
	SpentTokens     int64
	RewardedTokens  int64
}

type WalletRepository interface {
	LockAccount(ctx context.Context, userID int64) (*entity.User, error)
	CreateTransaction(ctx context.Context, transaction *entity.WalletTransaction) error
	FindByIdempotencyKey(ctx context.Context, userID int64, key string) (*entity.WalletTransaction, error)
	FindTransaction(ctx context.Context, userID, transactionID int64) (*entity.WalletTransaction, error)
	ListTransactions(ctx context.Context, userID int64, transactionType enum.WalletTransactionType, limit, offset int) ([]entity.WalletTransaction, int64, error)
	Totals(ctx context.Context, userID int64) (WalletTotals, error)
	ProfileCompletion(ctx context.Context, userID int64) (entity.ProfileCompletion, error)
}
