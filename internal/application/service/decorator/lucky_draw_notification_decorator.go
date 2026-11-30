package decorator

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/enum"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type luckyDrawNotificationDecorator struct {
	next      adapter.LuckyDrawService
	notifSvc  adapter.NotificationService
	groupRepo domain_db.GroupRepository
	userRepo  domain_db.UserRepository
}

// NewLuckyDrawNotificationDecorator wraps an adapter.LuckyDrawService to dispatch group notifications
// when lucky draws are created or completed.
func NewLuckyDrawNotificationDecorator(
	next adapter.LuckyDrawService,
	notifSvc adapter.NotificationService,
	groupRepo domain_db.GroupRepository,
	userRepo domain_db.UserRepository,
) adapter.LuckyDrawService {
	return &luckyDrawNotificationDecorator{
		next:      next,
		notifSvc:  notifSvc,
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (d *luckyDrawNotificationDecorator) Create(ctx context.Context, userID, groupID int64, req dto.CreateLuckyDrawRequest) (*dto.LuckyDrawResponse, error) {
	res, err := d.next.Create(ctx, userID, groupID, req)
	if err != nil {
		return nil, err
	}

	groupName := resolveGroupTitle(ctx, d.groupRepo, groupID)
	creatorName, _ := resolveUserProfile(ctx, d.userRepo, userID)
	candidates := make([]string, 0, len(res.Candidates))
	for _, c := range res.Candidates {
		_, displayName := d.resolveCandidate(ctx, c.Type, c.Value)
		candidates = append(candidates, displayName)
	}

	dispatchGroupNotification(ctx, d.notifSvc, groupID, &userID, payload.WSEventDrawCreated, payload.DrawCreatedData{
		DrawID:      res.ID,
		GroupID:     groupID,
		GroupName:   groupName,
		CreatorID:   strconv.FormatInt(userID, 10),
		CreatorName: creatorName,
		Title:       res.Title,
		Candidates:  candidates,
		CreatedAt:   res.CreatedAt,
	})

	return res, nil
}

func (d *luckyDrawNotificationDecorator) Run(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error) {
	// Check prior status so idempotent re-calls on an already completed draw do not emit duplicate notifications.
	var alreadyCompleted bool
	if prev, prevErr := d.next.Get(ctx, userID, groupID, drawID); prevErr == nil && prev != nil {
		alreadyCompleted = prev.Status == enum.LuckyDrawStatusCompleted
	}

	res, err := d.next.Run(ctx, userID, groupID, drawID)
	if err != nil {
		return nil, err
	}

	if !alreadyCompleted && res.Status == enum.LuckyDrawStatusCompleted && res.Result != nil {
		groupName := resolveGroupTitle(ctx, d.groupRepo, groupID)
		winnerID, winnerName := d.resolveCandidate(ctx, res.Result.Type, res.Result.Value)
		drawnAt := time.Now().UTC()
		if res.CompletedAt != nil {
			drawnAt = *res.CompletedAt
		}

		dispatchGroupNotification(ctx, d.notifSvc, groupID, &userID, payload.WSEventDrawResult, payload.DrawResultData{
			DrawID:     res.ID,
			GroupID:    groupID,
			GroupName:  groupName,
			Title:      res.Title,
			WinnerID:   winnerID,
			WinnerName: winnerName,
			DrawnAt:    drawnAt,
		})
	}

	return res, nil
}

func (d *luckyDrawNotificationDecorator) resolveCandidate(ctx context.Context, candidateType enum.LuckyDrawCandidateType, raw json.RawMessage) (string, string) {
	switch candidateType {
	case enum.LuckyDrawCandidateTypeUser:
		var uid int64
		if err := json.Unmarshal(raw, &uid); err == nil && uid > 0 {
			uidStr := strconv.FormatInt(uid, 10)
			name, _ := resolveUserProfile(ctx, d.userRepo, uid)
			if name == "" {
				name = uidStr
			}
			return uidStr, name
		}
	case enum.LuckyDrawCandidateTypeText:
		var title string
		if err := json.Unmarshal(raw, &title); err == nil {
			return "", title
		}
	}
	return "", string(raw)
}

func (d *luckyDrawNotificationDecorator) List(ctx context.Context, userID, groupID int64, req dto.ListLuckyDrawsRequest) (*dto.PaginatedResponse[dto.LuckyDrawListItem], error) {
	return d.next.List(ctx, userID, groupID, req)
}

func (d *luckyDrawNotificationDecorator) Get(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error) {
	return d.next.Get(ctx, userID, groupID, drawID)
}

func (d *luckyDrawNotificationDecorator) Update(ctx context.Context, userID, groupID, drawID int64, req dto.UpdateLuckyDrawRequest) (*dto.LuckyDrawResponse, error) {
	return d.next.Update(ctx, userID, groupID, drawID, req)
}
