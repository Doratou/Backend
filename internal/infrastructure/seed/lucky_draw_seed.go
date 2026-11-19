package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type luckyDrawSeedData struct {
	Title      string
	Candidates []entity.LuckyDrawCandidate
	DrawType   enum.LuckyDrawType
	Completed  bool
	Creator    int
}

func SeedLuckyDrawTestData(ctx context.Context, db database.Database) error {
	if db == nil {
		return nil
	}
	created := 0
	err := db.GetGormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(2026111102)).Error; err != nil {
			return err
		}
		groups, err := activitySeedGroups(tx)
		if err != nil {
			return err
		}
		for _, target := range groups {
			textCandidate := func(title string, percent entity.LuckyDrawPercentage) entity.LuckyDrawCandidate {
				value, _ := json.Marshal(title)
				return entity.LuckyDrawCandidate{Type: enum.LuckyDrawCandidateTypeText, Value: value, Weight: percent}
			}
			userCandidate := func(userID int64, percent entity.LuckyDrawPercentage) entity.LuckyDrawCandidate {
				value, _ := json.Marshal(userID)
				return entity.LuckyDrawCandidate{Type: enum.LuckyDrawCandidateTypeUser, Value: value, Weight: percent}
			}
			samples := []luckyDrawSeedData{
				{Title: "[نمونه] انتخاب غذای دورهمی", DrawType: enum.LuckyDrawTypeWeighted, Candidates: []entity.LuckyDrawCandidate{textCandidate("پیتزا", 2000), textCandidate("کباب", 5000), textCandidate("پاستا", 3000)}},
				{Title: "[نمونه] انتخاب بازی دورهمی", DrawType: enum.LuckyDrawTypeWeighted, Candidates: []entity.LuckyDrawCandidate{textCandidate("مافیا", 2000), textCandidate("پانتومیم", 3000), textCandidate("اسم فامیل", 5000)}, Creator: 1, Completed: true},
				{Title: "[نمونه] قرعه‌کشی انجام‌شده مقصد سفر", DrawType: enum.LuckyDrawTypeEqual, Candidates: []entity.LuckyDrawCandidate{textCandidate("شمال", 3334), textCandidate("شیراز", 3333), textCandidate("اصفهان", 3333)}, Completed: true},
			}
			if len(target.Members) >= 2 {
				samples[0].Title = "[نمونه] چه کسی شام حساب کند؟"
				samples[0].Candidates = []entity.LuckyDrawCandidate{userCandidate(target.Members[0].UserID, 3000), userCandidate(target.Members[1].UserID, 5000), textCandidate("مهمان گروه", 2000)}
			}
			if target.Group.ID != 1 {
				samples = []luckyDrawSeedData{{Title: "[نمونه] قرعه‌کشی گروه غیرفعال", DrawType: enum.LuckyDrawTypeEqual, Candidates: []entity.LuckyDrawCandidate{textCandidate("گزینه اول", 5000), textCandidate("گزینه دوم", 5000)}}}
			}
			for index, sample := range samples {
				createdAt := activitySeedTime(target.Group, index)
				exists, err := activitySeedExists(tx, &entity.LuckyDraw{}, target.Group.ID, createdAt)
				if err != nil {
					return err
				}
				if exists {
					continue
				}
				if !sample.Completed {
					var drafts int64
					if err := tx.Model(&entity.LuckyDraw{}).Where("group_id = ? AND status = ?", target.Group.ID, enum.LuckyDrawStatusDraft).Count(&drafts).Error; err != nil {
						return err
					}
					if drafts > 0 {
						continue
					}
				}
				draw := entity.LuckyDraw{
					BaseEntity:        entity.BaseEntity{CreatedAt: createdAt, UpdatedAt: createdAt},
					GroupID:           target.Group.ID,
					CreatedByMemberID: target.Members[sample.Creator%len(target.Members)].ID,
					Title:             sample.Title,
					DrawType:          sample.DrawType,
					Candidates:        sample.Candidates,
					Status:            enum.LuckyDrawStatusDraft,
				}
				if sample.Completed {
					resultIndex := 1
					completedAt := time.Now().UTC()
					draw.Status = enum.LuckyDrawStatusCompleted
					draw.ResultIndex = &resultIndex
					draw.CompletedAt = &completedAt
					draw.UpdatedAt = completedAt
				}
				if err := tx.Omit(clause.Associations).Create(&draw).Error; err != nil {
					return fmt.Errorf("create lucky draw sample for group %d: %w", target.Group.ID, err)
				}
				created++
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Printf("[SEED TEST] Lucky draw samples ready: %d created; existing drafts and results preserved", created)
	return nil
}
