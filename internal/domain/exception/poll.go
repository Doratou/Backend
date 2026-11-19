package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainPollNotFound             = errors.New("poll not found")
	ErrDomainPollOptionNotFound       = errors.New("poll option not found")
	ErrDomainPollVoteNotFound         = errors.New("poll vote not found")
	ErrDomainPollForbidden            = errors.New("poll access is forbidden")
	ErrDomainPollCreatorRequired      = errors.New("poll creator permission is required")
	ErrDomainPollClosed               = errors.New("poll is closed")
	ErrDomainPollPrivate              = errors.New("poll voter identities are private")
	ErrDomainPollGroupInactive        = errors.New("poll changes are frozen while the group is inactive")
	ErrDomainPollInvalidOptions       = errors.New("poll options are invalid")
	ErrDomainPollDuplicateOptions     = errors.New("poll options must be unique")
	ErrDomainPollDuplicateVoteOptions = errors.New("poll vote option IDs must be unique")
	ErrDomainPollSingleChoiceRequired = errors.New("poll allows exactly one selected option")
)

// Poll
var (
	ErrPollNotFound             = &AppError{Code: "POLL_NOT_FOUND", Message: "poll not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainPollNotFound}
	ErrPollOptionNotFound       = &AppError{Code: "POLL_OPTION_NOT_FOUND", Message: "option not found in this poll", HTTPStatus: http.StatusNotFound, Err: ErrDomainPollOptionNotFound}
	ErrPollVoteNotFound         = &AppError{Code: "POLL_VOTE_NOT_FOUND", Message: "no active vote was found for this member in this poll", HTTPStatus: http.StatusNotFound, Err: ErrDomainPollVoteNotFound}
	ErrPollForbidden            = &AppError{Code: "POLL_FORBIDDEN", Message: "only active group members can access this poll", HTTPStatus: http.StatusForbidden, Err: ErrDomainPollForbidden}
	ErrPollCreatorRequired      = &AppError{Code: "POLL_CREATOR_REQUIRED", Message: "only the poll creator can perform this action", HTTPStatus: http.StatusForbidden, Err: ErrDomainPollCreatorRequired}
	ErrPollClosed               = &AppError{Code: "POLL_CLOSED", Message: "poll is closed and cannot accept vote changes", HTTPStatus: http.StatusConflict, Err: ErrDomainPollClosed}
	ErrPollPrivate              = &AppError{Code: "POLL_PRIVATE", Message: "voter identities are not available for private polls", HTTPStatus: http.StatusForbidden, Err: ErrDomainPollPrivate}
	ErrPollGroupInactive        = &AppError{Code: "POLL_GROUP_INACTIVE", Message: "polls cannot be changed while the group is inactive", HTTPStatus: http.StatusConflict, Err: ErrDomainPollGroupInactive}
	ErrPollInvalidOptions       = &AppError{Code: "POLL_INVALID_OPTIONS", Message: "provide at least two non-empty options with titles of at most 255 characters", HTTPStatus: http.StatusBadRequest, Err: ErrDomainPollInvalidOptions}
	ErrPollDuplicateOptions     = &AppError{Code: "POLL_DUPLICATE_OPTIONS", Message: "poll options must be unique", HTTPStatus: http.StatusBadRequest, Err: ErrDomainPollDuplicateOptions}
	ErrPollDuplicateVoteOptions = &AppError{Code: "POLL_DUPLICATE_VOTE_OPTIONS", Message: "option_ids must not contain duplicate IDs", HTTPStatus: http.StatusBadRequest, Err: ErrDomainPollDuplicateVoteOptions}
	ErrPollSingleChoiceRequired = &AppError{Code: "POLL_SINGLE_CHOICE_REQUIRED", Message: "this poll allows exactly one selected option", HTTPStatus: http.StatusBadRequest, Err: ErrDomainPollSingleChoiceRequired}
)
