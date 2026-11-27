package postgres

import (
	"context"
	"errors"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainDB "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletRepository struct {
	db database.Database
}

func NewWalletRepository(db database.Database) domainDB.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *walletRepository) LockAccount(ctx context.Context, userID int64) (*entity.User, error) {
	var account entity.User
	err := r.getDB(ctx).Select("id", "status", "wallet").Clauses(clause.Locking{Strength: "UPDATE"}).First(&account, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrUserNotFound
	}
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &account, nil
}

func (r *walletRepository) CreateTransaction(ctx context.Context, transaction *entity.WalletTransaction) error {
	if err := r.getDB(ctx).Create(transaction).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *walletRepository) FindByIdempotencyKey(ctx context.Context, userID int64, key string) (*entity.WalletTransaction, error) {
	var transaction entity.WalletTransaction
	err := r.getDB(ctx).Where("user_id = ? AND idempotency_key = ?", userID, key).Take(&transaction).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &transaction, nil
}

func (r *walletRepository) FindTransaction(ctx context.Context, userID, transactionID int64) (*entity.WalletTransaction, error) {
	var transaction entity.WalletTransaction
	err := r.getDB(ctx).Where("user_id = ? AND id = ?", userID, transactionID).Take(&transaction).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrWalletTransactionNotFound
	}
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &transaction, nil
}

func (r *walletRepository) ListTransactions(ctx context.Context, userID int64, transactionType enum.WalletTransactionType, limit, offset int) ([]entity.WalletTransaction, int64, error) {
	query := r.getDB(ctx).Model(&entity.WalletTransaction{}).Where("user_id = ?", userID)
	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	var transactions []entity.WalletTransaction
	if err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return transactions, total, nil
}

func (r *walletRepository) Totals(ctx context.Context, userID int64) (domainDB.WalletTotals, error) {
	var totals domainDB.WalletTotals
	err := r.getDB(ctx).Model(&entity.WalletTransaction{}).
		Select(`COALESCE(SUM(token_amount) FILTER (WHERE type = ?), 0) AS purchased_tokens,
			COALESCE(SUM(token_amount) FILTER (WHERE type = ?), 0) AS spent_tokens,
			COALESCE(SUM(token_amount) FILTER (WHERE type = ?), 0) AS rewarded_tokens`,
			enum.WalletTransactionPurchase, enum.WalletTransactionSpend, enum.WalletTransactionReward).
		Where("user_id = ? AND status = ?", userID, enum.WalletTransactionSucceeded).Scan(&totals).Error
	if err != nil {
		return domainDB.WalletTotals{}, exception.Wrap(exception.ErrInternal, err)
	}
	return totals, nil
}

func (r *walletRepository) ProfileCompletion(ctx context.Context, userID int64) (entity.ProfileCompletion, error) {
	var completion entity.ProfileCompletion
	result := r.getDB(ctx).Raw(`
		SELECT steps.*, EXISTS (
			SELECT 1 FROM wallet_transactions
			WHERE user_id = account.id AND type = ? AND idempotency_key = ?
		) AS reward_granted
		FROM users AS account
		CROSS JOIN LATERAL profile_completion_steps(account) AS steps
		WHERE account.id = ? AND account.deleted_at IS NULL`,
		enum.WalletTransactionReward, entity.ProfileCompletionRewardKey, userID).Scan(&completion)
	if result.Error != nil {
		return entity.ProfileCompletion{}, exception.Wrap(exception.ErrInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ProfileCompletion{}, exception.ErrUserNotFound
	}
	return completion, nil
}
