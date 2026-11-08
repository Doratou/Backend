package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type inviteRepository struct{ db database.Database }

func NewInviteRepository(db database.Database) DB.InviteRepository {
	return &inviteRepository{db: db}
}

func (r *inviteRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *inviteRepository) CreateDirect(ctx context.Context, invite *entity.DirectInvite, rejectCooldown time.Duration) error {
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
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
		if group.CreatedByUserID != invite.SenderUserID {
			return exception.ErrDirectInviteForbidden
		}

		var member entity.GroupMember
		err := tx.Where("group_id = ? AND user_id = ?", invite.GroupID, invite.RecipientUserID).First(&member).Error
		if err == nil && member.Status == enum.GroupMemberStatusActive {
			return exception.ErrAlreadyGroupMember
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var previous entity.DirectInvite
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("group_id = ? AND recipient_user_id = ?", invite.GroupID, invite.RecipientUserID).
			Order("created_at DESC").Order("id DESC").First(&previous).Error
		if err == nil {
			switch previous.Status {
			case enum.DirectInviteStatusPending:
				return exception.ErrDirectInvitePending
			case enum.DirectInviteStatusRejected:
				if previous.RespondedAt != nil && tx.NowFunc().Before(previous.RespondedAt.Add(rejectCooldown)) {
					return exception.ErrDirectInviteCooldown
				}
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

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

func (r *inviteRepository) RespondDirect(ctx context.Context, inviteID, recipientID int64, accept bool) (*entity.DirectInvite, error) {
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var invite entity.DirectInvite
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND recipient_user_id = ?", inviteID, recipientID).First(&invite).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrDirectInviteNotFound
			}
			return err
		}
		if invite.Status != enum.DirectInviteStatusPending {
			return exception.ErrDirectInviteNotPending
		}
		now := tx.NowFunc()
		status := enum.DirectInviteStatusRejected
		if accept {
			status = enum.DirectInviteStatusAccepted
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
			var member entity.GroupMember
			err := tx.Where("group_id = ? AND user_id = ?", invite.GroupID, recipientID).First(&member).Error
			if err == nil && member.Status == enum.GroupMemberStatusActive {
				return exception.ErrAlreadyGroupMember
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				member = entity.GroupMember{GroupID: invite.GroupID, UserID: recipientID, Status: enum.GroupMemberStatusActive}
				if err := tx.Create(&member).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if err := tx.Model(&member).Updates(map[string]interface{}{
				"status": enum.GroupMemberStatusActive, "joined_at": now, "left_at": nil,
			}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&invite).Updates(map[string]interface{}{"status": status, "responded_at": now}).Error
	})
	if err != nil {
		var appErr *exception.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return r.findDirectByID(ctx, inviteID)
}

func (r *inviteRepository) FindDirectByRecipient(ctx context.Context, recipientID int64) ([]entity.DirectInvite, error) {
	var invites []entity.DirectInvite
	err := r.getDB(ctx).Preload("Group").Preload("Sender.AvatarAttachment").Where("recipient_user_id = ?", recipientID).Order("created_at DESC").Order("id DESC").Find(&invites).Error
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return invites, nil
}

func (r *inviteRepository) FindDirectByGroup(ctx context.Context, groupID int64) ([]entity.DirectInvite, error) {
	var invites []entity.DirectInvite
	err := r.getDB(ctx).Preload("Recipient.AvatarAttachment").Where("group_id = ?", groupID).Order("created_at DESC").Order("id DESC").Find(&invites).Error
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return invites, nil
}

func (r *inviteRepository) findDirectByID(ctx context.Context, inviteID int64) (*entity.DirectInvite, error) {
	var invite entity.DirectInvite
	if err := r.getDB(ctx).Preload("Group").Preload("Sender.AvatarAttachment").Preload("Recipient.AvatarAttachment").Where("id = ?", inviteID).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrDirectInviteNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &invite, nil
}

func (r *inviteRepository) CreateLink(ctx context.Context, groupID, creatorUserID int64, invite *entity.LinkInvite) error {
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var group entity.Group
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", groupID).First(&group).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return exception.ErrGroupNotFound
			}
			return err
		}

		var activeInvite entity.LinkInvite
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

func (r *inviteRepository) FindLinkByID(ctx context.Context, groupID, inviteID int64) (*entity.LinkInvite, error) {
	var invite entity.LinkInvite
	if err := r.getDB(ctx).Where("group_id = ? AND id = ?", groupID, inviteID).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrInviteNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &invite, nil
}

func (r *inviteRepository) FindLinksByGroupID(ctx context.Context, groupID int64) ([]entity.LinkInvite, error) {
	var invites []entity.LinkInvite
	if err := r.getDB(ctx).Where("group_id = ?", groupID).Order("created_at DESC").Find(&invites).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return invites, nil
}

func (r *inviteRepository) FindLinkByTokenHash(ctx context.Context, tokenHash string) (*entity.LinkInvite, error) {
	var invite entity.LinkInvite
	if err := r.getDB(ctx).Where("token_hash = ?", tokenHash).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrInviteNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &invite, nil
}

func (r *inviteRepository) RevokeLink(ctx context.Context, groupID, inviteID int64) error {
	result := r.getDB(ctx).Model(&entity.LinkInvite{}).Where("group_id = ? AND id = ? AND revoked_at IS NULL", groupID, inviteID).
		Update("revoked_at", gorm.Expr("CURRENT_TIMESTAMP"))
	if result.Error != nil {
		return exception.Wrap(exception.ErrInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindLinkByID(ctx, groupID, inviteID); err != nil {
			return err
		}
	}
	return nil
}

func (r *inviteRepository) UpdateLinkPassword(ctx context.Context, groupID, inviteID int64, passwordHash string) error {
	result := r.getDB(ctx).Model(&entity.LinkInvite{}).Where("group_id = ? AND id = ?", groupID, inviteID).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return exception.Wrap(exception.ErrInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindLinkByID(ctx, groupID, inviteID); err != nil {
			return err
		}
	}
	return nil
}

func (r *inviteRepository) UpdateLink(ctx context.Context, invite *entity.LinkInvite) error {
	updates := map[string]interface{}{}
	if invite.ExpiresAt != nil {
		updates["expires_at"] = invite.ExpiresAt
	}
	if invite.MaxUses != nil {
		updates["max_uses"] = invite.MaxUses
	}
	result := r.getDB(ctx).Model(&entity.LinkInvite{}).Where("group_id = ? AND id = ?", invite.GroupID, invite.ID).Updates(updates)
	if result.Error != nil {
		return exception.Wrap(exception.ErrInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindLinkByID(ctx, invite.GroupID, invite.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *inviteRepository) JoinGroupWithLink(ctx context.Context, inviteID, userID int64) (int64, error) {
	var groupID int64
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var invite entity.LinkInvite
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
