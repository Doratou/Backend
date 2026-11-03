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
)

type groupRepository struct{ db database.Database }

func NewGroupRepository(db database.Database) domainrepo.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *groupRepository) CreateWithOwner(ctx context.Context, group *entity.Group) error {
	db := r.getDB(ctx)
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(group).Error; err != nil {
			return err
		}
		member := entity.GroupMember{GroupID: group.ID, UserID: group.CreatedByUserID, Status: enum.GroupMemberStatusActive}
		return tx.Create(&member).Error
	})
	if err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *groupRepository) FindByID(ctx context.Context, id int64) (*entity.Group, int64, error) {
	var group entity.Group
	err := r.getDB(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, exception.ErrGroupNotFound
	}
	if err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	var count int64
	if err := r.getDB(ctx).Model(&entity.GroupMember{}).Where("group_id = ? AND status = ?", id, enum.GroupMemberStatusActive).Count(&count).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return &group, count, nil
}

func (r *groupRepository) FindByMember(ctx context.Context, userID int64, status *enum.GroupStatus, limit, offset int) ([]entity.Group, int64, error) {
	var groups []entity.Group
	query := r.getDB(ctx).Model(&entity.Group{}).Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ? AND group_members.status = ? AND groups.deleted_at IS NULL", userID, enum.GroupMemberStatusActive)
	if status != nil {
		query = query.Where("groups.status = ?", *status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	if err := query.Order("groups.created_at DESC").Limit(limit).Offset(offset).Find(&groups).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return groups, total, nil
}

func (r *groupRepository) FindAll(ctx context.Context, status *enum.GroupStatus, limit, offset int) ([]entity.Group, int64, error) {
	var groups []entity.Group
	query := r.getDB(ctx).Model(&entity.Group{}).Where("groups.deleted_at IS NULL")
	if status != nil {
		query = query.Where("groups.status = ?", *status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	if err := query.Order("groups.created_at DESC").Limit(limit).Offset(offset).Find(&groups).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}
	return groups, total, nil
}

func (r *groupRepository) Update(ctx context.Context, group *entity.Group) error {
	if err := r.getDB(ctx).Model(&entity.Group{}).Where("id = ? AND deleted_at IS NULL", group.ID).Updates(map[string]interface{}{"title": group.Title, "description": group.Description}).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *groupRepository) SetStatus(ctx context.Context, groupID int64, status enum.GroupStatus) error {
	updates := map[string]interface{}{"status": status}
	if status == enum.GroupStatusInactive {
		updates["archived_at"] = gorm.Expr("CURRENT_TIMESTAMP")
	} else {
		updates["archived_at"] = nil
	}
	if err := r.getDB(ctx).Model(&entity.Group{}).Where("id = ? AND deleted_at IS NULL", groupID).Updates(updates).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *groupRepository) Delete(ctx context.Context, groupID int64) error {
	if err := r.getDB(ctx).Delete(&entity.Group{}, groupID).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *groupRepository) CloneWithMembers(ctx context.Context, sourceGroupID int64, newGroup *entity.Group) error {
	if err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newGroup).Error; err != nil {
			return err
		}
		var members []entity.GroupMember
		if err := tx.Where("group_id = ? AND status = ?", sourceGroupID, enum.GroupMemberStatusActive).Find(&members).Error; err != nil {
			return err
		}
		for _, member := range members {
			if member.UserID == newGroup.CreatedByUserID {
				continue
			}
			copy := entity.GroupMember{GroupID: newGroup.ID, UserID: member.UserID, Status: enum.GroupMemberStatusActive}
			if err := tx.Create(&copy).Error; err != nil {
				return err
			}
		}
		owner := entity.GroupMember{GroupID: newGroup.ID, UserID: newGroup.CreatedByUserID, Status: enum.GroupMemberStatusActive}
		return tx.Create(&owner).Error
	}); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *groupRepository) IsActiveMember(ctx context.Context, groupID, userID int64) (bool, error) {
	var count int64
	err := r.getDB(ctx).Model(&entity.GroupMember{}).Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, enum.GroupMemberStatusActive).Count(&count).Error
	if err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}
