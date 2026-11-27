package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

type WalletSummaryResponse struct {
	Wallet              int64                      `json:"wallet"`
	Currency            string                     `json:"currency"`
	PurchasedTokens     int64                      `json:"purchased_tokens"`
	SpentTokens         int64                      `json:"spent_tokens"`
	RewardedTokens      int64                      `json:"rewarded_tokens"`
	TokenPriceToman     *int64                     `json:"token_price_toman"`
	PriceUpdatedAt      *time.Time                 `json:"price_updated_at"`
	EstimatedValueToman *int64                     `json:"estimated_value_toman"`
	ProfileCompletion   *ProfileCompletionResponse `json:"profile_completion"`
}

type ProfileCompletionSteps struct {
	FullName          bool `json:"full_name"`
	ProfilePhoto      bool `json:"profile_photo"`
	VerifiedEmail     bool `json:"verified_email"`
	VerifiedPhone     bool `json:"verified_phone"`
	TelegramConnected bool `json:"telegram_connected"`
}

type ProfileCompletionResponse struct {
	Steps          ProfileCompletionSteps `json:"steps"`
	CompletedSteps int                    `json:"completed_steps"`
	TotalSteps     int                    `json:"total_steps"`
	Progress       int                    `json:"progress_percent"`
	RewardTokens   int64                  `json:"reward_tokens"`
	RewardGranted  bool                   `json:"reward_granted"`
}

type WalletPackageResponse struct {
	ID              string `json:"id"`
	TokenAmount     int64  `json:"token_amount"`
	AmountToman     int64  `json:"amount_toman"`
	DiscountPercent int64  `json:"discount_percent"`
	IsPopular       bool   `json:"is_popular"`
}

type WalletOffersResponse struct {
	Currency        string                  `json:"currency"`
	TokenPriceToman int64                   `json:"token_price_toman"`
	PriceUpdatedAt  time.Time               `json:"price_updated_at"`
	CustomMinTokens int64                   `json:"custom_min_tokens"`
	CustomMaxTokens int64                   `json:"custom_max_tokens"`
	Packages        []WalletPackageResponse `json:"packages"`
}

type WalletPurchaseRequest struct {
	PackageID           *string `json:"package_id"`
	TokenAmount         *int64  `json:"token_amount"`
	ExpectedAmountToman *int64  `json:"expected_amount_toman" binding:"omitempty,gt=0"`
}

type WalletTransactionResponse struct {
	ID                 int64                        `json:"id"`
	Type               enum.WalletTransactionType   `json:"type"`
	Status             enum.WalletTransactionStatus `json:"status"`
	Description        string                       `json:"description"`
	TokenAmount        int64                        `json:"token_amount"`
	TokenDelta         int64                        `json:"token_delta"`
	BalanceAfter       int64                        `json:"balance_after"`
	Currency           string                       `json:"currency"`
	AmountToman        *int64                       `json:"amount_toman"`
	BaseUnitPriceToman *int64                       `json:"base_unit_price_toman"`
	DiscountPercent    int64                        `json:"discount_percent"`
	PackageID          *string                      `json:"package_id"`
	TrackingCode       *string                      `json:"tracking_code"`
	PaymentMode        *string                      `json:"payment_mode"`
	RewardCode         enum.WalletRewardCode        `json:"reward_code,omitempty"`
	CreatedAt          time.Time                    `json:"created_at"`
}
