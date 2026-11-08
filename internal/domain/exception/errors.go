// Package exception defines the application's domain error vocabulary and uniform AppError envelopes.
package exception

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the standard error container throughout the application.
// Handlers translate AppErrors into JSON error envelopes without leaking internal details.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Wrap returns a new copy of sentinel with underlying err attached.
func Wrap(sentinel *AppError, err error) *AppError {
	if sentinel == nil {
		return nil
	}
	return &AppError{
		Code:       sentinel.Code,
		Message:    sentinel.Message,
		HTTPStatus: sentinel.HTTPStatus,
		Err:        err,
	}
}

// ==========================================
// 1. Raw Domain Sentinel Errors
// ==========================================
var (
	ErrDomainUserNotFound           = errors.New("user not found")
	ErrDomainUserAlreadyExists      = errors.New("user already exists")
	ErrDomainUserEmailAlreadyExists = errors.New("user email already exists")
	ErrDomainUserPhoneAlreadyExists = errors.New("user phone already exists")
	ErrDomainUserInactive           = errors.New("user account is inactive")

	ErrDomainInvalidCredentials  = errors.New("invalid credentials")
	ErrDomainTokenInvalid        = errors.New("token is invalid")
	ErrDomainTokenExpired        = errors.New("token has expired")
	ErrDomainTokenRevoked        = errors.New("token has been revoked")
	ErrDomainTokenNotFound       = errors.New("token session not found or revoked")
	ErrDomainMissingToken        = errors.New("missing authorization token")
	ErrDomainRefreshTokenInvalid = errors.New("invalid or expired refresh token")

	ErrDomainOTPInvalid           = errors.New("otp code invalid or expired")
	ErrDomainOTPNotFound          = errors.New("otp code not found or expired")
	ErrDomainOTPUnsupportedType   = errors.New("unsupported otp type")
	ErrDomainUnverifiedCredential = errors.New("credentials are not verified with otp")
	ErrDomainRateLimitExceeded    = errors.New("rate limit exceeded, please try again later")

	ErrDomainAttachmentNotFound   = errors.New("attachment not found")

	ErrDomainCaptchaNotFound      = errors.New("captcha not found or expired")
	ErrDomainCaptchaInvalid       = errors.New("captcha code is invalid")
	ErrDomainCaptchaRequired      = errors.New("captcha verification is required")

	ErrDomainReportNotFound        = errors.New("report not found")
	ErrDomainReportTargetRequired  = errors.New("reported user id or reported group id must be provided")
	ErrDomainReportSelfForbidden   = errors.New("you cannot report yourself")
	ErrDomainInvalidReportReason   = errors.New("invalid report reason")
	ErrDomainInvalidReportStatus   = errors.New("invalid report status")
	ErrDomainReportAlreadyResolved = errors.New("report has already been resolved")
	ErrDomainForbiddenRole         = errors.New("forbidden: insufficient permissions")

	ErrDomainBadRequest           = errors.New("bad request")
	ErrDomainInternal             = errors.New("internal server error")

	ErrDomainAttachmentNotFound = errors.New("attachment not found")

	ErrDomainGroupNotFound                = errors.New("group not found")
	ErrDomainGroupForbidden               = errors.New("group access is forbidden")
	ErrDomainGroupMemberNotFound          = errors.New("active group member not found")
	ErrDomainGroupOwnerRequired           = errors.New("group owner permission is required")
	ErrDomainGroupOwnerCannotLeave        = errors.New("group owner must transfer ownership before leaving")
	ErrDomainGroupOwnerCannotBeRemoved    = errors.New("group owner cannot be removed")
	ErrDomainGroupTransferTargetNotMember = errors.New("new group owner must be an active member")
	ErrDomainAlreadyGroupMember           = errors.New("user is already an active group member")
	ErrDomainGroupInactive                = errors.New("group is inactive")
	ErrDomainGroupMustBeInactiveToClone   = errors.New("only inactive groups can be cloned")

	ErrDomainInviteNotFound          = errors.New("group invite not found")
	ErrDomainInviteExpired           = errors.New("group invite has expired")
	ErrDomainInviteRevoked           = errors.New("group invite has been deactivated")
	ErrDomainInviteUsesExhausted     = errors.New("group invite has reached its usage limit")
	ErrDomainInvitePasswordInvalid   = errors.New("group invite password is invalid")
	ErrDomainGroupActiveInviteExists = errors.New("group already has an active invitation link")
	ErrDomainDirectInvitePending     = errors.New("a direct group invitation is already pending")
	ErrDomainDirectInviteCooldown    = errors.New("direct group invitation cooldown is active")
	ErrDomainDirectInviteNotPending  = errors.New("direct group invitation is no longer pending")

	ErrDomainCaptchaNotFound = errors.New("captcha not found or expired")
	ErrDomainCaptchaInvalid  = errors.New("captcha code is invalid")
	ErrDomainCaptchaRequired = errors.New("captcha verification is required")

	ErrDomainBadRequest = errors.New("bad request")
	ErrDomainInternal   = errors.New("internal server error")
)

// ==========================================
// 2. Uniform Application Sentinel AppErrors (All messages in English)
// ==========================================
var (
	// User
	ErrUserNotFound           = &AppError{Code: "USER_NOT_FOUND", Message: "user not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainUserNotFound}
	ErrUserAlreadyExists      = &AppError{Code: "USER_ALREADY_EXISTS", Message: "user already exists", HTTPStatus: http.StatusConflict, Err: ErrDomainUserAlreadyExists}
	ErrUserEmailAlreadyExists = &AppError{Code: "USER_EMAIL_ALREADY_EXISTS", Message: "email is already registered", HTTPStatus: http.StatusConflict, Err: ErrDomainUserEmailAlreadyExists}
	ErrUserPhoneAlreadyExists = &AppError{Code: "USER_PHONE_ALREADY_EXISTS", Message: "phone number is already registered", HTTPStatus: http.StatusConflict, Err: ErrDomainUserPhoneAlreadyExists}
	ErrUserInactive           = &AppError{Code: "USER_INACTIVE", Message: "user account is deactivated", HTTPStatus: http.StatusForbidden, Err: ErrDomainUserInactive}

	// Auth & OTP
	ErrInvalidCredentials  = &AppError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials provided", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainInvalidCredentials}
	ErrTokenInvalid        = &AppError{Code: "TOKEN_INVALID", Message: "authorization token is invalid", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenInvalid}
	ErrTokenExpired        = &AppError{Code: "TOKEN_EXPIRED", Message: "authorization token has expired", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenExpired}
	ErrTokenRevoked        = &AppError{Code: "TOKEN_REVOKED", Message: "authorization token has been revoked", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenRevoked}
	ErrTokenNotFound       = &AppError{Code: "TOKEN_NOT_FOUND", Message: "token session not found or revoked", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenNotFound}
	ErrMissingToken        = &AppError{Code: "MISSING_TOKEN", Message: "authorization token is missing", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainMissingToken}
	ErrRefreshTokenInvalid = &AppError{Code: "REFRESH_TOKEN_INVALID", Message: "refresh token is invalid or expired", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainRefreshTokenInvalid}

	ErrOTPInvalid            = &AppError{Code: "OTP_INVALID", Message: "invalid or expired verification code", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainOTPInvalid}
	ErrOTPNotFound           = &AppError{Code: "OTP_NOT_FOUND", Message: "verification code not found or expired", HTTPStatus: http.StatusNotFound, Err: ErrDomainOTPNotFound}
	ErrOTPUnsupportedType    = &AppError{Code: "OTP_UNSUPPORTED_TYPE", Message: "unsupported verification type", HTTPStatus: http.StatusBadRequest, Err: ErrDomainOTPUnsupportedType}
	ErrEmptyEmailPhone       = &AppError{Code: "EMPTY_PHONE_EMAIL", Message: "at least one of email or phone must be provided", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrUnverifiedCredentials = &AppError{Code: "NOT_VERIFIED_CREDENTIALS", Message: "at least one credential (email or phone) must be verified via OTP", HTTPStatus: http.StatusForbidden, Err: ErrDomainUnverifiedCredential}
	ErrRateLimitExceeded     = &AppError{Code: "RATE_LIMIT_EXCEEDED", Message: "too many requests, please try again later", HTTPStatus: http.StatusTooManyRequests, Err: ErrDomainRateLimitExceeded}

	// Attachment
	ErrAttachmentNotFound = &AppError{Code: "ATTACHMENT_NOT_FOUND", Message: "attachment not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainAttachmentNotFound}

	// Captcha
	ErrCaptchaNotFound = &AppError{Code: "CAPTCHA_NOT_FOUND", Message: "captcha not found or expired", HTTPStatus: http.StatusNotFound, Err: ErrDomainCaptchaNotFound}
	ErrCaptchaInvalid  = &AppError{Code: "CAPTCHA_INVALID", Message: "captcha verification failed", HTTPStatus: http.StatusBadRequest, Err: ErrDomainCaptchaInvalid}
	ErrCaptchaRequired = &AppError{Code: "CAPTCHA_REQUIRED", Message: "captcha headers X-Captcha-ID and X-Captcha-Code are required", HTTPStatus: http.StatusBadRequest, Err: ErrDomainCaptchaRequired}

	// Admin Report & Moderation
	ErrReportNotFound            = &AppError{Code: "REPORT_NOT_FOUND", Message: "report not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainReportNotFound}
	ErrReportTargetRequired      = &AppError{Code: "REPORT_TARGET_REQUIRED", Message: "at least one of reported user id or group id is required", HTTPStatus: http.StatusBadRequest, Err: ErrDomainReportTargetRequired}
	ErrReportSelfForbidden       = &AppError{Code: "REPORT_SELF_FORBIDDEN", Message: "you cannot report yourself", HTTPStatus: http.StatusBadRequest, Err: ErrDomainReportSelfForbidden}
	ErrInvalidReportReason       = &AppError{Code: "INVALID_REPORT_REASON", Message: "invalid report reason provided", HTTPStatus: http.StatusBadRequest, Err: ErrDomainInvalidReportReason}
	ErrInvalidReportStatus       = &AppError{Code: "INVALID_REPORT_STATUS", Message: "invalid report status provided", HTTPStatus: http.StatusBadRequest, Err: ErrDomainInvalidReportStatus}
	ErrReportAlreadyResolved     = &AppError{Code: "REPORT_ALREADY_RESOLVED", Message: "report has already been reviewed and resolved", HTTPStatus: http.StatusBadRequest, Err: ErrDomainReportAlreadyResolved}
	ErrForbiddenRole             = &AppError{Code: "FORBIDDEN_ROLE", Message: "you do not have permission to access this resource", HTTPStatus: http.StatusForbidden, Err: ErrDomainForbiddenRole}

	// General
	ErrBadRequest = &AppError{Code: "BAD_REQUEST", Message: "bad request payload or parameters", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInternal   = &AppError{Code: "INTERNAL_SERVER_ERROR", Message: "an internal server error occurred", HTTPStatus: http.StatusInternalServerError, Err: ErrDomainInternal}

	// Group & Group member
	ErrGroupNotFound                = &AppError{Code: "GROUP_NOT_FOUND", Message: "group not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainGroupNotFound}
	ErrGroupForbidden               = &AppError{Code: "GROUP_FORBIDDEN", Message: "you do not have permission to access this group", HTTPStatus: http.StatusForbidden, Err: ErrDomainGroupForbidden}
	ErrGroupMemberNotFound          = &AppError{Code: "GROUP_MEMBER_NOT_FOUND", Message: "active group member not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainGroupMemberNotFound}
	ErrGroupOwnerRequired           = &AppError{Code: "GROUP_OWNER_REQUIRED", Message: "only the group owner can perform this action", HTTPStatus: http.StatusForbidden, Err: ErrDomainGroupOwnerRequired}
	ErrGroupOwnerCannotLeave        = &AppError{Code: "GROUP_OWNER_CANNOT_LEAVE", Message: "transfer group ownership before leaving the group", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupOwnerCannotLeave}
	ErrGroupOwnerCannotBeRemoved    = &AppError{Code: "GROUP_OWNER_CANNOT_BE_REMOVED", Message: "transfer group ownership before removing the owner", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupOwnerCannotBeRemoved}
	ErrGroupTransferTargetNotMember = &AppError{Code: "GROUP_TRANSFER_TARGET_NOT_MEMBER", Message: "the new owner must be an active group member", HTTPStatus: http.StatusBadRequest, Err: ErrDomainGroupTransferTargetNotMember}
	ErrGroupInvalidPayload          = &AppError{Code: "GROUP_INVALID_PAYLOAD", Message: "group data is invalid: title is required and must be 1–150 characters, and description must be at most 5000 characters", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidUpdatePayload    = &AppError{Code: "GROUP_INVALID_UPDATE_PAYLOAD", Message: "group update is invalid: provide a title and/or description; title must be at most 150 characters and description at most 5000 characters", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupNoFieldsToUpdate        = &AppError{Code: "GROUP_NO_FIELDS_TO_UPDATE", Message: "include title or description in the request body", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidID               = &AppError{Code: "GROUP_INVALID_ID", Message: "groupID must be a positive integer", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidMemberID         = &AppError{Code: "GROUP_INVALID_MEMBER_ID", Message: "memberID must be a positive integer", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidPagination       = &AppError{Code: "GROUP_INVALID_PAGINATION", Message: "page must be a positive integer and limit must be an integer from 1 to 100", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidStatus           = &AppError{Code: "GROUP_INVALID_STATUS", Message: "status must be ACTIVE or INACTIVE", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidTransferOwner    = &AppError{Code: "GROUP_INVALID_TRANSFER_OWNER", Message: "new_owner_id must be a positive integer", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrAlreadyGroupMember           = &AppError{Code: "ALREADY_GROUP_MEMBER", Message: "you are already an active member of this group", HTTPStatus: http.StatusConflict, Err: ErrDomainAlreadyGroupMember}
	ErrGroupInactive                = &AppError{Code: "GROUP_INACTIVE", Message: "group is inactive and cannot accept new members", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupInactive}
	ErrGroupMustBeInactiveToClone   = &AppError{Code: "GROUP_MUST_BE_INACTIVE_TO_CLONE", Message: "only inactive groups can be cloned", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupMustBeInactiveToClone}

	// Invite
	ErrInviteNotFound              = &AppError{Code: "INVITE_NOT_FOUND", Message: "group invite not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainInviteNotFound}
	ErrInviteExpired               = &AppError{Code: "INVITE_EXPIRED", Message: "group invite has expired", HTTPStatus: http.StatusGone, Err: ErrDomainInviteExpired}
	ErrInviteRevoked               = &AppError{Code: "INVITE_DEACTIVATED", Message: "group invite has been deactivated", HTTPStatus: http.StatusGone, Err: ErrDomainInviteRevoked}
	ErrInviteUsesExhausted         = &AppError{Code: "INVITE_USES_EXHAUSTED", Message: "group invite has reached its usage limit", HTTPStatus: http.StatusGone, Err: ErrDomainInviteUsesExhausted}
	ErrInvitePasswordInvalid       = &AppError{Code: "INVITE_PASSWORD_INVALID", Message: "invite password is incorrect", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainInvitePasswordInvalid}
	ErrGroupActiveInviteExists     = &AppError{Code: "GROUP_ACTIVE_INVITE_EXISTS", Message: "this group already has an active invite link; deactivate it or wait until it expires or reaches its usage limit before creating another", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupActiveInviteExists}
	ErrInviteInvalidPayload        = &AppError{Code: "INVITE_INVALID_PAYLOAD", Message: "invite data is invalid: provide a password of 8–72 bytes, a future expires_at timestamp, and a positive max_uses value when supplied", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInviteInvalidID             = &AppError{Code: "INVITE_INVALID_ID", Message: "inviteID must be a positive integer", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInviteInvalidExpiration     = &AppError{Code: "INVITE_INVALID_EXPIRATION", Message: "expires_at must be a future timestamp", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInviteInvalidMaxUses        = &AppError{Code: "INVITE_INVALID_MAX_USES", Message: "max_uses must be positive and cannot be lower than the invite's current usage count", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInviteNoFieldsToUpdate      = &AppError{Code: "INVITE_NO_FIELDS_TO_UPDATE", Message: "include expires_at or max_uses in the request body", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInviteInvalidPasswordFormat = &AppError{Code: "INVITE_INVALID_PASSWORD_FORMAT", Message: "invite password must be between 8 and 72 bytes", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInviteJoinPayload           = &AppError{Code: "INVITE_JOIN_PAYLOAD_INVALID", Message: "join request must include the invite password", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrDirectInvitePending         = &AppError{Code: "DIRECT_INVITE_PENDING", Message: "a direct group invitation is already pending", HTTPStatus: http.StatusConflict, Err: ErrDomainDirectInvitePending}
	ErrDirectInviteCooldown        = &AppError{Code: "DIRECT_INVITE_COOLDOWN", Message: "you cannot invite this friend again yet", HTTPStatus: http.StatusTooManyRequests, Err: ErrDomainDirectInviteCooldown}
	ErrDirectInviteNotPending      = &AppError{Code: "DIRECT_INVITE_NOT_PENDING", Message: "this invitation has already been answered", HTTPStatus: http.StatusConflict, Err: ErrDomainDirectInviteNotPending}
	ErrDirectInviteForbidden       = &AppError{Code: "DIRECT_INVITE_FORBIDDEN", Message: "only the group owner may send direct invitations", HTTPStatus: http.StatusForbidden, Err: ErrDomainGroupOwnerRequired}
	ErrDirectInviteNotFound        = &AppError{Code: "DIRECT_INVITE_NOT_FOUND", Message: "direct group invitation not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainInviteNotFound}
	ErrDirectInviteNotFriend       = &AppError{Code: "DIRECT_INVITE_NOT_FRIEND", Message: "recipient must be your friend", HTTPStatus: http.StatusForbidden, Err: ErrDomainFriendshipForbidden}
	ErrDirectInviteInvalidPayload  = &AppError{Code: "DIRECT_INVITE_INVALID_PAYLOAD", Message: "group_id and one or more positive user_ids are required", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrDirectInviteInvalidResponse = &AppError{Code: "DIRECT_INVITE_INVALID_RESPONSE", Message: "response must be ACCEPTED or REJECTED", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
)

var (
	ErrDomainFriendshipNotFound         = errors.New("friendship not found")
	ErrDomainFriendshipAlreadyExists    = errors.New("a friendship already exists between these users")
	ErrDomainFriendRequestAlreadyExists = errors.New("a pending friend request already exists")
	ErrDomainAlreadyFriends             = errors.New("users are already friends")
	ErrDomainSelfFriendRequest          = errors.New("cannot send a friend request to yourself")
	ErrDomainFriendshipForbidden        = errors.New("you are not allowed to perform this friendship operation")
	ErrDomainFriendshipInvalidStatus    = errors.New("friendship status does not allow this operation")
	ErrDomainFriendRequestCooldown      = errors.New("friend request cooldown is active")
)

var (
	ErrFriendshipNotFound         = &AppError{Code: "FRIENDSHIP_NOT_FOUND", Message: "friendship not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainFriendshipNotFound}
	ErrFriendshipAlreadyExists    = &AppError{Code: "FRIENDSHIP_ALREADY_EXISTS", Message: "a friendship already exists between these users", HTTPStatus: http.StatusConflict, Err: ErrDomainFriendshipAlreadyExists}
	ErrFriendRequestAlreadyExists = &AppError{Code: "FRIEND_REQUEST_ALREADY_EXISTS", Message: "a pending friend request already exists", HTTPStatus: http.StatusConflict, Err: ErrDomainFriendRequestAlreadyExists}
	ErrAlreadyFriends             = &AppError{Code: "ALREADY_FRIENDS", Message: "users are already friends", HTTPStatus: http.StatusConflict, Err: ErrDomainAlreadyFriends}
	ErrSelfFriendRequest          = &AppError{Code: "SELF_FRIEND_REQUEST", Message: "cannot send a friend request to yourself", HTTPStatus: http.StatusBadRequest, Err: ErrDomainSelfFriendRequest}
	ErrFriendshipForbidden        = &AppError{Code: "FRIENDSHIP_FORBIDDEN", Message: "you are not allowed to perform this friendship operation", HTTPStatus: http.StatusForbidden, Err: ErrDomainFriendshipForbidden}
	ErrFriendshipInvalidStatus    = &AppError{Code: "FRIENDSHIP_INVALID_STATUS", Message: "friendship status does not allow this operation", HTTPStatus: http.StatusConflict, Err: ErrDomainFriendshipInvalidStatus}
	ErrFriendRequestCooldown      = &AppError{Code: "FRIEND_REQUEST_COOLDOWN", Message: "please wait before sending another friend request to this user", HTTPStatus: http.StatusTooManyRequests, Err: ErrDomainFriendRequestCooldown}
)

var (
	ErrDomainPollNotFound         = errors.New("poll not found")
	ErrDomainPollOptionNotFound   = errors.New("poll option not found")
	ErrDomainPollVoteNotFound     = errors.New("poll vote not found")
	ErrDomainPollForbidden        = errors.New("poll access is forbidden")
	ErrDomainPollCreatorRequired  = errors.New("poll creator permission is required")
	ErrDomainPollClosed           = errors.New("poll is closed")
	ErrDomainPollPrivate          = errors.New("poll voter identities are private")
	ErrDomainPollGroupInactive    = errors.New("poll changes are frozen while the group is inactive")
	ErrDomainPollInvalidOptions   = errors.New("poll options are invalid")
	ErrDomainPollDuplicateOptions = errors.New("poll options must be unique")
)

var (
	ErrPollNotFound         = &AppError{Code: "POLL_NOT_FOUND", Message: "poll not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainPollNotFound}
	ErrPollOptionNotFound   = &AppError{Code: "POLL_OPTION_NOT_FOUND", Message: "option not found in this poll", HTTPStatus: http.StatusNotFound, Err: ErrDomainPollOptionNotFound}
	ErrPollVoteNotFound     = &AppError{Code: "POLL_VOTE_NOT_FOUND", Message: "no active vote was found for this member in this poll", HTTPStatus: http.StatusNotFound, Err: ErrDomainPollVoteNotFound}
	ErrPollForbidden        = &AppError{Code: "POLL_FORBIDDEN", Message: "only active group members can access this poll", HTTPStatus: http.StatusForbidden, Err: ErrDomainPollForbidden}
	ErrPollCreatorRequired  = &AppError{Code: "POLL_CREATOR_REQUIRED", Message: "only the poll creator can perform this action", HTTPStatus: http.StatusForbidden, Err: ErrDomainPollCreatorRequired}
	ErrPollClosed           = &AppError{Code: "POLL_CLOSED", Message: "poll is closed and cannot accept vote changes", HTTPStatus: http.StatusConflict, Err: ErrDomainPollClosed}
	ErrPollPrivate          = &AppError{Code: "POLL_PRIVATE", Message: "voter identities are not available for private polls", HTTPStatus: http.StatusForbidden, Err: ErrDomainPollPrivate}
	ErrPollGroupInactive    = &AppError{Code: "POLL_GROUP_INACTIVE", Message: "polls cannot be changed while the group is inactive", HTTPStatus: http.StatusConflict, Err: ErrDomainPollGroupInactive}
	ErrPollInvalidOptions   = &AppError{Code: "POLL_INVALID_OPTIONS", Message: "provide at least two non-empty options with titles of at most 255 characters", HTTPStatus: http.StatusBadRequest, Err: ErrDomainPollInvalidOptions}
	ErrPollDuplicateOptions = &AppError{Code: "POLL_DUPLICATE_OPTIONS", Message: "poll options must be unique", HTTPStatus: http.StatusBadRequest, Err: ErrDomainPollDuplicateOptions}
)
