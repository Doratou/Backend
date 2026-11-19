package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainSettlementNotFound             = errors.New("settlement not found")
	ErrDomainSettlementForbidden            = errors.New("settlement access forbidden")
	ErrDomainSettlementRecipientNotFound    = errors.New("settlement recipient not found")
	ErrDomainSettlementSelfPayment          = errors.New("settlement parties must differ")
	ErrDomainSettlementRecipientRequired    = errors.New("settlement recipient permission required")
	ErrDomainSettlementCreatorRequired      = errors.New("settlement creator permission required")
	ErrDomainSettlementNotPending           = errors.New("settlement is no longer pending")
	ErrDomainSettlementGroupClosed          = errors.New("settlement group closed")
	ErrDomainSettlementDuplicateAttachments = errors.New("settlement attachments must be unique")
	ErrDomainSettlementReceiptLocked        = errors.New("settlement receipt is immutable")
	ErrDomainGroupStatusTransition          = errors.New("invalid group lifecycle transition")
	ErrDomainGroupPendingSettlements        = errors.New("group has pending settlements")
	ErrDomainGroupHasSettlements            = errors.New("group has settlement history")
	ErrDomainGroupClosed                    = errors.New("group is closed")
)

var (
	ErrSettlementNotFound             = &AppError{Code: "SETTLEMENT_NOT_FOUND", Message: "settlement not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainSettlementNotFound}
	ErrSettlementForbidden            = &AppError{Code: "SETTLEMENT_FORBIDDEN", Message: "you do not have permission to access this settlement", HTTPStatus: http.StatusForbidden, Err: ErrDomainSettlementForbidden}
	ErrSettlementRecipientNotFound    = &AppError{Code: "SETTLEMENT_RECIPIENT_NOT_FOUND", Message: "recipient must be an active member with an active account in this group", HTTPStatus: http.StatusNotFound, Err: ErrDomainSettlementRecipientNotFound}
	ErrSettlementSelfPayment          = &AppError{Code: "SETTLEMENT_SELF_PAYMENT", Message: "you cannot register a payment to yourself", HTTPStatus: http.StatusBadRequest, Err: ErrDomainSettlementSelfPayment}
	ErrSettlementRecipientRequired    = &AppError{Code: "SETTLEMENT_RECIPIENT_REQUIRED", Message: "only the recipient can confirm or reject this payment", HTTPStatus: http.StatusForbidden, Err: ErrDomainSettlementRecipientRequired}
	ErrSettlementCreatorRequired      = &AppError{Code: "SETTLEMENT_CREATOR_REQUIRED", Message: "only the creator can delete this pending payment", HTTPStatus: http.StatusForbidden, Err: ErrDomainSettlementCreatorRequired}
	ErrSettlementNotPending           = &AppError{Code: "SETTLEMENT_NOT_PENDING", Message: "only pending payments can be confirmed, rejected or deleted", HTTPStatus: http.StatusConflict, Err: ErrDomainSettlementNotPending}
	ErrSettlementGroupClosed          = &AppError{Code: "SETTLEMENT_GROUP_CLOSED", Message: "payments cannot be changed after the group is closed", HTTPStatus: http.StatusConflict, Err: ErrDomainSettlementGroupClosed}
	ErrSettlementDuplicateAttachments = &AppError{Code: "SETTLEMENT_DUPLICATE_ATTACHMENTS", Message: "attachment_ids must not contain duplicate IDs", HTTPStatus: http.StatusBadRequest, Err: ErrDomainSettlementDuplicateAttachments}
	ErrSettlementReceiptLocked        = &AppError{Code: "SETTLEMENT_RECEIPT_LOCKED", Message: "attached payment receipts cannot be changed or deleted; delete the pending payment first", HTTPStatus: http.StatusConflict, Err: ErrDomainSettlementReceiptLocked}
	ErrGroupStatusTransition          = &AppError{Code: "GROUP_STATUS_TRANSITION", Message: "group status can only move from ACTIVE to SETTLING and then CLOSED", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupStatusTransition}
	ErrGroupPendingSettlements        = &AppError{Code: "GROUP_PENDING_SETTLEMENTS", Message: "resolve pending payments before closing the group", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupPendingSettlements}
	ErrGroupHasSettlements            = &AppError{Code: "GROUP_HAS_SETTLEMENTS", Message: "a group with payment history cannot be deleted", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupHasSettlements}
	ErrGroupClosed                    = &AppError{Code: "GROUP_CLOSED", Message: "closed groups are read-only", HTTPStatus: http.StatusConflict, Err: ErrDomainGroupClosed}
)
