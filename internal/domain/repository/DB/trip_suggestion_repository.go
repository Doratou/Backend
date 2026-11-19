package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
)

type TripSuggestionRepository interface {
	Create(ctx context.Context, suggestion *entity.TripSuggestion) error
	ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]entity.TripSuggestion, int64, error)
}
