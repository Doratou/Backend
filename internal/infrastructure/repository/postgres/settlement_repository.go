package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/finance"
	domainDB "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type settlementRepository struct{ db database.Database }

func NewSettlementRepository(db database.Database) domainDB.SettlementRepository {
	return &settlementRepository{db: db}
}

func (r *settlementRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *settlementRepository) FindGroupByID(ctx context.Context, id int64, exclusive bool) (*entity.Group, error) {
	var group entity.Group
	lock := "SHARE"
	if exclusive {
		lock = "NO KEY UPDATE"
	}
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: lock}).First(&group, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrGroupNotFound
	}
	return &group, settlementPersistenceError(err)
}

func (r *settlementRepository) FindMemberByUserID(ctx context.Context, groupID, userID int64) (*entity.GroupMember, error) {
	var member entity.GroupMember
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("group_id = ? AND user_id = ? AND status = ? AND left_at IS NULL", groupID, userID, enum.GroupMemberStatusActive).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrSettlementForbidden
	}
	return &member, settlementPersistenceError(err)
}

func (r *settlementRepository) FindMembersByIDs(ctx context.Context, groupID int64, ids []int64) ([]*entity.GroupMember, error) {
	members := make([]*entity.GroupMember, 0, len(ids))
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("group_id = ? AND id IN ?", groupID, ids).Order("id ASC").Find(&members).Error
	return members, settlementPersistenceError(err)
}

func (r *settlementRepository) Create(ctx context.Context, payment *entity.Settlement) error {
	return settlementPersistenceError(r.getDB(ctx).Omit(clause.Associations).Create(payment).Error)
}

func (r *settlementRepository) preloads(db *gorm.DB) *gorm.DB {
	return db.Preload("FromMember").Preload("ToMember").
		Preload("FromMember.User", func(db *gorm.DB) *gorm.DB { return db.Unscoped().Select("id", "name", "avatar_attachment_id") }).
		Preload("ToMember.User", func(db *gorm.DB) *gorm.DB { return db.Unscoped().Select("id", "name", "avatar_attachment_id") }).
		Preload("Attachments", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", enum.AttachmentTypeSettlementReceipt).Order("id ASC")
		})
}

func (r *settlementRepository) FindByID(ctx context.Context, groupID, id int64, lock bool) (*entity.Settlement, error) {
	db := r.getDB(ctx)
	if lock {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var payment entity.Settlement
	err := r.preloads(db).Where("group_id = ? AND id = ?", groupID, id).First(&payment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrSettlementNotFound
	}
	return &payment, settlementPersistenceError(err)
}

func (r *settlementRepository) FindGroupID(ctx context.Context, id int64) (int64, error) {
	var payment entity.Settlement
	err := r.getDB(ctx).Select("id", "group_id").First(&payment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, exception.ErrSettlementNotFound
	}
	return payment.GroupID, settlementPersistenceError(err)
}

func (r *settlementRepository) List(ctx context.Context, groupID, memberID int64, filter domainDB.SettlementFilter, limit, offset int) ([]*entity.Settlement, int64, error) {
	db := r.getDB(ctx).Model(&entity.Settlement{}).Where("group_id = ?", groupID).
		Where("(status IN ? OR from_member_id = ? OR to_member_id = ?)", []enum.SettlementStatus{enum.SettlementStatusSuccess, enum.SettlementStatusReject}, memberID, memberID)
	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}
	if filter.Direction == enum.SettlementDirectionIncoming {
		db = db.Where("to_member_id = ?", memberID)
	}
	if filter.Direction == enum.SettlementDirectionOutgoing {
		db = db.Where("from_member_id = ?", memberID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, settlementPersistenceError(err)
	}
	items := make([]*entity.Settlement, 0)
	err := r.preloads(db).Order("created_at DESC").Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error
	return items, total, settlementPersistenceError(err)
}

func (r *settlementRepository) Respond(ctx context.Context, groupID, id int64, status enum.SettlementStatus, legs []finance.Transfer) error {
	now := time.Now().UTC()
	fields := map[string]any{"status": status, "updated_at": now}
	if status == enum.SettlementStatusSuccess {
		if len(legs) == 0 {
			return exception.ErrInternal
		}
		encoded, err := json.Marshal(legs)
		if err != nil {
			return settlementPersistenceError(err)
		}
		fields["balance_legs"] = string(encoded)
		fields["confirmed_at"] = now
	} else if status == enum.SettlementStatusReject {
		fields["rejected_at"] = now
	} else {
		return exception.ErrBadRequest
	}
	res := r.getDB(ctx).Model(&entity.Settlement{}).Where("id = ? AND group_id = ? AND status = ?", id, groupID, enum.SettlementStatusPending).Updates(fields)
	if res.Error != nil {
		return settlementPersistenceError(res.Error)
	}
	if res.RowsAffected != 1 {
		return exception.ErrSettlementNotPending
	}
	return nil
}

func (r *settlementRepository) DeletePending(ctx context.Context, groupID, id int64) error {
	res := r.getDB(ctx).Where("id = ? AND group_id = ? AND status = ?", id, groupID, enum.SettlementStatusPending).Delete(&entity.Settlement{})
	if res.Error != nil {
		return settlementPersistenceError(res.Error)
	}
	if res.RowsAffected != 1 {
		return exception.ErrSettlementNotPending
	}
	return settlementPersistenceError(r.getDB(ctx).Model(&entity.Attachment{}).Where("settlement_id = ?", id).Updates(map[string]any{"settlement_id": nil, "updated_at": time.Now().UTC()}).Error)
}

func (r *settlementRepository) FindAttachmentsByIDs(ctx context.Context, ids []int64) ([]*entity.Attachment, error) {
	items := make([]*entity.Attachment, 0, len(ids))
	if len(ids) == 0 {
		return items, nil
	}
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id ASC").Find(&items).Error
	return items, settlementPersistenceError(err)
}

func (r *settlementRepository) BindAttachments(ctx context.Context, groupID, id int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return settlementPersistenceError(r.getDB(ctx).Model(&entity.Attachment{}).Where("id IN ?", ids).Updates(map[string]any{"settlement_id": id, "group_id": groupID, "is_public": false, "updated_at": time.Now().UTC()}).Error)
}

func settlementPersistenceError(err error) error {
	if err == nil {
		return nil
	}
	var app *exception.AppError
	if errors.As(err, &app) {
		return err
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.ConstraintName {
		case "settlement_history_immutable":
			return exception.ErrSettlementNotPending
		case "group_status_transition":
			return exception.ErrGroupStatusTransition
		case "group_pending_settlements":
			return exception.ErrGroupPendingSettlements
		case "group_has_settlements":
			return exception.ErrGroupHasSettlements
		case "group_closed":
			return exception.ErrGroupClosed
		case "fk_settlements_from_group", "fk_settlements_to_group":
			return exception.ErrSettlementRecipientNotFound
		}
		if pg.Code == "23514" || pg.Code == "23502" || pg.Code == "22003" {
			return exception.ErrBadRequest
		}
	}
	return exception.Wrap(exception.ErrInternal, err)
}
