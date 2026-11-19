package service

import (
	"context"
	"errors"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
)

func isReceipt(attType enum.AttachmentType) bool {
	return attType == enum.AttachmentTypeExpenseReceipt || attType == enum.AttachmentTypeSettlementReceipt
}

func (s *attachmentService) validateReceiptUpload(ctx context.Context, userID int64, attType enum.AttachmentType, expenseID, settlementID *int64) error {
	if settlementID != nil || (isReceipt(attType) && expenseID != nil) {
		return exception.ErrBadRequest
	}
	if !isReceipt(attType) {
		return nil
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Status != enum.UserStatusActive {
		return exception.ErrUserInactive
	}
	return nil
}

func (s *attachmentService) validateReceiptObjectKey(ctx context.Context, userID int64, attType enum.AttachmentType, key string) error {
	if strings.HasPrefix(key, "attachments/confirmed_settlement_receipt/") {
		return exception.ErrAttachmentForbidden
	}
	if !isReceipt(attType) {
		if strings.HasPrefix(key, "attachments/expense_receipt/") || strings.HasPrefix(key, "attachments/settlement_receipt/") {
			return exception.ErrAttachmentForbidden
		}
		return nil
	}
	prefix := "attachments/" + strings.ToLower(string(attType)) + "/" + strconv.FormatInt(userID, 10) + "/"
	if !strings.HasPrefix(key, prefix) || path.Clean(key) != key || strings.ContainsAny(key, "\\\r\n\x00") {
		return exception.ErrAttachmentForbidden
	}
	if _, err := s.attachmentRepo.FindByObjectKey(ctx, key); err == nil {
		return exception.ErrAttachmentForbidden
	} else if !errors.Is(err, exception.ErrAttachmentNotFound) {
		return err
	}
	return nil
}

func (s *attachmentService) validateExpenseReceiptAccess(ctx context.Context, attachment *entity.Attachment, userID int64, write bool) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Status != enum.UserStatusActive {
		return exception.ErrUserInactive
	}
	if attachment.ExpenseID == nil {
		if attachment.UploadedByUserID == nil || *attachment.UploadedByUserID != userID {
			return exception.ErrAttachmentForbidden
		}
		return nil
	}
	groupID, err := s.expenseRepo.FindGroupIDByExpenseID(ctx, *attachment.ExpenseID)
	if err != nil {
		return err
	}
	if attachment.GroupID != nil && *attachment.GroupID != groupID {
		return exception.ErrAttachmentForbidden
	}
	group, err := s.expenseRepo.FindGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	member, err := s.expenseRepo.FindActiveMemberByUserID(ctx, groupID, userID)
	if errors.Is(err, exception.ErrGroupMemberNotFound) {
		return exception.ErrAttachmentForbidden
	}
	if err != nil {
		return err
	}
	if write {
		if group.Status != enum.GroupStatusActive {
			return exception.ErrExpenseGroupInactive
		}
		expense, err := s.expenseRepo.FindByID(ctx, groupID, *attachment.ExpenseID)
		if err != nil {
			return err
		}
		if expense.CreatedByMemberID != member.ID && group.CreatedByUserID != userID {
			return exception.ErrExpenseEditorRequired
		}
	}
	return nil
}

func (s *attachmentService) lockExpenseReceipt(ctx context.Context, attachment *entity.Attachment) error {
	if !isReceipt(attachment.Type) {
		return nil
	}
	locked, err := s.expenseRepo.FindAttachmentsByIDs(ctx, []int64{attachment.ID})
	if err != nil {
		return err
	}
	if len(locked) != 1 {
		return exception.ErrAttachmentNotFound
	}
	current := locked[0]
	if !sameReceiptBinding(current.SettlementID, attachment.SettlementID) {
		return exception.ErrBadRequest
	}
	if (current.ExpenseID == nil) != (attachment.ExpenseID == nil) || (current.ExpenseID != nil && *current.ExpenseID != *attachment.ExpenseID) {
		return exception.ErrBadRequest
	}
	return nil
}

func (s *attachmentService) ResolveAccessURLs(ctx context.Context, userID int64, ids []int64) (map[int64]string, error) {
	return s.resolveAccessURLs(ctx, userID, ids, s.cfg.DefaultPresignedTTL)
}

func (s *attachmentService) resolveAccessURLs(ctx context.Context, userID int64, ids []int64, expiry time.Duration) (map[int64]string, error) {
	if expiry <= 0 {
		expiry = s.cfg.DefaultPresignedTTL
	}
	urls := make(map[int64]string)
	if len(ids) == 0 {
		return urls, nil
	}
	attachments, err := s.attachmentRepo.FindAccessibleByIDs(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		if attachment.ObjectKey == nil {
			continue
		}
		keys = append(keys, *attachment.ObjectKey)
	}
	if len(keys) > 0 {
		signed, err := s.storageRepo.GetPresignedURLsBatch(ctx, keys, expiry)
		if err != nil {
			return nil, exception.Wrap(exception.ErrStorageUnavailable, err)
		}
		for _, attachment := range attachments {
			if attachment.ObjectKey != nil {
				if url := signed[*attachment.ObjectKey]; url != "" {
					urls[attachment.ID] = url
				}
			}
		}
	}
	return urls, nil
}
