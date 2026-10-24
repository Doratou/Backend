package postgres

import (
	"context"
	"errors"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type userRepository struct {
	db database.Database
}

func NewUserRepository(db database.Database) domain_repo.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	db := r.getDB(ctx)
	if err := db.Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return exception.ErrUserAlreadyExists
		}
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	db := r.getDB(ctx)

	err := db.Preload("AvatarAttachment").Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	db := r.getDB(ctx)

	err := db.Preload("AvatarAttachment").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &user, nil
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	db := r.getDB(ctx)

	err := db.Preload("AvatarAttachment").Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &user, nil
}

func (r *userRepository) FindByTelegramChatID(ctx context.Context, chatID int64) (*entity.User, error) {
	var user entity.User
	db := r.getDB(ctx)

	err := db.Preload("AvatarAttachment").Where("telegram_chat_id = ?", chatID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &user, nil
}

func (r *userRepository) FindByTelegramUserID(ctx context.Context, telegramUserID int64) (*entity.User, error) {
	var user entity.User
	db := r.getDB(ctx)

	err := db.Preload("AvatarAttachment").Where("telegram_user_id = ?", telegramUserID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64
	db := r.getDB(ctx)

	if err := db.Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	query := db.Preload("AvatarAttachment").Order("id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	return users, total, nil
}

func (r *userRepository) FindByStatus(ctx context.Context, status enum.UserStatus, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	db := r.getDB(ctx).Where("status = ?", status).Order("id DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	if err := db.Find(&users).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return users, nil
}

func (r *userRepository) FindByRole(ctx context.Context, role enum.UserRole, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	db := r.getDB(ctx).Where("system_role = ?", role).Order("id DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	if err := db.Find(&users).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	db := r.getDB(ctx)
	if err := db.Save(user).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	db := r.getDB(ctx)
	res := db.Model(&entity.User{}).Where("id = ?", userID).Update("password_hash", passwordHash)
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) UpdateAvatar(ctx context.Context, userID int64, avatarID int64) error {
	db := r.getDB(ctx)
	var avatarVal any = avatarID
	if avatarID <= 0 {
		avatarVal = nil
	}
	res := db.Model(&entity.User{}).Where("id = ?", userID).Update("avatar_attachment_id", avatarVal)
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) UpdateVerifiedStatus(ctx context.Context, userID int64, isEmailVerified, isPhoneVerified bool) error {
	db := r.getDB(ctx)
	res := db.Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]any{
		"is_email_verified": isEmailVerified,
		"is_phone_verified": isPhoneVerified,
	})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) UpdateWallet(ctx context.Context, userID int64, delta int64) error {
	db := r.getDB(ctx)
	res := db.Model(&entity.User{}).Where("id = ?", userID).
		Update("wallet", gorm.Expr("wallet + ?", delta))
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) DeleteByID(ctx context.Context, id int64) error {
	db := r.getDB(ctx)
	res := db.Where("id = ?", id).Delete(&entity.User{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) RestoreByID(ctx context.Context, id int64) error {
	db := r.getDB(ctx)
	res := db.Unscoped().Model(&entity.User{}).Where("id = ?", id).Update("deleted_at", nil)
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) HardDeleteByID(ctx context.Context, id int64) error {
	db := r.getDB(ctx)
	res := db.Unscoped().Where("id = ?", id).Delete(&entity.User{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) ExistsByID(ctx context.Context, id int64) (bool, error) {
	var count int64
	db := r.getDB(ctx)
	if err := db.Model(&entity.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	db := r.getDB(ctx)
	if err := db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}

func (r *userRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	db := r.getDB(ctx)
	if err := db.Model(&entity.User{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	db := r.getDB(ctx)
	if err := db.Model(&entity.User{}).Count(&count).Error; err != nil {
		return 0, exception.Wrap(exception.ErrInternal, err)
	}
	return count, nil
}
