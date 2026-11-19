package postgres

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tripSuggestionRepository struct{ db database.Database }

func NewTripSuggestionRepository(db database.Database) domain_repo.TripSuggestionRepository {
	return &tripSuggestionRepository{db: db}
}

func (r *tripSuggestionRepository) Create(ctx context.Context, suggestion *entity.TripSuggestion) error {
	db := database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
	err := db.Transaction(func(tx *gorm.DB) error {
		items := suggestion.Items
		suggestion.Items = nil
		if err := tx.Omit(clause.Associations).Create(suggestion).Error; err != nil {
			suggestion.Items = items
			return err
		}
		for i := range items {
			items[i].SuggestionID = suggestion.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				suggestion.Items = items
				return err
			}
		}
		suggestion.Items = items
		return nil
	})
	if err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *tripSuggestionRepository) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]entity.TripSuggestion, int64, error) {
	db := database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
	var total int64
	if err := db.Model(&entity.TripSuggestion{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	suggestions := make([]entity.TripSuggestion, 0)
	if err := db.Preload("Items").Where("group_id = ?", groupID).
		Order("created_at DESC").Order("id DESC").Limit(limit).Offset(offset).Find(&suggestions).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return suggestions, total, nil
}
