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
)

var (
	ErrLuckyDrawNotFound            = &AppError{Code: "DRAW_NOT_FOUND", Message: "lucky draw not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainLuckyDrawNotFound}
	ErrLuckyDrawForbidden           = &AppError{Code: "DRAW_FORBIDDEN", Message: "only active group members can access this lucky draw", HTTPStatus: http.StatusForbidden, Err: ErrDomainLuckyDrawForbidden}
	ErrLuckyDrawCreatorRequired     = &AppError{Code: "DRAW_CREATOR_REQUIRED", Message: "only the lucky draw creator can perform this action", HTTPStatus: http.StatusForbidden, Err: ErrDomainLuckyDrawCreatorRequired}
	ErrLuckyDrawCompleted           = &AppError{Code: "DRAW_COMPLETED", Message: "a completed lucky draw cannot be edited", HTTPStatus: http.StatusConflict, Err: ErrDomainLuckyDrawCompleted}
	ErrLuckyDrawGroupInactive       = &AppError{Code: "DRAW_GROUP_INACTIVE", Message: "lucky draws cannot be changed while the group is inactive", HTTPStatus: http.StatusConflict, Err: ErrDomainLuckyDrawGroupInactive}
	ErrLuckyDrawInvalidCandidates   = &AppError{Code: "DRAW_INVALID_CANDIDATES", Message: "provide 2 to 100 candidates with non-empty titles of at most 255 characters and integer weights from 1 to 1000000", HTTPStatus: http.StatusBadRequest, Err: ErrDomainLuckyDrawInvalidCandidates}
	ErrLuckyDrawDuplicateCandidates = &AppError{Code: "DRAW_DUPLICATE_CANDIDATES", Message: "lucky draw candidate titles must be unique", HTTPStatus: http.StatusBadRequest, Err: ErrDomainLuckyDrawDuplicateCandidates}
)
