package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/google/uuid"
)

func (s *attachmentService) snapshotSettlementReceipt(ctx context.Context, key string, size int64) (string, string, error) {
	file, err := s.storageRepo.DownloadFile(ctx, key)
	if err != nil {
		return "", "", exception.Wrap(exception.ErrStorageUnavailable, err)
	}
	defer file.Close()
	limit := s.consts.Attachment.MaxReceiptSize
	if limit <= 0 {
		limit = 10 << 20
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return "", "", exception.Wrap(exception.ErrStorageUnavailable, err)
	}
	if int64(len(data)) != size {
		return "", "", exception.ErrBadRequest
	}
	mime := http.DetectContentType(data)
	if err := s.validateFile(enum.AttachmentTypeSettlementReceipt, int64(len(data)), mime); err != nil {
		return "", "", err
	}
	storedKey := strings.Replace(key, "attachments/settlement_receipt/", "attachments/confirmed_settlement_receipt/", 1) + "." + uuid.NewString()
	if _, err := s.storageRepo.UploadFile(ctx, storedKey, bytes.NewReader(data), int64(len(data)), mime); err != nil {
		return "", "", exception.Wrap(exception.ErrStorageUnavailable, err)
	}
	return storedKey, mime, nil
}

func sameReceiptBinding(a, b *int64) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func (s *attachmentService) validateStoredReceipt(ctx context.Context, kind enum.AttachmentType, key string, size int64) error {
	file, err := s.storageRepo.DownloadFile(ctx, key)
	if err != nil {
		return exception.Wrap(exception.ErrStorageUnavailable, err)
	}
	defer file.Close()
	header := make([]byte, 512)
	n, err := io.ReadFull(file, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return exception.Wrap(exception.ErrStorageUnavailable, err)
	}
	return s.validateFile(kind, size, http.DetectContentType(header[:n]))
}

func (s *attachmentService) validateSettlementReceiptAccess(ctx context.Context, attachment *entity.Attachment, userID int64, write bool) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Status != enum.UserStatusActive {
		return exception.ErrUserInactive
	}
	if attachment.SettlementID == nil {
		if attachment.UploadedByUserID == nil || *attachment.UploadedByUserID != userID {
			return exception.ErrAttachmentForbidden
		}
		return nil
	}
	groupID, err := s.settlementRepo.FindGroupID(ctx, *attachment.SettlementID)
	if err != nil {
		return err
	}
	if attachment.GroupID == nil || *attachment.GroupID != groupID {
		return exception.ErrAttachmentForbidden
	}
	if _, err = s.settlementRepo.FindGroupByID(ctx, groupID, false); err != nil {
		return err
	}
	member, err := s.settlementRepo.FindMemberByUserID(ctx, groupID, userID)
	if errors.Is(err, exception.ErrSettlementForbidden) {
		return exception.ErrAttachmentForbidden
	}
	if err != nil {
		return err
	}
	payment, err := s.settlementRepo.FindByID(ctx, groupID, *attachment.SettlementID, false)
	if err != nil {
		return err
	}
	if !settlementVisible(payment, member.ID) {
		return exception.ErrAttachmentForbidden
	}
	if write {
		return exception.ErrSettlementReceiptLocked
	}
	return nil
}
