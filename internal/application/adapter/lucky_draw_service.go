package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type LuckyDrawService interface {
	Create(ctx context.Context, userID, groupID int64, req dto.CreateLuckyDrawRequest) (*dto.LuckyDrawResponse, error)
	List(ctx context.Context, userID, groupID int64, req dto.ListLuckyDrawsRequest) (*dto.PaginatedResponse[dto.LuckyDrawListItem], error)
	Get(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error)
	Update(ctx context.Context, userID, groupID, drawID int64, req dto.UpdateLuckyDrawRequest) (*dto.LuckyDrawResponse, error)
	Run(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error)
}
