package postgres

import (
	"context"
	"errors"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type oauthRepository struct {
	db database.Database
}

// NewOAuthRepository creates a new PostgreSQL implementation of domain_repo.OAuthRepository.
func NewOAuthRepository(db database.Database) domain_repo.OAuthRepository {
	return &oauthRepository{db: db}
}

func (r *oauthRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *oauthRepository) FindByProvider(ctx context.Context, provider, providerID string) (*entity.UserOAuthAccount, error) {
	var account entity.UserOAuthAccount
	db := r.getDB(ctx)

	err := db.Preload("User").Where("provider = ? AND provider_id = ?", provider, providerID).First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &account, nil
}

func (r *oauthRepository) Create(ctx context.Context, account *entity.UserOAuthAccount) error {
	db := r.getDB(ctx)
	if err := db.Create(account).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *oauthRepository) ExistsByUserAndProvider(ctx context.Context, userID int64, provider string) (bool, error) {
	var count int64
	db := r.getDB(ctx)

	err := db.Model(&entity.UserOAuthAccount{}).Where("user_id = ? AND provider = ?", userID, provider).Count(&count).Error
	if err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}
