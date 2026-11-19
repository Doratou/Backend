package entity

// UserOAuthAccount stores an OAuth identity linked to a user account.
// A single user can have multiple rows here (one per provider).
type UserOAuthAccount struct {
	BaseEntity
	UserID     int64  `gorm:"column:user_id;not null"`
	Provider   string `gorm:"column:provider;size:32;not null"` // e.g. "google"
	ProviderID string `gorm:"column:provider_id;not null"`      // Provider-unique subject ID (e.g. Google "sub")
	User       *User  `gorm:"foreignKey:UserID"`
}

func (UserOAuthAccount) TableName() string {
	return "user_oauth_accounts"
}
