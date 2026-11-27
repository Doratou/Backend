package exception

import (
	"errors"
	"net/http"
)

var (
	ErrWalletInvalidPurchase = &AppError{
		Code: "WALLET_INVALID_PURCHASE", Message: "provide a valid package_id or a token quantity from 1 to 10000",
		HTTPStatus: http.StatusBadRequest, Err: errors.New("invalid wallet purchase selection"),
	}
	ErrWalletPriceUnavailable = &AppError{
		Code: "WALLET_PRICE_UNAVAILABLE", Message: "token pricing is temporarily unavailable",
		HTTPStatus: http.StatusServiceUnavailable, Err: errors.New("wallet price unavailable"),
	}
	ErrWalletPriceChanged = &AppError{
		Code: "WALLET_PRICE_CHANGED", Message: "purchase price changed; refresh wallet offers and confirm the new amount",
		HTTPStatus: http.StatusConflict, Err: errors.New("wallet price changed"),
	}
	ErrWalletIdempotencyConflict = &AppError{
		Code: "WALLET_IDEMPOTENCY_CONFLICT", Message: "this idempotency key was already used for a different purchase",
		HTTPStatus: http.StatusConflict, Err: errors.New("wallet idempotency key conflict"),
	}
	ErrWalletTransactionNotFound = &AppError{
		Code: "WALLET_TRANSACTION_NOT_FOUND", Message: "wallet transaction not found",
		HTTPStatus: http.StatusNotFound, Err: errors.New("wallet transaction not found"),
	}
)
