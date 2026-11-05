package entity

type PollVote struct {
	BaseEntity
	PollID   int64        `gorm:"column:poll_id;not null;uniqueIndex:uq_poll_votes_poll_member" json:"poll_id"`
	MemberID int64        `gorm:"column:member_id;not null;uniqueIndex:uq_poll_votes_poll_member" json:"-"`
	OptionID int64        `gorm:"column:option_id;not null" json:"option_id"`
	Poll     *Poll        `gorm:"foreignKey:PollID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Member   *GroupMember `gorm:"foreignKey:MemberID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Option   *PollOption  `gorm:"foreignKey:PollID,OptionID;references:PollID,ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

func (PollVote) TableName() string {
	return "poll_votes"
}
