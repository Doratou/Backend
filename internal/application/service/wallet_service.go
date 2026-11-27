package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainDB "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

const (
	walletCurrency        = "TOMAN"
	walletCustomMinTokens = 1
	walletCustomMaxTokens = 10000
)

type walletPackage struct {
	id       string
	tokens   int64
	discount int64
	popular  bool
}

var walletPackages = [...]walletPackage{
	{id: "TOKEN_10", tokens: 10},
	{id: "TOKEN_25", tokens: 25, discount: 8},
	{id: "TOKEN_50", tokens: 50, discount: 12, popular: true},
	{id: "TOKEN_100", tokens: 100, discount: 20},
}

type WalletServiceConfig struct {
	MockRandomEnabled  bool
	MockSuccessPercent int64
}

type walletService struct {
	repo   domainDB.WalletRepository
	users  domainDB.UserRepository
	prices domainDB.TokenPriceRepository
	write  database.TrxManager
	read   database.TrxManager
	random adapter.RandomGenerator
	config WalletServiceConfig
}

func NewWalletService(repo domainDB.WalletRepository, users domainDB.UserRepository, prices domainDB.TokenPriceRepository, write, read database.TrxManager, random adapter.RandomGenerator, config WalletServiceConfig) adapter.WalletService {
	return &walletService{repo: repo, users: users, prices: prices, write: write, read: read, random: random, config: config}
}

func (s *walletService) Summary(ctx context.Context, userID int64) (*dto.WalletSummaryResponse, error) {
	response := &dto.WalletSummaryResponse{Currency: walletCurrency}
	err := s.read.WithTransaction(ctx, func(txCtx context.Context) error {
		account, err := s.activeAccount(txCtx, userID)
		if err != nil {
			return err
		}
		totals, err := s.repo.Totals(txCtx, userID)
		if err != nil {
			return err
		}
		response.Wallet = account.Wallet
		response.PurchasedTokens = totals.PurchasedTokens
		response.SpentTokens = totals.SpentTokens
		response.RewardedTokens = totals.RewardedTokens
		response.ProfileCompletion, err = s.ProfileCompletion(txCtx, userID)
		if err != nil {
			return err
		}

		price, err := s.prices.CurrentPrice(txCtx)
		if errors.Is(err, exception.ErrWalletPriceUnavailable) {
			return nil
		}
		if err != nil {
			return err
		}
		if price.PriceToman > 0 {
			response.TokenPriceToman = &price.PriceToman
			response.PriceUpdatedAt = &price.UpdatedAt
			if account.Wallet >= 0 && account.Wallet <= math.MaxInt64/price.PriceToman {
				value := account.Wallet * price.PriceToman
				response.EstimatedValueToman = &value
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *walletService) ProfileCompletion(ctx context.Context, userID int64) (*dto.ProfileCompletionResponse, error) {
	completion, err := s.repo.ProfileCompletion(ctx, userID)
	if err != nil {
		return nil, err
	}
	steps := [...]bool{completion.FullName, completion.ProfilePhoto, completion.VerifiedEmail, completion.VerifiedPhone, completion.TelegramConnected}
	completed := 0
	for _, step := range steps {
		if step {
			completed++
		}
	}
	return &dto.ProfileCompletionResponse{
		Steps: dto.ProfileCompletionSteps{
			FullName: completion.FullName, ProfilePhoto: completion.ProfilePhoto,
			VerifiedEmail: completion.VerifiedEmail, VerifiedPhone: completion.VerifiedPhone, TelegramConnected: completion.TelegramConnected,
		},
		CompletedSteps: completed, TotalSteps: len(steps), Progress: 100 * completed / len(steps),
		RewardTokens: completion.RewardTokens, RewardGranted: completion.RewardGranted,
	}, nil
}

func (s *walletService) Offers(ctx context.Context, userID int64) (*dto.WalletOffersResponse, error) {
	if _, err := s.activeAccount(ctx, userID); err != nil {
		return nil, err
	}
	price, err := s.availablePrice(ctx)
	if err != nil {
		return nil, err
	}
	packages := make([]dto.WalletPackageResponse, 0, len(walletPackages))
	for _, pkg := range walletPackages {
		packages = append(packages, dto.WalletPackageResponse{
			ID: pkg.id, TokenAmount: pkg.tokens, AmountToman: walletAmount(price.PriceToman, pkg),
			DiscountPercent: pkg.discount, IsPopular: pkg.popular,
		})
	}
	return &dto.WalletOffersResponse{
		Currency: walletCurrency, TokenPriceToman: price.PriceToman, PriceUpdatedAt: price.UpdatedAt,
		CustomMinTokens: walletCustomMinTokens, CustomMaxTokens: walletCustomMaxTokens, Packages: packages,
	}, nil
}

func (s *walletService) Purchase(ctx context.Context, userID int64, key string, request dto.WalletPurchaseRequest) (*dto.WalletTransactionResponse, bool, error) {
	if userID <= 0 || len(key) == 0 || len(key) > 128 || strings.ContainsFunc(key, func(c rune) bool { return c < '!' || c > '~' }) {
		return nil, false, exception.ErrBadRequest
	}
	pkg, err := walletPurchasePackage(request)
	if err != nil {
		return nil, false, err
	}
	var transaction *entity.WalletTransaction
	var created bool
	err = s.write.WithTransaction(ctx, func(txCtx context.Context) error {
		account, err := s.repo.LockAccount(txCtx, userID)
		if err != nil {
			return err
		}
		if account.Status != enum.UserStatusActive {
			return exception.ErrUserInactive
		}
		transaction, err = s.repo.FindByIdempotencyKey(txCtx, userID, key)
		if err != nil {
			return err
		}
		if transaction != nil {
			if !matchesWalletPurchase(transaction, pkg, request.ExpectedAmountToman) {
				return exception.ErrWalletIdempotencyConflict
			}
			return nil
		}

		transaction, err = s.createPurchase(txCtx, account, key, pkg, request.ExpectedAmountToman)
		created = err == nil
		return err
	})
	if err != nil {
		return nil, false, err
	}
	response := walletTransactionResponse(*transaction)
	return &response, created, nil
}

func (s *walletService) createPurchase(ctx context.Context, account *entity.User, key string, pkg walletPackage, expectedAmount *int64) (*entity.WalletTransaction, error) {
	price, err := s.availablePrice(ctx)
	if err != nil {
		return nil, err
	}
	amount := walletAmount(price.PriceToman, pkg)
	if expectedAmount != nil && *expectedAmount != amount {
		return nil, exception.ErrWalletPriceChanged
	}
	status, err := s.mockPaymentStatus()
	if err != nil {
		return nil, err
	}
	transaction := &entity.WalletTransaction{
		UserID: account.ID, Type: enum.WalletTransactionPurchase, Status: status,
		Description: fmt.Sprintf("Purchase of %d tokens", pkg.tokens),
		TokenAmount: pkg.tokens, BalanceAfter: account.Wallet, AmountToman: &amount,
		BaseUnitPriceToman: &price.PriceToman, DiscountPercent: pkg.discount, IdempotencyKey: &key,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	if pkg.id != "" {
		transaction.PackageID = &pkg.id
	}
	if delta := transaction.TokenDelta(); delta > 0 {
		if account.Wallet > math.MaxInt64-delta {
			return nil, exception.ErrInternal
		}
		if err := s.users.UpdateWallet(ctx, account.ID, delta); err != nil {
			return nil, err
		}
		transaction.BalanceAfter += delta
	}
	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *walletService) mockPaymentStatus() (enum.WalletTransactionStatus, error) {
	if !s.config.MockRandomEnabled || s.config.MockSuccessPercent == 100 {
		return enum.WalletTransactionSucceeded, nil
	}
	if s.config.MockSuccessPercent == 0 {
		return enum.WalletTransactionFailed, nil
	}
	roll, err := s.random.Int64n(100)
	if err != nil {
		return "", exception.Wrap(exception.ErrInternal, err)
	}
	if roll < s.config.MockSuccessPercent {
		return enum.WalletTransactionSucceeded, nil
	}
	return enum.WalletTransactionFailed, nil
}

func (s *walletService) ListTransactions(ctx context.Context, userID int64, kind enum.WalletTransactionType, page, limit, offset int) (*dto.PaginatedResponse[dto.WalletTransactionResponse], error) {
	if kind != "" && !kind.IsValid() {
		return nil, exception.ErrBadRequest
	}
	var response *dto.PaginatedResponse[dto.WalletTransactionResponse]
	err := s.read.WithTransaction(ctx, func(txCtx context.Context) error {
		if _, err := s.activeAccount(txCtx, userID); err != nil {
			return err
		}
		transactions, total, err := s.repo.ListTransactions(txCtx, userID, kind, limit, offset)
		if err != nil {
			return err
		}
		items := make([]dto.WalletTransactionResponse, 0, len(transactions))
		for _, transaction := range transactions {
			items = append(items, walletTransactionResponse(transaction))
		}
		response = &dto.PaginatedResponse[dto.WalletTransactionResponse]{Items: items, Pagination: dto.PaginationResponse{Page: page, Limit: limit, Total: total}}
		return nil
	})
	return response, err
}

func (s *walletService) Transaction(ctx context.Context, userID, transactionID int64) (*dto.WalletTransactionResponse, error) {
	if transactionID <= 0 {
		return nil, exception.ErrBadRequest
	}
	if _, err := s.activeAccount(ctx, userID); err != nil {
		return nil, err
	}
	transaction, err := s.repo.FindTransaction(ctx, userID, transactionID)
	if err != nil {
		return nil, err
	}
	response := walletTransactionResponse(*transaction)
	return &response, nil
}

func (s *walletService) activeAccount(ctx context.Context, userID int64) (*entity.User, error) {
	if userID <= 0 {
		return nil, exception.ErrBadRequest
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Status != enum.UserStatusActive {
		return nil, exception.ErrUserInactive
	}
	return user, nil
}

func (s *walletService) availablePrice(ctx context.Context) (*entity.TokenPrice, error) {
	price, err := s.prices.CurrentPrice(ctx)
	if err != nil {
		return nil, err
	}
	if price.PriceToman <= 0 || price.PriceToman > math.MaxInt64/walletCustomMaxTokens {
		return nil, exception.ErrWalletPriceUnavailable
	}
	return price, nil
}

func walletPurchasePackage(request dto.WalletPurchaseRequest) (walletPackage, error) {
	if request.PackageID != nil && request.TokenAmount == nil {
		for _, pkg := range walletPackages {
			if pkg.id == *request.PackageID {
				return pkg, nil
			}
		}
	}
	if request.PackageID == nil && request.TokenAmount != nil && *request.TokenAmount >= walletCustomMinTokens && *request.TokenAmount <= walletCustomMaxTokens {
		return walletPackage{tokens: *request.TokenAmount}, nil
	}
	return walletPackage{}, exception.ErrWalletInvalidPurchase
}

func walletAmount(unitPrice int64, pkg walletPackage) int64 {
	subtotal := unitPrice * pkg.tokens
	percentage := 100 - pkg.discount
	return subtotal/100*percentage + (subtotal%100*percentage+99)/100
}

func matchesWalletPurchase(transaction *entity.WalletTransaction, pkg walletPackage, expectedAmount *int64) bool {
	if transaction.Type != enum.WalletTransactionPurchase || transaction.TokenAmount != pkg.tokens {
		return false
	}
	if expectedAmount != nil && (transaction.AmountToman == nil || *transaction.AmountToman != *expectedAmount) {
		return false
	}
	if transaction.PackageID == nil {
		return pkg.id == ""
	}
	return *transaction.PackageID == pkg.id
}

func walletTransactionResponse(transaction entity.WalletTransaction) dto.WalletTransactionResponse {
	response := dto.WalletTransactionResponse{
		ID: transaction.ID, Type: transaction.Type, Status: transaction.Status, Description: transaction.Description,
		TokenAmount: transaction.TokenAmount, TokenDelta: transaction.TokenDelta(), BalanceAfter: transaction.BalanceAfter,
		Currency: walletCurrency, AmountToman: transaction.AmountToman, BaseUnitPriceToman: transaction.BaseUnitPriceToman,
		DiscountPercent: transaction.DiscountPercent, PackageID: transaction.PackageID, CreatedAt: transaction.CreatedAt.UTC(),
	}
	if transaction.Type == enum.WalletTransactionPurchase {
		trackingCode := fmt.Sprintf("MOCK-%d", transaction.ID)
		paymentMode := "MOCK"
		response.TrackingCode = &trackingCode
		response.PaymentMode = &paymentMode
	}
	if transaction.Type == enum.WalletTransactionReward && transaction.IdempotencyKey != nil && *transaction.IdempotencyKey == entity.ProfileCompletionRewardKey {
		response.RewardCode = enum.WalletRewardProfileCompletion
	}
	return response
}
