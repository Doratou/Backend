package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainChecklistForbidden        = errors.New("checklist access forbidden")
	ErrDomainChecklistGroupInactive    = errors.New("checklist is frozen")
	ErrDomainChecklistItemNotFound     = errors.New("checklist item not found")
	ErrDomainChecklistItemClaimed      = errors.New("checklist item is already claimed")
	ErrDomainChecklistItemNotClaimable = errors.New("checklist item is not claimable")
	ErrDomainChecklistAssigneeRequired = errors.New("checklist assignee permission required")
	ErrDomainChecklistOwnerRequired    = errors.New("checklist group owner permission required")
)

var (
	ErrChecklistForbidden        = &AppError{Code: "CHECKLIST_FORBIDDEN", Message: "only active group members can access the checklist", HTTPStatus: http.StatusForbidden, Err: ErrDomainChecklistForbidden}
	ErrChecklistGroupInactive    = &AppError{Code: "CHECKLIST_GROUP_INACTIVE", Message: "the checklist cannot be changed while the group is inactive", HTTPStatus: http.StatusConflict, Err: ErrDomainChecklistGroupInactive}
	ErrChecklistItemNotFound     = &AppError{Code: "CHECKLIST_ITEM_NOT_FOUND", Message: "checklist item not found in this group", HTTPStatus: http.StatusNotFound, Err: ErrDomainChecklistItemNotFound}
	ErrChecklistItemClaimed      = &AppError{Code: "CHECKLIST_ITEM_ALREADY_CLAIMED", Message: "another member has already claimed this item", HTTPStatus: http.StatusConflict, Err: ErrDomainChecklistItemClaimed}
	ErrChecklistItemNotClaimable = &AppError{Code: "CHECKLIST_ITEM_NOT_CLAIMABLE", Message: "personal reminders cannot be claimed or released", HTTPStatus: http.StatusConflict, Err: ErrDomainChecklistItemNotClaimable}
	ErrChecklistAssigneeRequired = &AppError{Code: "CHECKLIST_ASSIGNEE_REQUIRED", Message: "only the member who claimed this item can release it", HTTPStatus: http.StatusForbidden, Err: ErrDomainChecklistAssigneeRequired}
	ErrChecklistOwnerRequired    = &AppError{Code: "CHECKLIST_OWNER_REQUIRED", Message: "only the group owner can delete a claimed item", HTTPStatus: http.StatusForbidden, Err: ErrDomainChecklistOwnerRequired}
)
