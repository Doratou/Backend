package postgres

import (
	"context"
	"errors"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainrepo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type checklistRepository struct{ db database.Database }

func NewChecklistRepository(db database.Database) domainrepo.ChecklistRepository {
	return &checklistRepository{db: db}
}

func (r *checklistRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *checklistRepository) FindGroup(ctx context.Context, groupID int64, lock bool) (*entity.Group, error) {
	db := r.getDB(ctx)
	if lock {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	var group entity.Group
	if err := db.First(&group, groupID).Error; err != nil {
		return nil, checklistLookupError(err, exception.ErrGroupNotFound)
	}
	return &group, nil
}

func (r *checklistRepository) FindActiveMember(ctx context.Context, groupID, userID int64, lock bool) (*entity.GroupMember, error) {
	db := r.activeMembers(ctx, groupID)
	if lock {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	var member entity.GroupMember
	if err := db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		return nil, checklistLookupError(err, exception.ErrChecklistForbidden)
	}
	return &member, nil
}

func (r *checklistRepository) ListActiveMembers(ctx context.Context, groupID int64) ([]entity.GroupMember, error) {
	var members []entity.GroupMember
	err := r.activeMembers(ctx, groupID).Order("id ASC").Find(&members).Error
	return members, checklistPersistenceError(err)
}

func (r *checklistRepository) activeMembers(ctx context.Context, groupID int64) *gorm.DB {
	return r.getDB(ctx).Preload("User", checklistProfileColumns).
		Where("group_id = ? AND status = ? AND left_at IS NULL", groupID, enum.GroupMemberStatusActive)
}

func (r *checklistRepository) ListItems(ctx context.Context, groupID, memberID int64, filter enum.ChecklistFilter, limit, offset int) ([]entity.ChecklistItem, error) {
	db := r.items(ctx, groupID)
	switch filter {
	case enum.ChecklistFilterClaimed:
		db = db.Where("assigned_to_member_id IS NOT NULL")
	case enum.ChecklistFilterUnclaimed:
		db = db.Where("type = ? AND assigned_to_member_id IS NULL", enum.ChecklistItemTypeGroup)
	case enum.ChecklistFilterMine:
		db = db.Where("assigned_to_member_id = ?", memberID)
	case enum.ChecklistFilterPersonal:
		db = db.Where("type = ?", enum.ChecklistItemTypePersonal)
	}
	var items []entity.ChecklistItem
	err := db.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error
	return items, checklistPersistenceError(err)
}

func (r *checklistRepository) CountItems(ctx context.Context, groupID int64) ([]domainrepo.ChecklistItemCount, error) {
	var counts []domainrepo.ChecklistItemCount
	err := r.getDB(ctx).Model(&entity.ChecklistItem{}).Where("group_id = ?", groupID).
		Select("type, assigned_to_member_id, COUNT(*) AS item_count").
		Group("type, assigned_to_member_id").Scan(&counts).Error
	return counts, checklistPersistenceError(err)
}

func (r *checklistRepository) FindItemForUpdate(ctx context.Context, groupID, itemID int64) (*entity.ChecklistItem, error) {
	var item entity.ChecklistItem
	if err := r.items(ctx, groupID).Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, itemID).Error; err != nil {
		return nil, checklistLookupError(err, exception.ErrChecklistItemNotFound)
	}
	return &item, nil
}

func (r *checklistRepository) items(ctx context.Context, groupID int64) *gorm.DB {
	return r.getDB(ctx).Where("group_id = ?", groupID).
		Preload("CreatedByMember.User", checklistProfileColumns).
		Preload("AssignedToMember.User", checklistProfileColumns)
}

func (r *checklistRepository) CreateItem(ctx context.Context, item *entity.ChecklistItem) error {
	return checklistPersistenceError(r.getDB(ctx).Omit(clause.Associations).Create(item).Error)
}

func (r *checklistRepository) SetAssignee(ctx context.Context, item *entity.ChecklistItem, memberID *int64) error {
	var assignee any
	if memberID != nil {
		assignee = *memberID
	}
	err := r.getDB(ctx).Raw(`
		UPDATE checklist_items
		SET assigned_to_member_id = ?, updated_at = clock_timestamp()
		WHERE group_id = ? AND id = ? AND deleted_at IS NULL
		RETURNING updated_at
	`, assignee, item.GroupID, item.ID).Row().Scan(&item.UpdatedAt)
	return checklistPersistenceError(err)
}

func (r *checklistRepository) DeleteItem(ctx context.Context, item *entity.ChecklistItem) error {
	return checklistPersistenceError(r.getDB(ctx).Delete(item).Error)
}

func checklistProfileColumns(db *gorm.DB) *gorm.DB {
	return db.Select("id", "name", "status", "avatar_attachment_id")
}

func checklistLookupError(err error, notFound *exception.AppError) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFound
	}
	return checklistPersistenceError(err)
}

func checklistPersistenceError(err error) error {
	if err == nil {
		return nil
	}
	return exception.Wrap(exception.ErrInternal, err)
}
