package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
)

// TokenPriceRepository manages persistence for the platform token price.
type TokenPriceRepository interface {
	CurrentPrice(ctx context.Context) (*entity.TokenPrice, error)
	// UpdatePrice atomically updates the token price in Toman and returns the previous price alongside the updated record.
	UpdatePrice(ctx context.Context, newPriceToman int64) (oldPriceToman int64, updated *entity.TokenPrice, err error)
}
