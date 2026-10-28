package postgres

import (
	"context"
	"errors"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type friendshipRepository struct {
	db database.Database
}

func NewFriendshipRepository(db database.Database) domain_repo.FriendshipRepository {
	return &friendshipRepository{db: db}
}

func (r *friendshipRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *friendshipRepository) Create(ctx context.Context, friendship *entity.Friendship) error {
	db := r.getDB(ctx)
	if err := db.Omit(clause.Associations).Create(friendship).Error; err != nil {
		return friendshipPersistenceError(err)
	}
	return nil
}

func (r *friendshipRepository) FindByID(ctx context.Context, id int64) (*entity.Friendship, error) {
	var friendship entity.Friendship
	db := r.getDB(ctx)

	err := db.Preload("RequesterUser.AvatarAttachment").Preload("ReceiverUser.AvatarAttachment").
		Where("id = ?", id).First(&friendship).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrFriendshipNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &friendship, nil
}

func (r *friendshipRepository) FindByUserPair(ctx context.Context, firstUserID, secondUserID int64) (*entity.Friendship, error) {
	var friendship entity.Friendship
	db := r.getDB(ctx)
	if firstUserID > secondUserID {
		firstUserID, secondUserID = secondUserID, firstUserID
	}

	err := db.Preload("RequesterUser.AvatarAttachment").Preload("ReceiverUser.AvatarAttachment").
		Where("LEAST(requester_user_id, receiver_user_id) = ? AND GREATEST(requester_user_id, receiver_user_id) = ?", firstUserID, secondUserID).
		First(&friendship).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrFriendshipNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &friendship, nil
}

func (r *friendshipRepository) FindIncomingRequests(ctx context.Context, userID int64, limit, offset int) ([]*entity.Friendship, int64, error) {
	db := r.getDB(ctx).Where("receiver_user_id = ? AND status = ?", userID, enum.FriendshipStatusPending)
	return r.findAll(db, limit, offset)
}

func (r *friendshipRepository) FindOutgoingRequests(ctx context.Context, userID int64, limit, offset int) ([]*entity.Friendship, int64, error) {
	db := r.getDB(ctx).Where("requester_user_id = ? AND status = ?", userID, enum.FriendshipStatusPending)
	return r.findAll(db, limit, offset)
}

func (r *friendshipRepository) FindFriends(ctx context.Context, userID int64, limit, offset int) ([]*entity.Friendship, int64, error) {
	db := r.getDB(ctx).Where("(requester_user_id = ? OR receiver_user_id = ?) AND status = ?", userID, userID, enum.FriendshipStatusAccepted)
	return r.findAll(db, limit, offset)
}

func (r *friendshipRepository) Update(ctx context.Context, friendship *entity.Friendship, expectedStatus enum.FriendshipStatus) error {
	db := r.getDB(ctx)
	updatedAt := db.NowFunc()
	res := db.Model(&entity.Friendship{}).Where("id = ? AND status = ?", friendship.ID, expectedStatus).
		Updates(map[string]any{
			"requester_user_id": friendship.RequesterUserID,
			"receiver_user_id":  friendship.ReceiverUserID,
			"status":            friendship.Status,
			"responded_at":      friendship.RespondedAt,
			"updated_at":        updatedAt,
		})
	if res.Error != nil {
		return friendshipPersistenceError(res.Error)
	}
	if res.RowsAffected == 0 {
		return r.writeConflict(ctx, friendship.ID)
	}
	friendship.UpdatedAt = updatedAt
	return nil
}

func (r *friendshipRepository) DeleteByID(ctx context.Context, id int64, expectedStatus enum.FriendshipStatus) error {
	db := r.getDB(ctx)
	res := db.Where("id = ? AND status = ?", id, expectedStatus).Delete(&entity.Friendship{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return r.writeConflict(ctx, id)
	}
	return nil
}

func (r *friendshipRepository) findAll(db *gorm.DB, limit, offset int) ([]*entity.Friendship, int64, error) {
	friendships := make([]*entity.Friendship, 0)
	var total int64

	if err := db.Model(&entity.Friendship{}).Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	query := db.Preload("RequesterUser.AvatarAttachment").Preload("ReceiverUser.AvatarAttachment").Order("id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&friendships).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return friendships, total, nil
}

func (r *friendshipRepository) writeConflict(ctx context.Context, id int64) error {
	var count int64
	db := r.getDB(ctx)
	if err := db.Model(&entity.Friendship{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	if count == 0 {
		return exception.ErrFriendshipNotFound
	}
	return exception.ErrFriendshipInvalidStatus
}

func friendshipPersistenceError(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return exception.ErrFriendshipAlreadyExists
	}
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return exception.ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return exception.ErrFriendshipAlreadyExists
		case "23503":
			return exception.ErrUserNotFound
		case "23514":
			if pgErr.ConstraintName == "chk_friendships_distinct_users" {
				return exception.ErrSelfFriendRequest
			}
		}
	}
	return exception.Wrap(exception.ErrInternal, err)
}
