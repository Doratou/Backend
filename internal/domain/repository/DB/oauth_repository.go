package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
)

// OAuthRepository defines the persistence operations for OAuth identity links.
type OAuthRepository interface {
	// FindByProvider retrieves an OAuth account by provider name and provider_id (e.g. Google sub).
	FindByProvider(ctx context.Context, provider, providerID string) (*entity.UserOAuthAccount, error)

	// Create persists a new OAuth account record.
	Create(ctx context.Context, account *entity.UserOAuthAccount) error

	// ExistsByUserAndProvider checks if a user is already linked with the given OAuth provider.
	ExistsByUserAndProvider(ctx context.Context, userID int64, provider string) (bool, error)
}
