package entity

import (
	"encoding/json"
	"time"

	"github.com/Doratou/Backend/internal/domain/broker/payload"
)

// Notification represents a persistent notification record in the database.
// Designed with zero foreign keys for full microservice independence.
type Notification struct {
	BaseEntity
	RecipientID int64           `gorm:"column:recipient_id;not null" json:"recipient_id"`
	ActorID     *int64          `gorm:"column:actor_id" json:"actor_id,omitempty"`
	GroupID     *int64          `gorm:"column:group_id" json:"group_id,omitempty"`
	Event       payload.WSEvent `gorm:"column:event;type:notification_event;not null" json:"event"`
	Data        json.RawMessage `gorm:"column:data;type:jsonb;not null;default:'{}'" json:"data"`
	LastError   *string         `gorm:"column:last_error" json:"last_error,omitempty"`

	// IsRead is a dynamic computed field populated from High-Water Mark (notification_reads).
	IsRead bool `gorm:"-" json:"is_read"`
}

// TableName overrides the default table name for GORM.
func (Notification) TableName() string {
	return "notifications"
}

// SetData marshals any typed data payload into the raw JSON data field.
func (n *Notification) SetData(v any) error {
	if v == nil {
		n.Data = json.RawMessage("{}")
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	n.Data = b
	return nil
}

// UnmarshalData unmarshals the raw JSON data into the destination target.
func (n *Notification) UnmarshalData(target any) error {
	if len(n.Data) == 0 {
		return nil
	}
	return json.Unmarshal(n.Data, target)
}

// NotificationReadState tracks the High-Water Mark cursor for a user.
type NotificationReadState struct {
	UserID     int64     `gorm:"primaryKey;column:user_id" json:"user_id"`
	LastReadID int64     `gorm:"column:last_read_id;not null;default:0" json:"last_read_id"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName overrides the default table name for GORM.
func (NotificationReadState) TableName() string {
	return "notification_reads"
}
