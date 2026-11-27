package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

type WalletTransaction struct {
	ID                 int64 `gorm:"primaryKey;autoIncrement"`
	UserID             int64
	Type               enum.WalletTransactionType
	Status             enum.WalletTransactionStatus
	Description        string
	TokenAmount        int64
	BalanceAfter       int64
	AmountToman        *int64
	BaseUnitPriceToman *int64
	DiscountPercent    int64
	PackageID          *string
	IdempotencyKey     *string
	CreatedAt          time.Time
}

func (WalletTransaction) TableName() string { return "wallet_transactions" }

func (t WalletTransaction) TokenDelta() int64 {
	if t.Status != enum.WalletTransactionSucceeded {
		return 0
	}
	if t.Type == enum.WalletTransactionSpend {
		return -t.TokenAmount
	}
	return t.TokenAmount
}
