package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainGroupNotFound                = errors.New("group not found")
	ErrDomainGroupForbidden               = errors.New("group access is forbidden")
	ErrDomainGroupMemberNotFound          = errors.New("active group member not found")
	ErrDomainGroupOwnerRequired           = errors.New("group owner permission is required")
	ErrDomainGroupOwnerCannotLeave        = errors.New("group owner must transfer ownership before leaving")
	ErrDomainGroupOwnerCannotBeRemoved    = errors.New("group owner cannot be removed")
	ErrDomainGroupTransferTargetNotMember = errors.New("new group owner must be an active member")
	ErrDomainAlreadyGroupMember           = errors.New("user is already an active group member")
	ErrDomainGroupInactive                = errors.New("group is inactive")
	ErrDomainGroupMustBeInactiveToClone   = errors.New("only SETTLING or CLOSED groups can be cloned")
)

// Group & Group member
var (
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
	ErrGroupInvalidStatus           = &AppError{Code: "GROUP_INVALID_STATUS", Message: "status must be ACTIVE, SETTLING or CLOSED", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidIsPremium        = &AppError{Code: "GROUP_INVALID_IS_PREMIUM", Message: "is_premium must be true or false", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidRole             = &AppError{Code: "GROUP_INVALID_ROLE", Message: "role must be OWNER or MEMBER", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidYear             = &AppError{Code: "GROUP_INVALID_YEAR", Message: "year must be a valid Jalali (1300-1599) or Gregorian (1900-2200) year", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidSortBy           = &AppError{Code: "GROUP_INVALID_SORT_BY", Message: "sort_by must be NEWEST, OLDEST, TITLE_ASC, MOST_MEMBERS or ARCHIVED_AT", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrGroupInvalidTransferOwner    = &AppError{Code: "GROUP_INVALID_TRANSFER_OWNER", Message: "new_owner_id must be a positive integer", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrAlreadyGroupMember           = &AppError{Code: "ALREADY_GROUP_MEMBER", Message: "you are already an active member of this group", HTTPStatus: http.StatusConflict, Err: ErrDomainAlreadyGroupMember}
	ErrGroupInactive                = &AppError{Code: "GROUP_INACTIVE", Message: "group is inactive and cannot accept new members", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupInactive}
	ErrGroupMustBeInactiveToClone   = &AppError{Code: "GROUP_MUST_BE_INACTIVE_TO_CLONE", Message: "only SETTLING or CLOSED groups can be cloned", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupMustBeInactiveToClone}
)
