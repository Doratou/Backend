package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

// UserRepository defines a fully-featured, comprehensive CRUD and query interface for User entities.
type UserRepository interface {
	// Create persists a new user entity in the database.
	Create(ctx context.Context, user *entity.User) error

	// FindByID retrieves a user by their primary key, including avatar relations.
	FindByID(ctx context.Context, id int64) (*entity.User, error)

	// FindByEmail retrieves a user by their email address.
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindByPhone retrieves a user by their mobile phone number.
	FindByPhone(ctx context.Context, phone string) (*entity.User, error)

	// FindByTelegramChatID retrieves a user by their Telegram chat ID.
	FindByTelegramChatID(ctx context.Context, chatID int64) (*entity.User, error)

	// FindByTelegramUserID retrieves a user by their Telegram user ID.
	FindByTelegramUserID(ctx context.Context, telegramUserID int64) (*entity.User, error)

	// FindAll returns a paginated slice of users and the total count.
	FindAll(ctx context.Context, limit, offset int) ([]*entity.User, int64, error)

	// FindByStatus returns users filtered by user status (ACTIVE, INACTIVE).
	FindByStatus(ctx context.Context, status enum.UserStatus, limit, offset int) ([]*entity.User, error)

	// FindByRole returns users filtered by system role (USER, ADMIN).
	FindByRole(ctx context.Context, role enum.UserRole, limit, offset int) ([]*entity.User, error)

	// Update updates all writable fields of an existing user.
	Update(ctx context.Context, user *entity.User) error

	// UpdatePassword updates only the password_hash column for the given user.
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error

	// UpdateAvatar updates only the avatar_attachment_id column for the given user.
	UpdateAvatar(ctx context.Context, userID int64, avatarID int64) error

	// UpdateVerifiedStatus updates verification flags (is_email_verified, is_phone_verified).
	UpdateVerifiedStatus(ctx context.Context, userID int64, isEmailVerified, isPhoneVerified bool) error

	// UpdateWallet adds delta (positive or negative) to user's wallet atomically.
	UpdateWallet(ctx context.Context, userID int64, delta int64) error

	// DeleteByID performs a soft delete on a user record.
	DeleteByID(ctx context.Context, id int64) error

	// RestoreByID restores a soft-deleted user.
	RestoreByID(ctx context.Context, id int64) error

	// HardDeleteByID permanently deletes a user record from the database.
	HardDeleteByID(ctx context.Context, id int64) error

	// ExistsByID checks if a user exists by ID.
	ExistsByID(ctx context.Context, id int64) (bool, error)

	// ExistsByEmail checks if an email is already registered.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByPhone checks if a phone number is already registered.
	ExistsByPhone(ctx context.Context, phone string) (bool, error)

	// Count returns the total number of non-deleted users.
	Count(ctx context.Context) (int64, error)
}
