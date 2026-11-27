package payload

import "time"

// TargetType defines the delivery scope for real-time WebSocket events.
type TargetType string

const (
	// TargetUsers targets a single or multiple connected users by IDs.
	TargetUsers TargetType = "users"

	// TargetGroups targets all connected members of one or multiple groups.
	TargetGroups TargetType = "groups"

	// TargetBroadcast targets all active connected sessions.
	TargetBroadcast TargetType = "broadcast"
)

func (t TargetType) IsValid() bool {
	switch t {
	case TargetUsers, TargetGroups, TargetBroadcast:
		return true
	default:
		return false
	}
}

// WSEvent defines strongly-typed identifiers for all real-time WebSocket events.
type WSEvent string

const (
	// 1. Friend Request Events
	WSEventFriendRequestReceived  WSEvent = "friend_request.received"
	WSEventFriendRequestResponded WSEvent = "friend_request.responded"

	// 2. Settlement Events
	WSEventSettlementPaid WSEvent = "settlement.paid"

	// 3. Poll Events
	WSEventPollCreated WSEvent = "poll.created"
	WSEventPollClosed  WSEvent = "poll.closed"

	// 4. Draw / Lottery Events
	WSEventDrawCreated WSEvent = "draw.created"
	WSEventDrawResult  WSEvent = "draw.result"

	// 5. Checklist Events
	WSEventChecklistUpdated WSEvent = "checklist.updated"

	// 6. Admin Report Events
	WSEventAdminReportAnswered WSEvent = "admin_report.answered"

	// 7. Group Membership Events
	WSEventGroupMemberJoined WSEvent = "group.member_joined"
	WSEventGroupMemberLeft   WSEvent = "group.member_left"

	// 8. Token Price Events
	WSEventTokenPriceUpdated WSEvent = "token.price_updated"

	// 9. Manual / Custom Announcement Events (Owner to Group)
	WSEventGroupAnnouncement WSEvent = "group.announcement"
)

func (e WSEvent) IsValid() bool {
	switch e {
	case WSEventFriendRequestReceived,
		WSEventFriendRequestResponded,
		WSEventSettlementPaid,
		WSEventPollCreated,
		WSEventPollClosed,
		WSEventDrawCreated,
		WSEventDrawResult,
		WSEventChecklistUpdated,
		WSEventAdminReportAnswered,
		WSEventGroupMemberJoined,
		WSEventGroupMemberLeft,
		WSEventTokenPriceUpdated,
		WSEventGroupAnnouncement:
		return true
	default:
		return false
	}
}

// WSPayload represents the message contract published to q.ws.classic.
type WSPayload struct {
	Event      WSEvent    `json:"event"`
	TargetType TargetType `json:"target_type"`
	TargetIDs  []int64    `json:"target_ids,omitempty"`
	Data       any        `json:"data"`
	SenderID   int64      `json:"sender_id,omitempty"`
	Timestamp  time.Time  `json:"timestamp"`
}
