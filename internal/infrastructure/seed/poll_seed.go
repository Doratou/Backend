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

type pollSeedData struct {
	Question   string
	Options    []string
	Visibility enum.PollVisibility
	Closed     bool
	Creator    int
}

func SeedPollTestData(ctx context.Context, db database.Database) error {
	if db == nil {
		return nil
	}
	created := 0
	err := db.GetGormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(2026111101)).Error; err != nil {
			return err
		}
		groups, err := activitySeedGroups(tx)
		if err != nil {
			return err
		}
		for _, target := range groups {
			samples := []pollSeedData{
				{Question: "[نمونه] مقصد سفر بعدی کجا باشد؟", Options: []string{"شمال", "شیراز", "اصفهان"}, Visibility: enum.PollVisibilityPublic},
				{Question: "[نمونه] بودجه مناسب برای سفر چقدر است؟", Options: []string{"اقتصادی", "متوسط", "آزاد"}, Visibility: enum.PollVisibilityPrivate, Creator: 1},
				{Question: "[نمونه] روز دورهمی قبلی", Options: []string{"پنجشنبه", "جمعه", "شنبه"}, Visibility: enum.PollVisibilityPublic, Closed: true},
			}
			if target.Group.ID != 1 {
				samples = []pollSeedData{{Question: "[نمونه] نظرسنجی گروه غیرفعال", Options: []string{"گزینه اول", "گزینه دوم"}, Visibility: enum.PollVisibilityPublic}}
			}
			for index, sample := range samples {
				createdAt := activitySeedTime(target.Group, index)
				exists, err := activitySeedExists(tx, &entity.Poll{}, target.Group.ID, createdAt)
				if err != nil {
					return err
				}
				if exists {
					continue
				}
				poll := entity.Poll{
					BaseEntity:        entity.BaseEntity{CreatedAt: createdAt, UpdatedAt: createdAt},
					GroupID:           target.Group.ID,
					CreatedByMemberID: target.Members[sample.Creator%len(target.Members)].ID,
					Question:          sample.Question,
					Visibility:        sample.Visibility,
					Status:            enum.PollStatusActive,
				}
				if sample.Closed {
					closedAt := time.Now().UTC()
					poll.Status = enum.PollStatusInactive
					poll.ClosedAt = &closedAt
					poll.UpdatedAt = closedAt
				}
				if err := tx.Omit(clause.Associations).Create(&poll).Error; err != nil {
					return fmt.Errorf("create poll sample for group %d: %w", target.Group.ID, err)
				}
				options := make([]entity.PollOption, len(sample.Options))
				for i, title := range sample.Options {
					options[i] = entity.PollOption{BaseEntity: entity.BaseEntity{CreatedAt: createdAt, UpdatedAt: createdAt}, PollID: poll.ID, Title: title}
				}
				if err := tx.Omit(clause.Associations).Create(&options).Error; err != nil {
					return fmt.Errorf("create options for poll sample %d: %w", poll.ID, err)
				}
				votes := make([]entity.PollVote, len(target.Members))
				for i, member := range target.Members {
					votedAt := createdAt.Add(time.Duration(i+1) * time.Microsecond)
					votes[i] = entity.PollVote{
						BaseEntity: entity.BaseEntity{CreatedAt: votedAt, UpdatedAt: votedAt},
						PollID:     poll.ID,
						MemberID:   member.ID,
						OptionID:   options[i%len(options)].ID,
					}
				}
				if err := tx.Omit(clause.Associations).Create(&votes).Error; err != nil {
					return fmt.Errorf("create votes for poll sample %d: %w", poll.ID, err)
				}
				created++
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Printf("[SEED TEST] Poll samples ready: %d created; existing polls and votes preserved", created)
	return nil
}
