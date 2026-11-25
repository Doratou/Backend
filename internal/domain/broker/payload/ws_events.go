package payload

import "time"

// ============================================================================
// 1. Friend Request Payloads
// ============================================================================

// FriendRequestReceivedData is the data contract for WSEventFriendRequestReceived.
// Dispatched to target user when a new friend request arrives.
type FriendRequestReceivedData struct {
	RequestID      int64     `json:"request_id"`
	SenderID       string    `json:"sender_id"`
	SenderName     string    `json:"sender_name"`
	SenderAvatarID int64     `json:"sender_avatar_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// FriendRequestRespondedData is the data contract for WSEventFriendRequestAccepted and WSEventFriendRequestRejected.
// Dispatched to initial sender when the target responds.
type FriendRequestRespondedData struct {
	RequestID         int64     `json:"request_id"`
	ResponderID       string    `json:"responder_id"`
	ResponderAvatarID int64     `json:"responder_avatar_id"`
	ResponderName     string    `json:"responder_name"`
	Status            string    `json:"status"` // "accepted" or "rejected"
	RespondedAt       time.Time `json:"responded_at"`
}

// ============================================================================
// 2. Settlement Payloads
// ============================================================================

// SettlementPaidData is the data contract for WSEventSettlementPaid.
// Dispatched to creditor when a debtor settles their balance.
type SettlementPaidData struct {
	SettlementID int64     `json:"settlement_id"`
	GroupID      int64     `json:"group_id"`
	GroupName    string    `json:"group_name"`
	DebtorID     string    `json:"debtor_id"`
	DebtorName   string    `json:"debtor_name"`
	CreditorID   string    `json:"creditor_id"`
	CreditorName string    `json:"creditor_name"`
	AmountToman  int64     `json:"amount_toman"`
	SettledAt    time.Time `json:"settled_at"`
}

// ============================================================================
// 3. Poll Payloads
// ============================================================================

// PollCreatedData is the data contract for WSEventPollCreated.
// Dispatched to all group members when a new poll is opened.
type PollCreatedData struct {
	PollID      int64      `json:"poll_id"`
	GroupID     int64      `json:"group_id"`
	GroupName   string     `json:"group_name"`
	CreatorID   string     `json:"creator_id"`
	CreatorName string     `json:"creator_name"`
	Question    string     `json:"question"`
	Options     []string   `json:"options"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// PollOptionResult represents vote aggregation for an individual poll option.
type PollOptionResult struct {
	OptionID   int64   `json:"option_id"`
	Text       string  `json:"text"`
	VoteCount  int     `json:"vote_count"`
	Percentage float64 `json:"percentage"`
}

// PollClosedData is the data contract for WSEventPollClosed.
// Dispatched to all group members when a poll concludes with final tally.
type PollClosedData struct {
	PollID         int64              `json:"poll_id"`
	GroupID        int64              `json:"group_id"`
	GroupName      string             `json:"group_name"`
	Question       string             `json:"question"`
	TotalVotes     int                `json:"total_votes"`
	Results        []PollOptionResult `json:"results"`
	WinningOptions []PollOptionResult `json:"winning_option"`
	ClosedAt       time.Time          `json:"closed_at"`
}

// ============================================================================
// 4. Draw / Lottery Payloads
// ============================================================================

// DrawCreatedData is the data contract for WSEventDrawCreated.
// Dispatched to all group members when a new lottery/draw is established.
type DrawCreatedData struct {
	DrawID      int64     `json:"draw_id"`
	GroupID     int64     `json:"group_id"`
	GroupName   string    `json:"group_name"`
	CreatorID   string    `json:"creator_id"`
	CreatorName string    `json:"creator_name"`
	Title       string    `json:"title"`
	Candidates  []string  `json:"candidates"`
	CreatedAt   time.Time `json:"created_at"`
}

// DrawResultData is the data contract for WSEventDrawResult.
// Dispatched to all group members when winner(s) are selected.
type DrawResultData struct {
	DrawID     int64     `json:"draw_id"`
	GroupID    int64     `json:"group_id"`
	GroupName  string    `json:"group_name"`
	Title      string    `json:"title"`
	WinnerID   string    `json:"winner_id"`
	WinnerName string    `json:"winner_name"`
	DrawnAt    time.Time `json:"drawn_at"`
}

// ============================================================================
// 5. Checklist Payloads
// ============================================================================

// ChecklistUpdatedData is the data contract for WSEventChecklistUpdated.
// Dispatched to group members when a checklist item is added, toggled, or modified.
type ChecklistUpdatedData struct {
	GroupID       int64     `json:"group_id"`
	GroupName     string    `json:"group_name"`
	UpdatedByID   string    `json:"updated_by_id"`
	UpdatedByName string    `json:"updated_by_name"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ============================================================================
// 6. Admin Report Payloads
// ============================================================================

// AdminReportAnsweredData is the data contract for WSEventAdminReportAnswered.
// Dispatched to reporting user when support/admin answers their report.
type AdminReportAnsweredData struct {
	ReportID   int64     `json:"report_id"`
	UserID     string    `json:"user_id"`
	Subject    string    `json:"subject"`
	Answer     string    `json:"answer"`
	AdminName  string    `json:"admin_name,omitempty"`
	AnsweredAt time.Time `json:"answered_at"`
}

// ============================================================================
// 7. Group Membership Payloads
// ============================================================================

// GroupMemberJoinedData is the data contract for WSEventGroupMemberJoined.
// Dispatched to group members when a new member joins.
type GroupMemberJoinedData struct {
	GroupID   int64     `json:"group_id"`
	GroupName string    `json:"group_name"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	JoinedAt  time.Time `json:"joined_at"`
}

// GroupMemberLeftData is the data contract for WSEventGroupMemberLeft.
// Dispatched to group members when an existing member leaves or is removed.
type GroupMemberLeftData struct {
	GroupID   int64     `json:"group_id"`
	GroupName string    `json:"group_name"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Reason    string    `json:"reason,omitempty"` // "left" or "kicked"
	LeftAt    time.Time `json:"left_at"`
}

// GroupInviteReceivedData is the data contract for WSEventGroupInviteReceived.
// Dispatched to an invited user when they receive a direct group invitation.
type GroupInviteReceivedData struct {
	InviteID       int64     `json:"invite_id"`
	GroupID        int64     `json:"group_id"`
	GroupName      string    `json:"group_name"`
	SenderID       string    `json:"sender_id"`
	SenderName     string    `json:"sender_name"`
	SenderAvatarID int64     `json:"sender_avatar_id,omitempty"`
	InvitedAt      time.Time `json:"invited_at"`
}

// ============================================================================
// 8. Token Price Payloads
// ============================================================================

// TokenPriceUpdatedData is the data contract for WSEventTokenPriceUpdated.
// Broadcasted to all connected sessions when market pricing fluctuates.
type TokenPriceUpdatedData struct {
	NewPriceToman int64     `json:"new_price_toman"`
	OldPriceToman int64     `json:"old_price_toman"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ============================================================================
// 9. Manual / Custom Announcement Payloads
// ============================================================================

// GroupAnnouncementData is the data contract for WSEventGroupAnnouncement.
// Dispatched to all group members when group owner broadcasts an arbitrary announcement.
type GroupAnnouncementData struct {
	GroupID    int64     `json:"group_id"`
	GroupName  string    `json:"group_name"`
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	SentAt     time.Time `json:"sent_at"`
}
