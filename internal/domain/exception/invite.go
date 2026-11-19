package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainInviteNotFound          = errors.New("group invite not found")
	ErrDomainInviteExpired           = errors.New("group invite has expired")
	ErrDomainInviteRevoked           = errors.New("group invite has been deactivated")
	ErrDomainInviteUsesExhausted     = errors.New("group invite has reached its usage limit")
	ErrDomainInvitePasswordInvalid   = errors.New("group invite password is invalid")
	ErrDomainGroupActiveInviteExists = errors.New("group already has an active invitation link")
	ErrDomainDirectInvitePending     = errors.New("a direct group invitation is already pending")
	ErrDomainDirectInviteCooldown    = errors.New("direct group invitation cooldown is active")
	ErrDomainDirectInviteNotPending  = errors.New("direct group invitation is no longer pending")
)

// Invite
var (
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
