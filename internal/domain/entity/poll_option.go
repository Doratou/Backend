package entity

type PollOption struct {
	BaseEntity
	PollID int64  `gorm:"column:poll_id;not null" json:"poll_id"`
	Poll   *Poll  `gorm:"foreignKey:PollID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Title  string `gorm:"column:title;size:255;not null" json:"title"`
}

func (PollOption) TableName() string {
	return "poll_options"
}
