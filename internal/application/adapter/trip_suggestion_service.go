package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type TripSuggestionService interface {
	SuggestItems(ctx context.Context, req dto.TripSuggestionRequest) (*dto.TripSuggestionResponse, error)
}