package entity

import "time"

// TokenPrice represents the singleton persisted platform token price in Toman.
type TokenPrice struct {
	ID         int16     `gorm:"primaryKey;default:1" json:"id"`
	PriceToman int64     `gorm:"column:price_toman;not null;default:0" json:"price_toman"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TokenPrice) TableName() string {
	return "token_prices"
}
