package decorator

import (
	"context"
	"strconv"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type settlementNotificationDecorator struct {
	next      adapter.SettlementService
	notifSvc  adapter.NotificationService
	groupRepo domain_db.GroupRepository
}

// NewSettlementNotificationDecorator wraps an adapter.SettlementService to notify the creditor
// when a debtor records a settlement payment.
func NewSettlementNotificationDecorator(
	next adapter.SettlementService,
	notifSvc adapter.NotificationService,
	groupRepo domain_db.GroupRepository,
) adapter.SettlementService {
	return &settlementNotificationDecorator{
		next:      next,
		notifSvc:  notifSvc,
		groupRepo: groupRepo,
	}
}

func (d *settlementNotificationDecorator) Create(ctx context.Context, userID, groupID int64, req dto.CreateSettlementRequest) (*dto.SettlementResponse, error) {
	res, err := d.next.Create(ctx, userID, groupID, req)
	if err != nil {
		return nil, err
	}

	groupName := resolveGroupTitle(ctx, d.groupRepo, groupID)
	dispatchUserNotification(ctx, d.notifSvc, res.ToMember.UserID, &userID, &groupID, payload.WSEventSettlementPaid, payload.SettlementPaidData{
		SettlementID: res.ID,
		GroupID:      res.GroupID,
		GroupName:    groupName,
		DebtorID:     strconv.FormatInt(res.FromMember.UserID, 10),
		DebtorName:   res.FromMember.Name,
		CreditorID:   strconv.FormatInt(res.ToMember.UserID, 10),
		CreditorName: res.ToMember.Name,
		AmountToman:  res.AmountToman,
		SettledAt:    res.PaidAt,
	})

	return res, nil
}

func (d *settlementNotificationDecorator) List(ctx context.Context, userID, groupID int64, req dto.ListSettlementsRequest) (*dto.SettlementListResponse, error) {
	return d.next.List(ctx, userID, groupID, req)
}

func (d *settlementNotificationDecorator) Get(ctx context.Context, userID, groupID, id int64) (*dto.SettlementResponse, error) {
	return d.next.Get(ctx, userID, groupID, id)
}

func (d *settlementNotificationDecorator) Confirm(ctx context.Context, userID, groupID, id int64) (*dto.SettlementResponse, error) {
	return d.next.Confirm(ctx, userID, groupID, id)
}

func (d *settlementNotificationDecorator) Reject(ctx context.Context, userID, groupID, id int64) (*dto.SettlementResponse, error) {
	return d.next.Reject(ctx, userID, groupID, id)
}

func (d *settlementNotificationDecorator) Delete(ctx context.Context, userID, groupID, id int64) error {
	return d.next.Delete(ctx, userID, groupID, id)
}
