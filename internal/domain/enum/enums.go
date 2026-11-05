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

type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "PENDING"
	FriendshipStatusAccepted FriendshipStatus = "ACCEPTED"
	FriendshipStatusRejected FriendshipStatus = "REJECTED"
)
