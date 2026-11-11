package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"gorm.io/gorm"
)

type activitySeedGroup struct {
	Group   entity.Group
	Members []entity.GroupMember
}

func activitySeedGroups(tx *gorm.DB) ([]activitySeedGroup, error) {
	var primary entity.Group
	if err := tx.First(&primary, 1).Error; err != nil {
		return nil, fmt.Errorf("find group 1 for activity samples: %w", err)
	}
	groups := []entity.Group{primary}
	var inactive []entity.Group
	if err := tx.Where("status = ? AND id <> ?", enum.GroupStatusInactive, primary.ID).
		Order("id ASC").Limit(1).Find(&inactive).Error; err != nil {
		return nil, fmt.Errorf("find inactive group for activity samples: %w", err)
	}
	groups = append(groups, inactive...)

	result := make([]activitySeedGroup, 0, len(groups))
	for _, group := range groups {
		var members []entity.GroupMember
		if err := tx.Model(&entity.GroupMember{}).Select("group_members.*").
			Joins("JOIN users ON users.id = group_members.user_id AND users.deleted_at IS NULL").
			Where("group_members.group_id = ? AND group_members.status = ? AND group_members.left_at IS NULL AND users.status = ?",
				group.ID, enum.GroupMemberStatusActive, enum.UserStatusActive).
			Order("CASE users.email WHEN 'owner@test.com' THEN 0 WHEN 'member@test.com' THEN 1 ELSE 2 END").
			Order("group_members.id ASC").Limit(5).Find(&members).Error; err != nil {
			return nil, fmt.Errorf("find active members for group %d: %w", group.ID, err)
		}
		if len(members) == 0 {
			if group.ID == 1 {
				return nil, fmt.Errorf("group 1 needs an active member for activity samples")
			}
			log.Printf("[SEED NOTICE] Skipping activity samples for group %d: no active members", group.ID)
			continue
		}
		result = append(result, activitySeedGroup{Group: group, Members: members})
	}
	return result, nil
}

func activitySeedTime(group entity.Group, index int) time.Time {
	return group.CreatedAt.UTC().Truncate(time.Microsecond).Add(time.Duration(index+1) * time.Microsecond)
}

func activitySeedExists(tx *gorm.DB, model any, groupID int64, createdAt time.Time) (bool, error) {
	var count int64
	err := tx.Unscoped().Model(model).Where("group_id = ? AND created_at = ?", groupID, createdAt).Count(&count).Error
	return count > 0, err
}
