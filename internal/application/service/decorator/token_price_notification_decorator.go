package decorator

import (
	"context"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
)

type tokenPriceNotificationDecorator struct {
	next     adapter.TokenPriceService
	notifSvc adapter.NotificationService
}

// NewTokenPriceNotificationDecorator wraps an adapter.TokenPriceService to broadcast
// token.price_updated events when the token price changes.
func NewTokenPriceNotificationDecorator(
	next adapter.TokenPriceService,
	notifSvc adapter.NotificationService,
) adapter.TokenPriceService {
	return &tokenPriceNotificationDecorator{
		next:     next,
		notifSvc: notifSvc,
	}
}

func (d *tokenPriceNotificationDecorator) UpdatePrice(ctx context.Context, adminUserID int64, req dto.UpdateTokenPriceRequest) (*dto.TokenPriceResponse, error) {
	res, err := d.next.UpdatePrice(ctx, adminUserID, req)
	if err != nil {
		return nil, err
	}

	if d.notifSvc != nil && res.NewPriceToman != res.OldPriceToman {
		dispatchBroadcastNotification(ctx, d.notifSvc, &adminUserID, payload.WSEventTokenPriceUpdated, payload.TokenPriceUpdatedData{
			NewPriceToman: res.NewPriceToman,
			OldPriceToman: res.OldPriceToman,
			UpdatedAt:     res.UpdatedAt,
		})
	}

	return res, nil
}
