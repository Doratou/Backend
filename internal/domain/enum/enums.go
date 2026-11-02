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

type AttachmentType string

const (
	AttachmentTypeProfileImage   AttachmentType = "PROFILE_IMAGE"
	AttachmentTypeGroupCover     AttachmentType = "GROUP_COVER"
	AttachmentTypeGalleryImage   AttachmentType = "GALLERY_IMAGE"
	AttachmentTypeGalleryVideo   AttachmentType = "GALLERY_VIDEO"
	AttachmentTypeExpenseReceipt AttachmentType = "EXPENSE_RECEIPT"
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
