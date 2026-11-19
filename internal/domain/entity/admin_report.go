package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

// AdminReport represents a moderation report submitted by a user against another user or group.
type AdminReport struct {
	BaseEntity
	ReporterUserID   *int64            `gorm:"column:reporter_user_id" json:"reporter_user_id,omitempty"`
	ReporterUser     *User             `gorm:"foreignKey:ReporterUserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"reporter_user,omitempty"`
	ReportedUserID   *int64            `gorm:"column:reported_user_id" json:"reported_user_id,omitempty"`
	ReportedUser     *User             `gorm:"foreignKey:ReportedUserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"reported_user,omitempty"`
	ReportedGroupID  *int64            `gorm:"column:reported_group_id" json:"reported_group_id,omitempty"`
	Reason           enum.ReportReason `gorm:"column:reason;type:report_reason;not null" json:"reason"`
	Description      string            `gorm:"column:description" json:"description"`
	Status           enum.ReportStatus `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	ReviewedByUserID *int64            `gorm:"column:reviewed_by_user_id" json:"reviewed_by_user_id,omitempty"`
	ReviewedByUser   *User             `gorm:"foreignKey:ReviewedByUserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"reviewed_by_user,omitempty"`
	Response         *string           `gorm:"column:response" json:"response,omitempty"`
	ResolvedAt       *time.Time        `gorm:"column:resolved_at" json:"resolved_at,omitempty"`
}

func (AdminReport) TableName() string {
	return "admin_reports"
}
