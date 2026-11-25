package decorator

import (
	"context"
	"log"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

func resolveGroupTitle(ctx context.Context, groupRepo domain_db.GroupRepository, groupID int64) string {
	if groupRepo == nil || groupID <= 0 {
		return ""
	}
	group, _, err := groupRepo.FindByID(ctx, groupID)
	if err != nil || group == nil {
		return ""
	}
	return group.Title
}

func resolveUserProfile(ctx context.Context, userRepo domain_db.UserRepository, userID int64) (string, int64) {
	if userRepo == nil || userID <= 0 {
		return "", 0
	}
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return "", 0
	}
	var avatarID int64
	if user.AvatarAttachmentID != nil {
		avatarID = *user.AvatarAttachmentID
	}
	return user.Name, avatarID
}

func dispatchUserNotification(ctx context.Context, notifSvc adapter.NotificationService, recipientID int64, actorID, groupID *int64, event payload.WSEvent, data any) {
	if notifSvc == nil || recipientID <= 0 {
		return
	}
	if err := notifSvc.DispatchToUser(ctx, recipientID, actorID, groupID, event, data); err != nil {
		log.Printf("[NOTIFICATION WARNING] failed to dispatch %q to user %d: %v", event, recipientID, err)
	}
}

func dispatchGroupNotification(ctx context.Context, notifSvc adapter.NotificationService, groupID int64, actorID *int64, event payload.WSEvent, data any) {
	if notifSvc == nil || groupID <= 0 {
		return
	}
	if err := notifSvc.DispatchToGroup(ctx, groupID, actorID, nil, event, data); err != nil {
		log.Printf("[NOTIFICATION WARNING] failed to dispatch %q to group %d: %v", event, groupID, err)
	}
}

func dispatchBroadcastNotification(ctx context.Context, notifSvc adapter.NotificationService, actorID *int64, event payload.WSEvent, data any) {
	if notifSvc == nil {
		return
	}
	if err := notifSvc.Broadcast(ctx, actorID, event, data); err != nil {
		log.Printf("[NOTIFICATION WARNING] failed to broadcast %q: %v", event, err)
	}
}
