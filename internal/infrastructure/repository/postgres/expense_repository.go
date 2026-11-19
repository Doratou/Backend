package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type expenseRepository struct{ db database.Database }

func NewExpenseRepository(db database.Database) domain_db.ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *expenseRepository) CreateWithShares(ctx context.Context, expense *entity.Expense) error {
	if expense == nil || expense.GroupID <= 0 || expense.CreatedByMemberID <= 0 || len(expense.Shares) == 0 {
		return exception.ErrBadRequest
	}
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		group, err := r.findGroup(tx, expense.GroupID)
		if err != nil {
			return err
		}
		if group.Status != enum.GroupStatusActive {
			return exception.ErrExpenseGroupInactive
		}
		var payer entity.GroupMember
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ? AND group_id = ? AND status = ? AND left_at IS NULL", expense.CreatedByMemberID, expense.GroupID, enum.GroupMemberStatusActive).First(&payer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrExpenseForbidden
			}
			return err
		}
		if err := tx.Omit(clause.Associations).Create(expense).Error; err != nil {
			return err
		}
		return r.replaceShares(tx, expense)
	})
	return expensePersistenceError(err)
}

func (r *expenseRepository) UpdateWithShares(ctx context.Context, expense *entity.Expense) error {
	if expense == nil || expense.ID <= 0 || expense.GroupID <= 0 || len(expense.Shares) == 0 {
		return exception.ErrBadRequest
	}
	return r.withWrite(ctx, expense.GroupID, expense.ID, func(tx *gorm.DB, stored *entity.Expense) error {
		updatedAt := time.Now().UTC()
		result := tx.Model(&entity.Expense{}).Where("id = ? AND group_id = ?", stored.ID, stored.GroupID).Updates(map[string]any{
			"title": expense.Title, "description": expense.Description, "total_amount_toman": expense.TotalAmountToman,
			"split_type": expense.SplitType, "occurred_at": expense.OccurredAt, "updated_at": updatedAt,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return exception.ErrExpenseNotFound
		}
		return r.replaceShares(tx, expense)
	})
}

func (r *expenseRepository) DeleteByID(ctx context.Context, groupID, expenseID int64) error {
	return r.withWrite(ctx, groupID, expenseID, func(tx *gorm.DB, expense *entity.Expense) error {
		if err := tx.Where("expense_id = ?", expense.ID).Delete(&entity.ExpenseShare{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND group_id = ?", expense.ID, groupID).Delete(&entity.Expense{}).Error
	})
}

func (r *expenseRepository) FindByID(ctx context.Context, groupID, expenseID int64) (*entity.Expense, error) {
	return r.findExpense(r.getDB(ctx).Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "expenses"}}).
		Preload("Shares", func(db *gorm.DB) *gorm.DB { return db.Order("member_id ASC") }).
		Preload("Attachments", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", enum.AttachmentTypeExpenseReceipt).Order("id ASC")
		}), groupID, expenseID)
}

func (r *expenseRepository) FindByGroupID(ctx context.Context, groupID, memberID int64, scope enum.ExpenseListScope, limit, offset int) ([]*entity.Expense, int64, error) {
	expenses := make([]*entity.Expense, 0)
	query := r.getDB(ctx).Model(&entity.Expense{}).
		Joins("JOIN groups ON groups.id = expenses.group_id AND groups.deleted_at IS NULL").Where("expenses.group_id = ?", groupID)
	switch scope {
	case enum.ExpenseListScopeCreatedByMe:
		query = query.Where("expenses.created_by_member_id = ?", memberID)
	case enum.ExpenseListScopeOwedByMe:
		shares := r.getDB(ctx).Model(&entity.ExpenseShare{}).Select("1").
			Where("expense_shares.expense_id = expenses.id AND expense_shares.group_id = expenses.group_id").
			Where("expense_shares.member_id = ? AND expense_shares.share_amount_toman > 0", memberID)
		query = query.Where("expenses.created_by_member_id <> ?", memberID).Where("EXISTS (?)", shares)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, expensePersistenceError(err)
	}
	err := query.Select("expenses.*").Clauses(clause.Locking{Strength: "SHARE", Table: clause.Table{Name: "expenses"}}).
		Preload("Shares", func(db *gorm.DB) *gorm.DB { return db.Order("member_id ASC") }).
		Preload("Attachments", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", enum.AttachmentTypeExpenseReceipt).Order("id ASC")
		}).
		Order("expenses.occurred_at DESC NULLS LAST").Order("expenses.id DESC").Limit(limit).Offset(offset).Find(&expenses).Error
	if err != nil {
		return nil, 0, expensePersistenceError(err)
	}
	return expenses, total, nil
}

func (r *expenseRepository) FindGroupByID(ctx context.Context, groupID int64) (*entity.Group, error) {
	return r.findGroup(r.getDB(ctx), groupID)
}

func (r *expenseRepository) FindActiveMemberByUserID(ctx context.Context, groupID, userID int64) (*entity.GroupMember, error) {
	var member entity.GroupMember
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("group_id = ? AND user_id = ? AND status = ? AND left_at IS NULL", groupID, userID, enum.GroupMemberStatusActive).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrGroupMemberNotFound
	}
	if err != nil {
		return nil, expensePersistenceError(err)
	}
	return &member, nil
}

func (r *expenseRepository) FindMembersByIDs(ctx context.Context, groupID int64, memberIDs []int64) ([]*entity.GroupMember, error) {
	members := make([]*entity.GroupMember, 0)
	if len(memberIDs) == 0 {
		return members, nil
	}
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("group_id = ? AND id IN ?", groupID, memberIDs).Order("id ASC").Find(&members).Error
	if err != nil {
		return nil, expensePersistenceError(err)
	}
	return members, nil
}

func (r *expenseRepository) FindParticipants(ctx context.Context, groupID int64, limit, offset int) ([]*entity.GroupMember, int64, error) {
	query := r.getDB(ctx).Model(&entity.GroupMember{}).Where("group_id = ?", groupID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, expensePersistenceError(err)
	}
	members := make([]*entity.GroupMember, 0)
	err := query.Preload("User", func(db *gorm.DB) *gorm.DB { return db.Unscoped().Select("id", "name", "avatar_attachment_id") }).
		Order("id ASC").Limit(limit).Offset(offset).Find(&members).Error
	if err != nil {
		return nil, 0, expensePersistenceError(err)
	}
	return members, total, nil
}

func (r *expenseRepository) FindGroupIDByExpenseID(ctx context.Context, expenseID int64) (int64, error) {
	var expense entity.Expense
	err := r.getDB(ctx).Select("id", "group_id").Where("id = ?", expenseID).First(&expense).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, exception.ErrExpenseNotFound
	}
	return expense.GroupID, expensePersistenceError(err)
}

func (r *expenseRepository) FindAttachmentsByIDs(ctx context.Context, ids []int64) ([]*entity.Attachment, error) {
	attachments := make([]*entity.Attachment, 0, len(ids))
	if len(ids) == 0 {
		return attachments, nil
	}
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id ASC").Find(&attachments).Error
	if err != nil {
		return nil, expensePersistenceError(err)
	}
	return attachments, nil
}

func (r *expenseRepository) ReplaceAttachments(ctx context.Context, groupID, expenseID int64, ids []int64) error {
	query := r.getDB(ctx).Model(&entity.Attachment{}).Where("expense_id = ? AND type = ?", expenseID, enum.AttachmentTypeExpenseReceipt)
	if len(ids) > 0 {
		query = query.Where("id NOT IN ?", ids)
	}
	if err := query.Updates(map[string]any{"expense_id": nil, "updated_at": time.Now().UTC()}).Error; err != nil {
		return expensePersistenceError(err)
	}
	if len(ids) == 0 {
		return nil
	}
	return expensePersistenceError(r.getDB(ctx).Model(&entity.Attachment{}).Where("id IN ?", ids).
		Updates(map[string]any{"expense_id": expenseID, "group_id": groupID, "is_public": false, "updated_at": time.Now().UTC()}).Error)
}

func (r *expenseRepository) replaceShares(tx *gorm.DB, expense *entity.Expense) error {
	ids := make([]int64, 0, len(expense.Shares))
	for _, share := range expense.Shares {
		if share == nil || share.MemberID <= 0 || share.ShareAmountToman < 0 || share.SplitValue < 0 {
			return exception.ErrExpenseInvalidSplit
		}
		share.ExpenseID = expense.ID
		share.GroupID = expense.GroupID
		ids = append(ids, share.MemberID)
	}
	if err := tx.Where("expense_id = ? AND member_id NOT IN ?", expense.ID, ids).Delete(&entity.ExpenseShare{}).Error; err != nil {
		return err
	}
	return tx.Omit(clause.Associations).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "expense_id"}, {Name: "member_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"share_amount_toman", "split_value", "updated_at", "deleted_at"}),
		Where:     clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "expense_shares.share_amount_toman IS DISTINCT FROM excluded.share_amount_toman OR expense_shares.split_value IS DISTINCT FROM excluded.split_value OR expense_shares.deleted_at IS NOT NULL"}}},
	}).CreateInBatches(&expense.Shares, 1000).Error
}

func (r *expenseRepository) findGroup(db *gorm.DB, groupID int64) (*entity.Group, error) {
	var group entity.Group
	err := db.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", groupID).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrGroupNotFound
	}
	if err != nil {
		return nil, expensePersistenceError(err)
	}
	return &group, nil
}

func (r *expenseRepository) findExpense(db *gorm.DB, groupID, expenseID int64) (*entity.Expense, error) {
	var expense entity.Expense
	err := db.Where("expenses.id = ? AND expenses.group_id = ?", expenseID, groupID).First(&expense).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrExpenseNotFound
	}
	if err != nil {
		return nil, expensePersistenceError(err)
	}
	return &expense, nil
}

func (r *expenseRepository) withWrite(ctx context.Context, groupID, expenseID int64, fn func(*gorm.DB, *entity.Expense) error) error {
	if groupID <= 0 || expenseID <= 0 {
		return exception.ErrBadRequest
	}
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		group, err := r.findGroup(tx, groupID)
		if err != nil {
			return err
		}
		if group.Status != enum.GroupStatusActive {
			return exception.ErrExpenseGroupInactive
		}
		expense, err := r.findExpense(tx.Clauses(clause.Locking{Strength: "UPDATE"}), groupID, expenseID)
		if err != nil {
			return err
		}
		return fn(tx, expense)
	})
	return expensePersistenceError(err)
}

func expensePersistenceError(err error) error {
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
			case "fk_expense_shares_member_group", "expenses_created_by_member_id_fkey":
				return exception.ErrExpenseParticipantNotFound
			case "fk_expense_shares_expense_group":
				return exception.ErrExpenseNotFound
			case "expenses_group_id_fkey":
				return exception.ErrGroupNotFound
			}
		case "23505":
			return exception.ErrExpenseDuplicateMembers
		case "23514", "23502", "22001", "22P02", "22003":
			return exception.ErrBadRequest
		}
	}
	return exception.Wrap(exception.ErrInternal, err)
}
