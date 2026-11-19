package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainLuckyDrawNotFound            = errors.New("lucky draw not found")
	ErrDomainLuckyDrawForbidden           = errors.New("lucky draw access is forbidden")
	ErrDomainLuckyDrawCreatorRequired     = errors.New("lucky draw creator permission is required")
	ErrDomainLuckyDrawCompleted           = errors.New("lucky draw has already been completed")
	ErrDomainLuckyDrawGroupInactive       = errors.New("lucky draw changes are frozen while the group is inactive")
	ErrDomainLuckyDrawInvalidCandidates   = errors.New("lucky draw candidates are invalid")
	ErrDomainLuckyDrawDuplicateCandidates = errors.New("lucky draw candidate titles must be unique")
	ErrDomainLuckyDrawDraftExists         = errors.New("a draft lucky draw already exists in this group")
	ErrDomainLuckyDrawInvalidPercentages  = errors.New("lucky draw percentages are invalid")
	ErrDomainLuckyDrawInvalidType         = errors.New("lucky draw type is invalid")
	ErrDomainLuckyDrawCandidateNotMember  = errors.New("a lucky draw candidate is not an active group member")
)

var (
	ErrLuckyDrawNotFound            = &AppError{Code: "DRAW_NOT_FOUND", Message: "lucky draw not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainLuckyDrawNotFound}
	ErrLuckyDrawForbidden           = &AppError{Code: "DRAW_FORBIDDEN", Message: "only active group members can access this lucky draw", HTTPStatus: http.StatusForbidden, Err: ErrDomainLuckyDrawForbidden}
	ErrLuckyDrawCreatorRequired     = &AppError{Code: "DRAW_CREATOR_REQUIRED", Message: "only the lucky draw creator can perform this action", HTTPStatus: http.StatusForbidden, Err: ErrDomainLuckyDrawCreatorRequired}
	ErrLuckyDrawCompleted           = &AppError{Code: "DRAW_COMPLETED", Message: "a completed lucky draw cannot be edited", HTTPStatus: http.StatusConflict, Err: ErrDomainLuckyDrawCompleted}
	ErrLuckyDrawGroupInactive       = &AppError{Code: "DRAW_GROUP_INACTIVE", Message: "lucky draws cannot be changed while the group is inactive", HTTPStatus: http.StatusConflict, Err: ErrDomainLuckyDrawGroupInactive}
	ErrLuckyDrawInvalidCandidates   = &AppError{Code: "DRAW_INVALID_CANDIDATES", Message: "provide 2 to 100 candidates of type TEXT with a non-empty string of at most 255 characters, or USER with a positive integer user ID", HTTPStatus: http.StatusBadRequest, Err: ErrDomainLuckyDrawInvalidCandidates}
	ErrLuckyDrawDuplicateCandidates = &AppError{Code: "DRAW_DUPLICATE_CANDIDATES", Message: "text options and user IDs must be unique within their candidate type", HTTPStatus: http.StatusBadRequest, Err: ErrDomainLuckyDrawDuplicateCandidates}
	ErrLuckyDrawDraftExists         = &AppError{Code: "DRAW_DRAFT_EXISTS", Message: "this group already has an unexecuted lucky draw", HTTPStatus: http.StatusConflict, Err: ErrDomainLuckyDrawDraftExists}
	ErrLuckyDrawInvalidPercentages  = &AppError{Code: "DRAW_INVALID_PERCENTAGES", Message: "weighted chances must be greater than zero, have at most two decimal places and total exactly 100; equal chances are assigned automatically", HTTPStatus: http.StatusBadRequest, Err: ErrDomainLuckyDrawInvalidPercentages}
	ErrLuckyDrawInvalidType         = &AppError{Code: "DRAW_INVALID_TYPE", Message: "draw_type must be EQUAL or WEIGHTED", HTTPStatus: http.StatusBadRequest, Err: ErrDomainLuckyDrawInvalidType}
	ErrLuckyDrawCandidateNotMember  = &AppError{Code: "DRAW_CANDIDATE_NOT_MEMBER", Message: "user candidates must be active accounts and active members of this group", HTTPStatus: http.StatusBadRequest, Err: ErrDomainLuckyDrawCandidateNotMember}
)
