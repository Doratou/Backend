package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type SettlementService interface {
	Create(ctx context.Context, userID, groupID int64, req dto.CreateSettlementRequest) (*dto.SettlementResponse, error)
	List(ctx context.Context, userID, groupID int64, req dto.ListSettlementsRequest) (*dto.SettlementListResponse, error)
	Get(ctx context.Context, userID, groupID, id int64) (*dto.SettlementResponse, error)
	Confirm(ctx context.Context, userID, groupID, id int64) (*dto.SettlementResponse, error)
	Reject(ctx context.Context, userID, groupID, id int64, req dto.RejectSettlementRequest) (*dto.SettlementResponse, error)
	Delete(ctx context.Context, userID, groupID, id int64) error
}
