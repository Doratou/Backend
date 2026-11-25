package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

// TokenPriceService defines business operations for managing platform token pricing.
type TokenPriceService interface {
	UpdatePrice(ctx context.Context, adminUserID int64, req dto.UpdateTokenPriceRequest) (*dto.TokenPriceResponse, error)
}
