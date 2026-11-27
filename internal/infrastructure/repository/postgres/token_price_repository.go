package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tokenPriceRepository struct {
	db database.Database
}

// NewTokenPriceRepository creates a PostgreSQL-backed TokenPriceRepository.
func NewTokenPriceRepository(db database.Database) domain_repo.TokenPriceRepository {
	return &tokenPriceRepository{db: db}
}

func (r *tokenPriceRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *tokenPriceRepository) CurrentPrice(ctx context.Context) (*entity.TokenPrice, error) {
	var price entity.TokenPrice
	err := r.getDB(ctx).First(&price, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrWalletPriceUnavailable
	}
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &price, nil
}

func (r *tokenPriceRepository) UpdatePrice(ctx context.Context, newPriceToman int64) (int64, *entity.TokenPrice, error) {
	var oldPrice int64
	var updated entity.TokenPrice

	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.TokenPrice
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", 1).First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			seed := entity.TokenPrice{ID: 1, PriceToman: 0}
			if cErr := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&seed).Error; cErr != nil {
				return cErr
			}
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", 1).First(&current).Error
		}
		if err != nil {
			return err
		}

		oldPrice = current.PriceToman
		now := time.Now().UTC()
		if err := tx.Model(&current).Updates(map[string]any{
			"price_toman": newPriceToman,
			"updated_at":  now,
		}).Error; err != nil {
			return err
		}

		updated = entity.TokenPrice{
			ID:         current.ID,
			PriceToman: newPriceToman,
			UpdatedAt:  now,
		}
		return nil
	})
	if err != nil {
		return 0, nil, exception.Wrap(exception.ErrInternal, err)
	}

	return oldPrice, &updated, nil
}
