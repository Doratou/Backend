package dto

import "time"

// UpdateTokenPriceRequest holds the validated input to update the platform token price.
type UpdateTokenPriceRequest struct {
	PriceToman int64 `json:"price_toman" binding:"required,gt=0"`
}

// TokenPriceResponse represents the updated token price state.
type TokenPriceResponse struct {
	OldPriceToman int64     `json:"old_price_toman"`
	NewPriceToman int64     `json:"new_price_toman"`
	UpdatedAt     time.Time `json:"updated_at"`
}
