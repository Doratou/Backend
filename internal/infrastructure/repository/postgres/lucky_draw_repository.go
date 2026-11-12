package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type luckyDrawRepository struct{ db database.Database }

func NewLuckyDrawRepository(db database.Database) domain_repo.LuckyDrawRepository {
	return &luckyDrawRepository{db: db}
}

func (r *luckyDrawRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *luckyDrawRepository) Create(ctx context.Context, draw *entity.LuckyDraw) error {
	return luckyDrawPersistenceError(r.getDB(ctx).Omit(clause.Associations).Create(draw).Error)
}

func (r *luckyDrawRepository) FindByID(ctx context.Context, groupID, drawID int64) (*entity.LuckyDraw, error) {
	return r.findDraw(r.getDB(ctx), groupID, drawID)
}

func (r *luckyDrawRepository) FindByIDForUpdate(ctx context.Context, groupID, drawID int64) (*entity.LuckyDraw, error) {
	return r.findDraw(r.getDB(ctx).Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "lucky_draws"}}), groupID, drawID)
}

func (r *luckyDrawRepository) findDraw(db *gorm.DB, groupID, drawID int64) (*entity.LuckyDraw, error) {
	var draw entity.LuckyDraw
	err := db.Select("lucky_draws.*").
		Joins("JOIN groups ON groups.id = lucky_draws.group_id AND groups.deleted_at IS NULL").
		Where("lucky_draws.id = ? AND lucky_draws.group_id = ?", drawID, groupID).First(&draw).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrLuckyDrawNotFound
	}
	if err != nil {
		return nil, luckyDrawPersistenceError(err)
	}
	return &draw, nil
}

func (r *luckyDrawRepository) FindByGroupID(ctx context.Context, groupID int64, status *enum.LuckyDrawStatus, limit, offset int) ([]*entity.LuckyDraw, int64, error) {
	draws := make([]*entity.LuckyDraw, 0)
	var total int64
	db := r.getDB(ctx).Model(&entity.LuckyDraw{}).
		Joins("JOIN groups ON groups.id = lucky_draws.group_id AND groups.deleted_at IS NULL").
		Where("lucky_draws.group_id = ?", groupID)
	if status != nil {
		db = db.Where("lucky_draws.status = ?", *status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, luckyDrawPersistenceError(err)
	}
	err := db.Select("lucky_draws.*").Order("lucky_draws.created_at DESC").Order("lucky_draws.id DESC").
		Limit(limit).Offset(offset).Find(&draws).Error
	if err != nil {
		return nil, 0, luckyDrawPersistenceError(err)
	}
	return draws, total, nil
}

func (r *luckyDrawRepository) FindGroupByID(ctx context.Context, groupID int64) (*entity.Group, error) {
	var group entity.Group
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", groupID).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrGroupNotFound
	}
	if err != nil {
		return nil, luckyDrawPersistenceError(err)
	}
	return &group, nil
}

func (r *luckyDrawRepository) FindActiveMemberByUserID(ctx context.Context, groupID, userID int64) (*entity.GroupMember, error) {
	db := r.getDB(ctx)
	var user entity.User
	err := db.Clauses(clause.Locking{Strength: "SHARE"}).Select("id", "status").Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrUserNotFound
	}
	if err != nil {
		return nil, luckyDrawPersistenceError(err)
	}
	if user.Status != enum.UserStatusActive {
		return nil, exception.ErrUserInactive
	}
	var member entity.GroupMember
	err = db.Clauses(clause.Locking{Strength: "SHARE"}).
		Where("group_id = ? AND user_id = ? AND status = ? AND left_at IS NULL", groupID, userID, enum.GroupMemberStatusActive).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrLuckyDrawForbidden
	}
	if err != nil {
		return nil, luckyDrawPersistenceError(err)
	}
	return &member, nil
}

func (r *luckyDrawRepository) UpdateDraft(ctx context.Context, draw *entity.LuckyDraw) error {
	data, err := json.Marshal(draw.Candidates)
	if err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	db := r.getDB(ctx)
	updatedAt := db.NowFunc().UTC().Truncate(time.Microsecond)
	result := db.Model(&entity.LuckyDraw{}).
		Where("id = ? AND group_id = ? AND created_by_member_id = ? AND status = ?", draw.ID, draw.GroupID, draw.CreatedByMemberID, enum.LuckyDrawStatusDraft).
		Updates(map[string]any{"title": draw.Title, "candidates": string(data), "updated_at": updatedAt})
	if result.Error != nil {
		return luckyDrawPersistenceError(result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrLuckyDrawCompleted
	}
	draw.UpdatedAt = updatedAt
	return nil
}

func (r *luckyDrawRepository) Complete(ctx context.Context, draw *entity.LuckyDraw) error {
	db := r.getDB(ctx)
	completedAt := db.NowFunc().UTC().Truncate(time.Microsecond)
	result := db.Model(&entity.LuckyDraw{}).
		Where("id = ? AND group_id = ? AND created_by_member_id = ? AND status = ?", draw.ID, draw.GroupID, draw.CreatedByMemberID, enum.LuckyDrawStatusDraft).
		Updates(map[string]any{"status": enum.LuckyDrawStatusCompleted, "result_index": draw.ResultIndex, "completed_at": completedAt, "updated_at": completedAt})
	if result.Error != nil {
		return luckyDrawPersistenceError(result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrLuckyDrawCompleted
	}
	draw.Status = enum.LuckyDrawStatusCompleted
	draw.CompletedAt = &completedAt
	draw.UpdatedAt = completedAt
	return nil
}

func luckyDrawPersistenceError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *exception.AppError
	if errors.As(err, &appErr) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			switch pgErr.ConstraintName {
			case "lucky_draws_group_id_fkey":
				return exception.ErrGroupNotFound
			case "lucky_draws_created_by_member_id_fkey":
				return exception.ErrGroupMemberNotFound
			}
		case "23514":
			if pgErr.ConstraintName == "chk_lucky_draws_candidates" {
				return exception.ErrLuckyDrawInvalidCandidates
			}
			if pgErr.ConstraintName == "chk_lucky_draws_title" {
				return exception.ErrBadRequest
			}
		case "22001", "22P02", "23502":
			return exception.ErrBadRequest
		}
	}
	return exception.Wrap(exception.ErrInternal, err)
}
