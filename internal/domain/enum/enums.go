package enum

type UserRole string

const (
	UserRoleUser  UserRole = "USER"
	UserRoleAdmin UserRole = "ADMIN"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
)

type GroupStatus string

const (
	GroupStatusActive   GroupStatus = "ACTIVE"
	GroupStatusInactive GroupStatus = "INACTIVE"
)

type GroupMemberStatus string

const (
	GroupMemberStatusActive   GroupMemberStatus = "ACTIVE"
	GroupMemberStatusInactive GroupMemberStatus = "INACTIVE"
)

type GroupMemberRole string

const (
	GroupMemberRoleOwner  GroupMemberRole = "OWNER"
	GroupMemberRoleMember GroupMemberRole = "MEMBER"
)

type DirectInviteStatus string

const (
	DirectInviteStatusPending  DirectInviteStatus = "PENDING"
	DirectInviteStatusAccepted DirectInviteStatus = "ACCEPTED"
	DirectInviteStatusRejected DirectInviteStatus = "REJECTED"
)

type AttachmentType string

const (
	AttachmentTypeProfileImage      AttachmentType = "PROFILE_IMAGE"
	AttachmentTypeGroupCover        AttachmentType = "GROUP_COVER"
	AttachmentTypeGalleryImage      AttachmentType = "GALLERY_IMAGE"
	AttachmentTypeGalleryVideo      AttachmentType = "GALLERY_VIDEO"
	AttachmentTypeExpenseReceipt    AttachmentType = "EXPENSE_RECEIPT"
	AttachmentTypeSettlementReceipt AttachmentType = "SETTLEMENT_RECEIPT"
)

type OTPType string

const (
	OTPTypeEmail OTPType = "email"
	OTPTypePhone OTPType = "phone"
)

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
	case ReportReasonSpam,
		ReportReasonHarassment,
		ReportReasonInappropriateContent,
		ReportReasonFraud,
		ReportReasonPrivacyViolation,
		ReportReasonOther:
		return true
	default:
		return false
	}
}

type ReportStatus string

const (
	ReportStatusPending     ReportStatus = "PENDING"
	ReportStatusUnderReview ReportStatus = "UNDER_REVIEW"
	ReportStatusResolved    ReportStatus = "RESOLVED"
	ReportStatusRejected    ReportStatus = "REJECTED"
	ReportStatusDismissed   ReportStatus = "DISMISSED"
)

func (s ReportStatus) IsValid() bool {
	switch s {
	case ReportStatusPending,
		ReportStatusUnderReview,
		ReportStatusResolved,
		ReportStatusRejected,
		ReportStatusDismissed:
		return true
	default:
		return false
	}
}

type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "PENDING"
	FriendshipStatusAccepted FriendshipStatus = "ACCEPTED"
	FriendshipStatusRejected FriendshipStatus = "REJECTED"
)

type PollStatus string

const (
	PollStatusActive   PollStatus = "ACTIVE"
	PollStatusInactive PollStatus = "INACTIVE"
)

type PollVisibility string

const (
	PollVisibilityPublic  PollVisibility = "PUBLIC"
	PollVisibilityPrivate PollVisibility = "PRIVATE"
)

type LuckyDrawStatus string

const (
	LuckyDrawStatusDraft     LuckyDrawStatus = "DRAFT"
	LuckyDrawStatusCompleted LuckyDrawStatus = "COMPLETED"
)

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
