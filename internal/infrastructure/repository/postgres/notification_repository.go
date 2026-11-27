package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db database.Database
}

// NewNotificationRepository creates a new postgres-backed NotificationRepository.
func NewNotificationRepository(db database.Database) domain_repo.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *notificationRepository) Create(ctx context.Context, notif *entity.Notification) error {
	if notif == nil {
		return exception.ErrBadRequest
	}
	if err := r.getDB(ctx).Create(notif).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *notificationRepository) CreateBatch(ctx context.Context, notifs []*entity.Notification) error {
	if len(notifs) == 0 {
		return nil
	}
	if err := r.getDB(ctx).Create(&notifs).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *notificationRepository) FindByID(ctx context.Context, id int64) (*entity.Notification, error) {
	var notif entity.Notification
	err := r.getDB(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&notif).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrNotificationNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &notif, nil
}

type notificationQueryResult struct {
	ID          int64     `gorm:"column:id"`
	RecipientID int64     `gorm:"column:recipient_id"`
	ActorID     *int64    `gorm:"column:actor_id"`
	GroupID     *int64    `gorm:"column:group_id"`
	Event       string    `gorm:"column:event"`
	Data        []byte    `gorm:"column:data"`
	LastError   *string   `gorm:"column:last_error"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	IsRead      bool      `gorm:"column:is_read"`
}

func (r *notificationRepository) ListByRecipient(ctx context.Context, filter domain_repo.NotificationFilter) ([]*entity.Notification, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	query := r.getDB(ctx).
		Table("notifications n").
		Select(`
			n.id,
			n.recipient_id,
			n.actor_id,
			n.group_id,
			n.event,
			n.data,
			n.last_error,
			n.created_at,
			n.updated_at,
			(n.id <= COALESCE(r.last_read_id, 0)) AS is_read
		`).
		Joins("LEFT JOIN notification_reads r ON r.user_id = n.recipient_id").
		Where("n.recipient_id = ? AND n.deleted_at IS NULL", filter.RecipientID)

	if filter.CursorID > 0 {
		query = query.Where("n.id < ?", filter.CursorID)
	}

	var rows []notificationQueryResult
	if err := query.Order("n.id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}

	results := make([]*entity.Notification, len(rows))
	for i, row := range rows {
		notif := &entity.Notification{
			RecipientID: row.RecipientID,
			ActorID:     row.ActorID,
			GroupID:     row.GroupID,
			Event:       payload.WSEvent(row.Event),
			Data:        row.Data,
			LastError:   row.LastError,
			IsRead:      row.IsRead,
		}
		notif.ID = row.ID
		notif.CreatedAt = row.CreatedAt
		notif.UpdatedAt = row.UpdatedAt
		results[i] = notif
	}
	return results, nil
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.getDB(ctx).
		Table("notifications n").
		Joins("LEFT JOIN notification_reads r ON r.user_id = n.recipient_id").
		Where("n.recipient_id = ? AND n.deleted_at IS NULL AND n.id > COALESCE(r.last_read_id, 0)", userID).
		Count(&count).Error
	if err != nil {
		return 0, exception.Wrap(exception.ErrInternal, err)
	}
	return count, nil
}

func (r *notificationRepository) GetLastReadID(ctx context.Context, userID int64) (int64, error) {
	var state entity.NotificationReadState
	err := r.getDB(ctx).Where("user_id = ?", userID).First(&state).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, exception.Wrap(exception.ErrInternal, err)
	}
	return state.LastReadID, nil
}

func (r *notificationRepository) Update(ctx context.Context, notif *entity.Notification) error {
	if notif == nil || notif.ID == 0 {
		return exception.ErrBadRequest
	}
	res := r.getDB(ctx).Model(&entity.Notification{}).
		Where("id = ? AND deleted_at IS NULL", notif.ID).
		Updates(map[string]any{
			"data":       notif.Data,
			"last_error": notif.LastError,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotificationNotFound
	}
	return nil
}

func (r *notificationRepository) UpdateLastError(ctx context.Context, id int64, lastErr *string) error {
	res := r.getDB(ctx).Model(&entity.Notification{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"last_error": lastErr,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotificationNotFound
	}
	return nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, userID int64, upToID int64) error {
	query := `
		INSERT INTO notification_reads (user_id, last_read_id, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id)
		DO UPDATE SET
			last_read_id = GREATEST(notification_reads.last_read_id, EXCLUDED.last_read_id),
			updated_at = CURRENT_TIMESTAMP;
	`
	if err := r.getDB(ctx).Exec(query, userID, upToID).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *notificationRepository) SoftDelete(ctx context.Context, id int64) error {
	res := r.getDB(ctx).Where("id = ?", id).Delete(&entity.Notification{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotificationNotFound
	}
	return nil
}

func (r *notificationRepository) SoftDeleteAllByRecipient(ctx context.Context, recipientID int64) error {
	err := r.getDB(ctx).
		Where("recipient_id = ? AND deleted_at IS NULL", recipientID).
		Delete(&entity.Notification{}).Error
	if err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *notificationRepository) HardDelete(ctx context.Context, id int64) error {
	res := r.getDB(ctx).Unscoped().Where("id = ?", id).Delete(&entity.Notification{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrNotificationNotFound
	}
	return nil
}

func (r *notificationRepository) PurgeSoftDeletedBefore(ctx context.Context, before time.Time) (int64, error) {
	res := r.getDB(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", before).
		Delete(&entity.Notification{})
	if res.Error != nil {
		return 0, exception.Wrap(exception.ErrInternal, res.Error)
	}
	return res.RowsAffected, nil
}
