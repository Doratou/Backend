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
