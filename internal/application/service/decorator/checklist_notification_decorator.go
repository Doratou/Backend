package decorator

import (
	"context"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type checklistNotificationDecorator struct {
	next          adapter.ChecklistService
	notifSvc      adapter.NotificationService
	groupRepo     domain_db.GroupRepository
	userRepo      domain_db.UserRepository
	checklistRepo domain_db.ChecklistRepository
}

// NewChecklistNotificationDecorator wraps an adapter.ChecklistService to dispatch group notifications
// whenever checklist items are created, claimed/released, or deleted.
func NewChecklistNotificationDecorator(
	next adapter.ChecklistService,
	notifSvc adapter.NotificationService,
	groupRepo domain_db.GroupRepository,
	userRepo domain_db.UserRepository,
	checklistRepo domain_db.ChecklistRepository,
) adapter.ChecklistService {
	return &checklistNotificationDecorator{
		next:          next,
		notifSvc:      notifSvc,
		groupRepo:     groupRepo,
		userRepo:      userRepo,
		checklistRepo: checklistRepo,
	}
}

func (d *checklistNotificationDecorator) Get(ctx context.Context, userID, groupID int64, req dto.GetChecklistRequest) (*dto.ChecklistResponse, error) {
	return d.next.Get(ctx, userID, groupID, req)
}

func (d *checklistNotificationDecorator) CreateItem(ctx context.Context, userID, groupID int64, req dto.CreateChecklistItemRequest) (*dto.ChecklistItemResponse, error) {
	res, err := d.next.CreateItem(ctx, userID, groupID, req)
	if err != nil {
		return nil, err
	}

	var updatedByName string
	if res.CreatedBy != nil {
		updatedByName = res.CreatedBy.Name
	}
	d.notifyChecklistUpdated(ctx, userID, groupID, updatedByName, res.UpdatedAt)
	return res, nil
}

func (d *checklistNotificationDecorator) SetClaim(ctx context.Context, userID, groupID, itemID int64, claimed bool) (*dto.ChecklistItemResponse, error) {
	var priorFound bool
	var wasClaimed bool
	if d.checklistRepo != nil {
		if existing, fErr := d.checklistRepo.FindItemForUpdate(ctx, groupID, itemID); fErr == nil && existing != nil {
			priorFound = true
			wasClaimed = existing.AssignedToMemberID != nil
		}
	}

	res, err := d.next.SetClaim(ctx, userID, groupID, itemID, claimed)
	if err != nil {
		return nil, err
	}

	// Only notify if the claim state actually changed.
	if !priorFound || wasClaimed != claimed {
		d.notifyChecklistUpdated(ctx, userID, groupID, "", res.UpdatedAt)
	}
	return res, nil
}

func (d *checklistNotificationDecorator) DeleteItem(ctx context.Context, userID, groupID, itemID int64) error {
	if err := d.next.DeleteItem(ctx, userID, groupID, itemID); err != nil {
		return err
	}

	d.notifyChecklistUpdated(ctx, userID, groupID, "", time.Now().UTC())
	return nil
}

func (d *checklistNotificationDecorator) notifyChecklistUpdated(ctx context.Context, userID, groupID int64, updatedByName string, updatedAt time.Time) {
	if d.notifSvc == nil {
		return
	}
	groupName := resolveGroupTitle(ctx, d.groupRepo, groupID)
	if updatedByName == "" {
		updatedByName, _ = resolveUserProfile(ctx, d.userRepo, userID)
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	dispatchGroupNotification(ctx, d.notifSvc, groupID, &userID, payload.WSEventChecklistUpdated, payload.ChecklistUpdatedData{
		GroupID:       groupID,
		GroupName:     groupName,
		UpdatedByID:   strconv.FormatInt(userID, 10),
		UpdatedByName: updatedByName,
		UpdatedAt:     updatedAt,
	})
}
