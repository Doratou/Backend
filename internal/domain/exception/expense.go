package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainExpenseNotFound             = errors.New("expense not found")
	ErrDomainExpenseForbidden            = errors.New("expense access is forbidden")
	ErrDomainExpenseEditorRequired       = errors.New("expense creator or group owner permission is required")
	ErrDomainExpenseGroupInactive        = errors.New("expense changes are frozen while the group is inactive")
	ErrDomainExpenseInvalidSplit         = errors.New("expense split values are invalid")
	ErrDomainExpenseSplitTotalMismatch   = errors.New("expense split total does not match")
	ErrDomainExpenseDuplicateMembers     = errors.New("expense participants must be unique")
	ErrDomainExpenseParticipantNotFound  = errors.New("expense participant was never a member of this group")
	ErrDomainExpenseDuplicateAttachments = errors.New("expense receipt attachments must be unique")
)

var (
	ErrExpenseNotFound             = &AppError{Code: "EXPENSE_NOT_FOUND", Message: "expense not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainExpenseNotFound}
	ErrExpenseForbidden            = &AppError{Code: "EXPENSE_FORBIDDEN", Message: "only active group members can access expenses", HTTPStatus: http.StatusForbidden, Err: ErrDomainExpenseForbidden}
	ErrExpenseEditorRequired       = &AppError{Code: "EXPENSE_EDITOR_REQUIRED", Message: "only the expense creator or current group owner can edit or delete this expense", HTTPStatus: http.StatusForbidden, Err: ErrDomainExpenseEditorRequired}
	ErrExpenseGroupInactive        = &AppError{Code: "EXPENSE_GROUP_INACTIVE", Message: "expenses cannot be changed while the group is inactive", HTTPStatus: http.StatusConflict, Err: ErrDomainExpenseGroupInactive}
	ErrExpenseInvalidSplit         = &AppError{Code: "EXPENSE_INVALID_SPLIT", Message: "provide valid split values for the selected split type", HTTPStatus: http.StatusBadRequest, Err: ErrDomainExpenseInvalidSplit}
	ErrExpenseSplitTotalMismatch   = &AppError{Code: "EXPENSE_SPLIT_TOTAL_MISMATCH", Message: "exact shares must sum to total_amount_toman and percentages must sum to 100", HTTPStatus: http.StatusBadRequest, Err: ErrDomainExpenseSplitTotalMismatch}
	ErrExpenseDuplicateMembers     = &AppError{Code: "EXPENSE_DUPLICATE_MEMBERS", Message: "shares must not contain duplicate member IDs", HTTPStatus: http.StatusBadRequest, Err: ErrDomainExpenseDuplicateMembers}
	ErrExpenseParticipantNotFound  = &AppError{Code: "EXPENSE_PARTICIPANT_NOT_FOUND", Message: "a selected participant does not belong to this group's current or former members", HTTPStatus: http.StatusNotFound, Err: ErrDomainExpenseParticipantNotFound}
	ErrExpenseDuplicateAttachments = &AppError{Code: "EXPENSE_DUPLICATE_ATTACHMENTS", Message: "attachment_ids must not contain duplicate IDs", HTTPStatus: http.StatusBadRequest, Err: ErrDomainExpenseDuplicateAttachments}
)
