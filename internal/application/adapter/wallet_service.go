package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type WalletService interface {
	Summary(ctx context.Context, userID int64) (*dto.WalletSummaryResponse, error)
	ProfileCompletion(ctx context.Context, userID int64) (*dto.ProfileCompletionResponse, error)
	Offers(ctx context.Context, userID int64) (*dto.WalletOffersResponse, error)
	Purchase(ctx context.Context, userID int64, idempotencyKey string, request dto.WalletPurchaseRequest) (*dto.WalletTransactionResponse, bool, error)
	ListTransactions(ctx context.Context, userID int64, transactionType enum.WalletTransactionType, page, limit, offset int) (*dto.PaginatedResponse[dto.WalletTransactionResponse], error)
	Transaction(ctx context.Context, userID, transactionID int64) (*dto.WalletTransactionResponse, error)
}
