package seed

import (
	"context"
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
			samples := []luckyDrawSeedData{
				{Title: "[نمونه] انتخاب غذای دورهمی", Candidates: []entity.LuckyDrawCandidate{{Title: "پیتزا", Weight: 1}, {Title: "کباب", Weight: 3}, {Title: "پاستا", Weight: 2}}},
				{Title: "[نمونه] انتخاب بازی دورهمی", Candidates: []entity.LuckyDrawCandidate{{Title: "مافیا", Weight: 2}, {Title: "پانتومیم", Weight: 3}, {Title: "اسم فامیل", Weight: 5}}, Creator: 1},
				{Title: "[نمونه] قرعه‌کشی انجام‌شده مقصد سفر", Candidates: []entity.LuckyDrawCandidate{{Title: "شمال", Weight: 1}, {Title: "شیراز", Weight: 3}, {Title: "اصفهان", Weight: 2}}, Completed: true},
			}
			if target.Group.ID != 1 {
				samples = []luckyDrawSeedData{{Title: "[نمونه] قرعه‌کشی گروه غیرفعال", Candidates: []entity.LuckyDrawCandidate{{Title: "گزینه اول", Weight: 1}, {Title: "گزینه دوم", Weight: 2}}}}
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
				draw := entity.LuckyDraw{
					BaseEntity:        entity.BaseEntity{CreatedAt: createdAt, UpdatedAt: createdAt},
					GroupID:           target.Group.ID,
					CreatedByMemberID: target.Members[sample.Creator%len(target.Members)].ID,
					Title:             sample.Title,
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
