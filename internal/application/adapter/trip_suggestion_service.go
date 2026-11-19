package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type TripSuggestionService interface {
	SuggestItems(ctx context.Context, groupID, userID int64, req dto.TripSuggestionRequest) (*dto.TripSuggestionResponse, error)
	ListByGroup(ctx context.Context, groupID int64, page, limit, offset int) (*dto.TripSuggestionHistoryResponse, error)
}
