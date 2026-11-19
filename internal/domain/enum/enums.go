package enum

// UserRole enum values: "USER", "ADMIN".
type UserRole string

const (
	UserRoleUser  UserRole = "USER"
	UserRoleAdmin UserRole = "ADMIN"
)

func (u UserRole) IsValid() bool {
	switch u {
	case UserRoleUser, UserRoleAdmin:
		return true
	default:
		return false
	}
}

// UserStatus enum values: "ACTIVE", "INACTIVE".
type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
)

func (u UserStatus) IsValid() bool {
	switch u {
	case UserStatusActive, UserStatusInactive:
		return true
	default:
		return false
	}
}

type GroupStatus string

const (
	GroupStatusActive   GroupStatus = "ACTIVE"
	GroupStatusSettling GroupStatus = "SETTLING"
	GroupStatusClosed   GroupStatus = "CLOSED"
)

func (g GroupStatus) IsValid() bool {
	switch g {
	case GroupStatusActive, GroupStatusSettling, GroupStatusClosed:
		return true
	default:
		return false
	}
}

// GroupMemberStatus enum values: "ACTIVE", "INACTIVE".
type GroupMemberStatus string

const (
	GroupMemberStatusActive   GroupMemberStatus = "ACTIVE"
	GroupMemberStatusInactive GroupMemberStatus = "INACTIVE"
)

func (g GroupMemberStatus) IsValid() bool {
	switch g {
	case GroupMemberStatusActive, GroupMemberStatusInactive:
		return true
	default:
		return false
	}
}

// GroupMemberRole enum values: "OWNER", "MEMBER".
type GroupMemberRole string

const (
	GroupMemberRoleOwner  GroupMemberRole = "OWNER"
	GroupMemberRoleMember GroupMemberRole = "MEMBER"
)

func (g GroupMemberRole) IsValid() bool {
	switch g {
	case GroupMemberRoleOwner, GroupMemberRoleMember:
		return true
	default:
		return false
	}
}

// DirectInviteStatus enum values: "PENDING", "ACCEPTED", "REJECTED".
type DirectInviteStatus string

const (
	DirectInviteStatusPending  DirectInviteStatus = "PENDING"
	DirectInviteStatusAccepted DirectInviteStatus = "ACCEPTED"
	DirectInviteStatusRejected DirectInviteStatus = "REJECTED"
)

func (d DirectInviteStatus) IsValid() bool {
	switch d {
	case DirectInviteStatusPending, DirectInviteStatusAccepted, DirectInviteStatusRejected:
		return true
	default:
		return false
	}
}

// AttachmentType enum values: "PROFILE_IMAGE", "GROUP_COVER", "GALLERY_IMAGE", "GALLERY_VIDEO", "EXPENSE_RECEIPT", "SETTLEMENT_RECEIPT".
type AttachmentType string

const (
	AttachmentTypeProfileImage      AttachmentType = "PROFILE_IMAGE"
	AttachmentTypeGroupCover        AttachmentType = "GROUP_COVER"
	AttachmentTypeGalleryImage      AttachmentType = "GALLERY_IMAGE"
	AttachmentTypeGalleryVideo      AttachmentType = "GALLERY_VIDEO"
	AttachmentTypeExpenseReceipt    AttachmentType = "EXPENSE_RECEIPT"
	AttachmentTypeSettlementReceipt AttachmentType = "SETTLEMENT_RECEIPT"
)

func (a AttachmentType) IsValid() bool {
	switch a {
	case AttachmentTypeProfileImage, AttachmentTypeGroupCover, AttachmentTypeGalleryImage, AttachmentTypeGalleryVideo, AttachmentTypeExpenseReceipt, AttachmentTypeSettlementReceipt:
		return true
	default:
		return false
	}
}

// OTPType enum values: "email", "phone".
type OTPType string

const (
	OTPTypeEmail OTPType = "email"
	OTPTypePhone OTPType = "phone"
)

func (o OTPType) IsValid() bool {
	switch o {
	case OTPTypeEmail, OTPTypePhone:
		return true
	default:
		return false
	}
}

// ReportReason enum values: "SPAM", "HARASSMENT", "INAPPROPRIATE_CONTENT", "FRAUD", "PRIVACY_VIOLATION", "OTHER".
type ReportReason string

const (
	ReportReasonSpam                 ReportReason = "SPAM"
	ReportReasonHarassment           ReportReason = "HARASSMENT"
	ReportReasonInappropriateContent ReportReason = "INAPPROPRIATE_CONTENT"
	ReportReasonFraud                ReportReason = "FRAUD"
	ReportReasonPrivacyViolation     ReportReason = "PRIVACY_VIOLATION"
	ReportReasonOther                ReportReason = "OTHER"
)

func (r ReportReason) IsValid() bool {
	switch r {
	case ReportReasonSpam, ReportReasonHarassment, ReportReasonInappropriateContent, ReportReasonFraud, ReportReasonPrivacyViolation, ReportReasonOther:
		return true
	default:
		return false
	}
}

// ReportStatus enum values: "PENDING", "UNDER_REVIEW", "RESOLVED", "REJECTED", "DISMISSED".
type ReportStatus string

const (
	ReportStatusPending     ReportStatus = "PENDING"
	ReportStatusUnderReview ReportStatus = "UNDER_REVIEW"
	ReportStatusResolved    ReportStatus = "RESOLVED"
	ReportStatusRejected    ReportStatus = "REJECTED"
	ReportStatusDismissed   ReportStatus = "DISMISSED"
)

func (r ReportStatus) IsValid() bool {
	switch r {
	case ReportStatusPending, ReportStatusUnderReview, ReportStatusResolved, ReportStatusRejected, ReportStatusDismissed:
		return true
	default:
		return false
	}
}

// FriendshipStatus enum values: "PENDING", "ACCEPTED", "REJECTED".
type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "PENDING"
	FriendshipStatusAccepted FriendshipStatus = "ACCEPTED"
	FriendshipStatusRejected FriendshipStatus = "REJECTED"
)

func (f FriendshipStatus) IsValid() bool {
	switch f {
	case FriendshipStatusPending, FriendshipStatusAccepted, FriendshipStatusRejected:
		return true
	default:
		return false
	}
}

// PollStatus enum values: "ACTIVE", "INACTIVE".
type PollStatus string

const (
	PollStatusActive   PollStatus = "ACTIVE"
	PollStatusInactive PollStatus = "INACTIVE"
)

func (p PollStatus) IsValid() bool {
	switch p {
	case PollStatusActive, PollStatusInactive:
		return true
	default:
		return false
	}
}

// PollVisibility enum values: "PUBLIC", "PRIVATE".
type PollVisibility string

const (
	PollVisibilityPublic  PollVisibility = "PUBLIC"
	PollVisibilityPrivate PollVisibility = "PRIVATE"
)

func (p PollVisibility) IsValid() bool {
	switch p {
	case PollVisibilityPublic, PollVisibilityPrivate:
		return true
	default:
		return false
	}
}

type LuckyDrawType string

const (
	LuckyDrawTypeEqual    LuckyDrawType = "EQUAL"
	LuckyDrawTypeWeighted LuckyDrawType = "WEIGHTED"
)

type LuckyDrawCandidateType string

const (
	LuckyDrawCandidateTypeText LuckyDrawCandidateType = "TEXT"
	LuckyDrawCandidateTypeUser LuckyDrawCandidateType = "USER"
)

type ExpenseType string

const (
	ExpenseTypeNormal  ExpenseType = "NORMAL"
	ExpenseTypePremium ExpenseType = "PREMIUM"
)

func (e ExpenseType) IsValid() bool {
	switch e {
	case ExpenseTypeNormal, ExpenseTypePremium:
		return true
	default:
		return false
	}
}

type ExpenseSplitType string

const (
	ExpenseSplitTypeEqual      ExpenseSplitType = "EQUAL"
	ExpenseSplitTypeExact      ExpenseSplitType = "EXACT"
	ExpenseSplitTypePercentage ExpenseSplitType = "PERCENTAGE"
	ExpenseSplitTypeWeighted   ExpenseSplitType = "WEIGHTED"
)

func (e ExpenseSplitType) IsValid() bool {
	switch e {
	case ExpenseSplitTypeEqual, ExpenseSplitTypeExact, ExpenseSplitTypePercentage, ExpenseSplitTypeWeighted:
		return true
	default:
		return false
	}
}

type ExpenseListScope string

const (
	ExpenseListScopeAll         ExpenseListScope = "all"
	ExpenseListScopeCreatedByMe ExpenseListScope = "created_by_me"
	ExpenseListScopeOwedByMe    ExpenseListScope = "owed_by_me"
)

func (e ExpenseListScope) IsValid() bool {
	return e == ExpenseListScopeAll || e == ExpenseListScopeCreatedByMe || e == ExpenseListScopeOwedByMe
}

// LuckyDrawStatus enum values: "DRAFT", "COMPLETED".
type LuckyDrawStatus string

const (
	LuckyDrawStatusDraft     LuckyDrawStatus = "DRAFT"
	LuckyDrawStatusCompleted LuckyDrawStatus = "COMPLETED"
)

func (l LuckyDrawStatus) IsValid() bool {
	switch l {
	case LuckyDrawStatusDraft, LuckyDrawStatusCompleted:
		return true
	default:
		return false
	}
}

type SettlementStatus string

const (
	SettlementStatusPending  SettlementStatus = "PENDING"
	SettlementStatusSuccess  SettlementStatus = "SUCCESS"
	SettlementStatusReject   SettlementStatus = "REJECT"
	SettlementStatusCanceled SettlementStatus = "CANCELED"
)

func (s SettlementStatus) IsValid() bool {
	switch s {
	case SettlementStatusPending, SettlementStatusSuccess, SettlementStatusReject, SettlementStatusCanceled:
		return true
	default:
		return false
	}
}

type SettlementDirection string

const (
	SettlementDirectionAll      SettlementDirection = "all"
	SettlementDirectionIncoming SettlementDirection = "incoming"
	SettlementDirectionOutgoing SettlementDirection = "outgoing"
)

func (s SettlementDirection) IsValid() bool {
	return s == SettlementDirectionAll || s == SettlementDirectionIncoming || s == SettlementDirectionOutgoing
}
