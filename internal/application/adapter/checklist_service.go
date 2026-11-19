package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type ChecklistService interface {
	Get(ctx context.Context, userID, groupID int64, req dto.GetChecklistRequest) (*dto.ChecklistResponse, error)
	CreateItem(ctx context.Context, userID, groupID int64, req dto.CreateChecklistItemRequest) (*dto.ChecklistItemResponse, error)
	SetClaim(ctx context.Context, userID, groupID, itemID int64, claimed bool) (*dto.ChecklistItemResponse, error)
	DeleteItem(ctx context.Context, userID, groupID, itemID int64) error
}
