package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainrepo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *groupRepository) ListActiveMembers(ctx context.Context, groupID int64) ([]entity.GroupMember, error) {
	var members []entity.GroupMember
	if err := r.getDB(ctx).Preload("User").Where("group_id = ? AND status = ?", groupID, enum.GroupMemberStatusActive).
		Order("joined_at ASC").Find(&members).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return members, nil
}

func (r *groupRepository) FindActiveMemberByID(ctx context.Context, groupID, memberID int64) (*entity.GroupMember, error) {
	var member entity.GroupMember
	if err := r.getDB(ctx).Where("group_id = ? AND id = ? AND status = ?", groupID, memberID, enum.GroupMemberStatusActive).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrGroupMemberNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &member, nil
}

func (r *groupRepository) SetMemberInactive(ctx context.Context, groupID, userID int64) error {
	return r.setMemberInactive(ctx, groupID, "user_id", userID)
}

func (r *groupRepository) SetMemberInactiveByID(ctx context.Context, groupID, memberID int64) error {
	return r.setMemberInactive(ctx, groupID, "id", memberID)
}

func (r *groupRepository) setMemberInactive(ctx context.Context, groupID int64, idColumn string, id int64) error {
	result := r.getDB(ctx).Model(&entity.GroupMember{}).
		Where("group_id = ? AND "+idColumn+" = ? AND status = ?", groupID, id, enum.GroupMemberStatusActive).
		Updates(map[string]interface{}{"status": enum.GroupMemberStatusInactive, "left_at": gorm.Expr("CURRENT_TIMESTAMP")})
	if result.Error != nil {
		return exception.Wrap(exception.ErrInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrGroupMemberNotFound
	}
	return nil
}

func (r *groupRepository) TransferOwnership(ctx context.Context, groupID, newOwnerID int64) error {
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var group entity.Group
		if err := tx.Where("id = ? AND deleted_at IS NULL", groupID).First(&group).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrGroupNotFound
			}
			return err
		}

		var member entity.GroupMember
		if err := tx.Where("group_id = ? AND user_id = ? AND status = ?", groupID, newOwnerID, enum.GroupMemberStatusActive).First(&member).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrGroupTransferTargetNotMember
			}
			return err
		}

		return tx.Model(&entity.Group{}).Where("id = ?", groupID).Update("created_by_user_id", newOwnerID).Error
	})
	if err != nil {
		var appErr *exception.AppError
		if errors.As(err, &appErr) {
			return err
		}
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *groupRepository) CreateInvite(ctx context.Context, groupID, creatorUserID int64, invite *entity.GroupInvite) error {
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var group entity.Group
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", groupID).First(&group).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrGroupNotFound
			}
			return err
		}

		var activeInvite entity.GroupInvite
		err := tx.Where("group_id = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP) AND (max_uses IS NULL OR used_count < max_uses)", groupID).
			First(&activeInvite).Error
		if err == nil {
			return exception.ErrGroupActiveInviteExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var creator entity.GroupMember
		err = tx.Where("group_id = ? AND user_id = ? AND status = ?", groupID, creatorUserID, enum.GroupMemberStatusActive).First(&creator).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = tx.Where("group_id = ? AND user_id = ? AND status = ?", groupID, group.CreatedByUserID, enum.GroupMemberStatusActive).First(&creator).Error
		}
		if err != nil {
			return err
		}
		invite.CreatedByMemberID = creator.ID
		return tx.Create(invite).Error
	})
	if err != nil {
		var appErr *exception.AppError
		if errors.As(err, &appErr) {
			return err
		}
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *groupRepository) FindInviteByID(ctx context.Context, groupID, inviteID int64) (*entity.GroupInvite, error) {
	var invite entity.GroupInvite
	if err := r.getDB(ctx).Where("group_id = ? AND id = ?", groupID, inviteID).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrInviteNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &invite, nil
}

func (r *groupRepository) FindInvitesByGroupID(ctx context.Context, groupID int64) ([]entity.GroupInvite, error) {
	var invites []entity.GroupInvite
	if err := r.getDB(ctx).Where("group_id = ?", groupID).Order("created_at DESC").Find(&invites).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return invites, nil
}

func (r *groupRepository) FindInviteByTokenHash(ctx context.Context, tokenHash string) (*entity.GroupInvite, error) {
	var invite entity.GroupInvite
	if err := r.getDB(ctx).Where("token_hash = ?", tokenHash).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrInviteNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &invite, nil
}

func (r *groupRepository) RevokeInvite(ctx context.Context, groupID, inviteID int64) error {
	result := r.getDB(ctx).Model(&entity.GroupInvite{}).Where("group_id = ? AND id = ? AND revoked_at IS NULL", groupID, inviteID).
		Update("revoked_at", gorm.Expr("CURRENT_TIMESTAMP"))
	if result.Error != nil {
		return exception.Wrap(exception.ErrInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindInviteByID(ctx, groupID, inviteID); err != nil {
			return err
		}
	}
	return nil
}

func (r *groupRepository) UpdateInvitePassword(ctx context.Context, groupID, inviteID int64, passwordHash string) error {
	result := r.getDB(ctx).Model(&entity.GroupInvite{}).Where("group_id = ? AND id = ?", groupID, inviteID).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return exception.Wrap(exception.ErrInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindInviteByID(ctx, groupID, inviteID); err != nil {
			return err
		}
	}
	return nil
}

func (r *groupRepository) UpdateInvite(ctx context.Context, invite *entity.GroupInvite) error {
	updates := map[string]interface{}{}
	if invite.ExpiresAt != nil {
		updates["expires_at"] = invite.ExpiresAt
	}
	if invite.MaxUses != nil {
		updates["max_uses"] = invite.MaxUses
	}
	result := r.getDB(ctx).Model(&entity.GroupInvite{}).Where("group_id = ? AND id = ?", invite.GroupID, invite.ID).Updates(updates)
	if result.Error != nil {
		return exception.Wrap(exception.ErrInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindInviteByID(ctx, invite.GroupID, invite.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *groupRepository) JoinGroupWithInvite(ctx context.Context, inviteID, userID int64) (int64, error) {
	var groupID int64
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var invite entity.GroupInvite
		if err := tx.Select("id", "group_id").Where("id = ?", inviteID).First(&invite).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrInviteNotFound
			}
			return err
		}

		var group entity.Group
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", invite.GroupID).First(&group).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrGroupNotFound
			}
			return err
		}
		if group.Status != enum.GroupStatusActive {
			return exception.ErrGroupInactive
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", inviteID).First(&invite).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrInviteNotFound
			}
			return err
		}
		if invite.RevokedAt != nil {
			return exception.ErrInviteRevoked
		}
		if invite.ExpiresAt != nil && !invite.ExpiresAt.After(time.Now()) {
			return exception.ErrInviteExpired
		}
		if invite.MaxUses != nil && invite.UsedCount >= *invite.MaxUses {
			return exception.ErrInviteUsesExhausted
		}

		var member entity.GroupMember
		err := tx.Where("group_id = ? AND user_id = ?", invite.GroupID, userID).First(&member).Error
		if err == nil && member.Status == enum.GroupMemberStatusActive {
			return exception.ErrAlreadyGroupMember
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			member = entity.GroupMember{GroupID: invite.GroupID, UserID: userID, Status: enum.GroupMemberStatusActive}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		} else if err := tx.Model(&member).Updates(map[string]interface{}{
			"status": enum.GroupMemberStatusActive, "joined_at": gorm.Expr("CURRENT_TIMESTAMP"), "left_at": nil,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&invite).Update("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
			return err
		}
		groupID = invite.GroupID
		return nil
	})
	if err != nil {
		var appErr *exception.AppError
		if errors.As(err, &appErr) {
			return 0, err
		}
		return 0, exception.Wrap(exception.ErrInternal, err)
	}
	return groupID, nil
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
