package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/finance"
	domainDB "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

type settlementService struct {
	repo     domainDB.SettlementRepository
	balances domainDB.FinancialBalanceRepository
	users    domainDB.UserRepository
	trx      database.TrxManager
	urls     adapter.AttachmentURLResolver
}

func NewSettlementService(repo domainDB.SettlementRepository, balances domainDB.FinancialBalanceRepository, users domainDB.UserRepository, trx database.TrxManager, urls adapter.AttachmentURLResolver) adapter.SettlementService {
	return &settlementService{repo: repo, balances: balances, users: users, trx: trx, urls: urls}
}

func (s *settlementService) withGroup(ctx context.Context, userID, groupID int64, write bool, fn func(context.Context, *entity.Group, *entity.GroupMember) error) error {
	if userID <= 0 || groupID <= 0 {
		return exception.ErrBadRequest
	}
	return s.trx.WithTransaction(ctx, func(tx context.Context) error {
		user, err := s.users.FindByID(tx, userID)
		if err != nil {
			return err
		}
		if user.Status != enum.UserStatusActive {
			return exception.ErrUserInactive
		}
		group, err := s.repo.FindGroupByID(tx, groupID, write)
		if err != nil {
			return err
		}
		member, err := s.repo.FindMemberByUserID(tx, groupID, userID)
		if err != nil {
			return err
		}
		if write && group.Status == enum.GroupStatusClosed {
			return exception.ErrSettlementGroupClosed
		}
		return fn(tx, group, member)
	})
}

func (s *settlementService) Create(ctx context.Context, userID, groupID int64, req dto.CreateSettlementRequest) (*dto.SettlementResponse, error) {
	if req.ToMemberID <= 0 || req.AmountToman <= 0 || !settlementText(req.Note, 2000) || len(req.AttachmentIDs) > 100 {
		return nil, exception.ErrBadRequest
	}
	if req.PaidAt != nil && (req.PaidAt.Year() < 1 || req.PaidAt.Year() > 9999) {
		return nil, exception.ErrBadRequest
	}
	var res dto.SettlementResponse
	err := s.withGroup(ctx, userID, groupID, true, func(tx context.Context, group *entity.Group, member *entity.GroupMember) error {
		if member.ID == req.ToMemberID {
			return exception.ErrSettlementSelfPayment
		}
		members, err := s.repo.FindMembersByIDs(tx, groupID, []int64{req.ToMemberID})
		if err != nil {
			return err
		}
		if len(members) != 1 || !activeSettlementMember(members[0]) {
			return exception.ErrSettlementRecipientNotFound
		}
		recipient, err := s.users.FindByID(tx, members[0].UserID)
		if err != nil {
			if errors.Is(err, exception.ErrUserNotFound) {
				return exception.ErrSettlementRecipientNotFound
			}
			return err
		}
		if recipient.Status != enum.UserStatusActive {
			return exception.ErrSettlementRecipientNotFound
		}
		if err := s.validateAttachments(tx, userID, groupID, req.AttachmentIDs); err != nil {
			return err
		}
		ledger, err := loadFinancialLedger(tx, s.balances, groupID)
		if err != nil {
			return err
		}
		positions, err := ledger.Positions()
		if err != nil {
			return financialCalculationError(err)
		}
		var debt int64
		for _, position := range positions {
			if position.MemberID == member.ID {
				debt = max(0, -position.NetToman)
				break
			}
		}
		paidAt := time.Now().UTC()
		if req.PaidAt != nil {
			paidAt = *req.PaidAt
		}
		revision := ledger.Revision()
		payment := &entity.Settlement{GroupID: groupID, FromMemberID: member.ID, ToMemberID: req.ToMemberID, AmountToman: req.AmountToman, Status: enum.SettlementStatusPending, Note: req.Note, PaidAt: paidAt, BalanceRevision: &revision}
		if err := s.repo.Create(tx, payment); err != nil {
			return err
		}
		if err := s.repo.BindAttachments(tx, groupID, payment.ID, req.AttachmentIDs); err != nil {
			return err
		}
		stored, err := s.repo.FindByID(tx, groupID, payment.ID, false)
		if err != nil {
			return err
		}
		res = dto.ToSettlementResponse(stored, group, member)
		if req.AmountToman > debt {
			res.Warnings = append(res.Warnings, dto.SettlementWarning{Code: "PAYMENT_EXCEEDS_CURRENT_DEBT", Message: "payment exceeds your current net debt in this group; registration is allowed"})
		}
		return s.resolveURLs(tx, userID, []*entity.Settlement{stored}, []*dto.SettlementResponse{&res})
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *settlementService) validateAttachments(ctx context.Context, userID, groupID int64, ids []int64) error {
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return exception.ErrBadRequest
		}
		if seen[id] {
			return exception.ErrSettlementDuplicateAttachments
		}
		seen[id] = true
	}
	items, err := s.repo.FindAttachmentsByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if len(items) != len(ids) {
		return exception.ErrAttachmentNotFound
	}
	for _, att := range items {
		if att.Type != enum.AttachmentTypeSettlementReceipt {
			return exception.ErrInvalidFileType
		}
		if att.UploadedByUserID == nil || *att.UploadedByUserID != userID || att.SettlementID != nil || att.ExpenseID != nil || (att.GroupID != nil && *att.GroupID != groupID) {
			return exception.ErrAttachmentForbidden
		}
	}
	return nil
}

func activeSettlementMember(m *entity.GroupMember) bool {
	return m != nil && m.Status == enum.GroupMemberStatusActive && m.LeftAt == nil
}

func settlementText(text *string, max int) bool {
	return text == nil || (utf8.ValidString(*text) && !strings.ContainsRune(*text, '\x00') && utf8.RuneCountInString(*text) <= max)
}

func settlementVisible(payment *entity.Settlement, memberID int64) bool {
	return payment.Status == enum.SettlementStatusSuccess || payment.Status == enum.SettlementStatusReject || payment.FromMemberID == memberID || payment.ToMemberID == memberID
}

func (s *settlementService) Get(ctx context.Context, userID, groupID, id int64) (*dto.SettlementResponse, error) {
	if id <= 0 {
		return nil, exception.ErrBadRequest
	}
	var res dto.SettlementResponse
	err := s.withGroup(ctx, userID, groupID, false, func(tx context.Context, group *entity.Group, member *entity.GroupMember) error {
		payment, err := s.repo.FindByID(tx, groupID, id, false)
		if err != nil {
			return err
		}
		if !settlementVisible(payment, member.ID) {
			return exception.ErrSettlementForbidden
		}
		res = dto.ToSettlementResponse(payment, group, member)
		if err := s.addBalanceWarnings(tx, groupID, []*entity.Settlement{payment}, []*dto.SettlementResponse{&res}); err != nil {
			return err
		}
		return s.resolveURLs(tx, userID, []*entity.Settlement{payment}, []*dto.SettlementResponse{&res})
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *settlementService) List(ctx context.Context, userID, groupID int64, req dto.ListSettlementsRequest) (*dto.SettlementListResponse, error) {
	page, limit, offset, err := expensePagination(req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	if req.Direction == "" {
		req.Direction = enum.SettlementDirectionAll
	}
	if !req.Direction.IsValid() || (req.Status != nil && !req.Status.IsValid()) {
		return nil, exception.ErrBadRequest
	}
	var res dto.SettlementListResponse
	err = s.withGroup(ctx, userID, groupID, false, func(tx context.Context, group *entity.Group, member *entity.GroupMember) error {
		payments, total, err := s.repo.List(tx, groupID, member.ID, domainDB.SettlementFilter{Status: req.Status, Direction: req.Direction}, limit, offset)
		if err != nil {
			return err
		}
		items := make([]dto.SettlementResponse, 0, len(payments))
		ptrs := make([]*dto.SettlementResponse, len(payments))
		for _, payment := range payments {
			items = append(items, dto.ToSettlementResponse(payment, group, member))
		}
		for i := range items {
			ptrs[i] = &items[i]
		}
		if err := s.addBalanceWarnings(tx, groupID, payments, ptrs); err != nil {
			return err
		}
		if err := s.resolveURLs(tx, userID, payments, ptrs); err != nil {
			return err
		}
		res = dto.SettlementListResponse{Items: items, Pagination: dto.ExpensePaginationResponse{Page: page, Limit: limit, Total: total}}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *settlementService) withPending(ctx context.Context, userID, groupID, id int64, recipientAction bool, fn func(context.Context, *entity.Group, *entity.GroupMember, *entity.Settlement) error) error {
	if id <= 0 {
		return exception.ErrBadRequest
	}
	return s.withGroup(ctx, userID, groupID, true, func(tx context.Context, group *entity.Group, member *entity.GroupMember) error {
		initial, err := s.repo.FindByID(tx, groupID, id, false)
		if err != nil {
			return err
		}
		if recipientAction && initial.ToMemberID != member.ID {
			return exception.ErrSettlementRecipientRequired
		}
		if !recipientAction && initial.FromMemberID != member.ID {
			return exception.ErrSettlementCreatorRequired
		}
		if _, err := s.repo.FindMembersByIDs(tx, groupID, []int64{initial.FromMemberID, initial.ToMemberID}); err != nil {
			return err
		}
		payment, err := s.repo.FindByID(tx, groupID, id, true)
		if err != nil {
			return err
		}
		if payment.Status != enum.SettlementStatusPending {
			return exception.ErrSettlementNotPending
		}
		return fn(tx, group, member, payment)
	})
}

func (s *settlementService) Confirm(ctx context.Context, userID, groupID, id int64) (*dto.SettlementResponse, error) {
	return s.respond(ctx, userID, groupID, id, enum.SettlementStatusSuccess)
}

func (s *settlementService) Reject(ctx context.Context, userID, groupID, id int64) (*dto.SettlementResponse, error) {
	return s.respond(ctx, userID, groupID, id, enum.SettlementStatusReject)
}

func (s *settlementService) respond(ctx context.Context, userID, groupID, id int64, status enum.SettlementStatus) (*dto.SettlementResponse, error) {
	var res dto.SettlementResponse
	err := s.withPending(ctx, userID, groupID, id, true, func(tx context.Context, group *entity.Group, member *entity.GroupMember, payment *entity.Settlement) error {
		var legs []finance.Transfer
		var revision string
		if status == enum.SettlementStatusSuccess {
			ledger, err := loadFinancialLedger(tx, s.balances, groupID)
			if err != nil {
				return err
			}
			revision = ledger.Revision()
			legs, err = ledger.ApplyPayment(finance.Transfer{FromMemberID: payment.FromMemberID, ToMemberID: payment.ToMemberID, AmountToman: payment.AmountToman})
			if err != nil {
				return financialCalculationError(err)
			}
			if _, err := ledger.Positions(); err != nil {
				return financialCalculationError(err)
			}
		}
		if err := s.repo.Respond(tx, groupID, id, status, legs); err != nil {
			return err
		}
		stored, err := s.repo.FindByID(tx, groupID, id, false)
		if err != nil {
			return err
		}
		res = dto.ToSettlementResponse(stored, group, member)
		if revision != "" {
			addBalanceWarning(payment, revision, &res)
		}
		return s.resolveURLs(tx, userID, []*entity.Settlement{stored}, []*dto.SettlementResponse{&res})
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *settlementService) Delete(ctx context.Context, userID, groupID, id int64) error {
	return s.withPending(ctx, userID, groupID, id, false, func(tx context.Context, _ *entity.Group, _ *entity.GroupMember, _ *entity.Settlement) error {
		return s.repo.DeletePending(tx, groupID, id)
	})
}

func (s *settlementService) addBalanceWarnings(ctx context.Context, groupID int64, payments []*entity.Settlement, responses []*dto.SettlementResponse) error {
	var revision string
	for i, payment := range payments {
		if payment.Status != enum.SettlementStatusPending || payment.BalanceRevision == nil {
			continue
		}
		if revision == "" {
			ledger, err := loadFinancialLedger(ctx, s.balances, groupID)
			if err != nil {
				return err
			}
			revision = ledger.Revision()
		}
		addBalanceWarning(payment, revision, responses[i])
	}
	return nil
}

func addBalanceWarning(payment *entity.Settlement, revision string, response *dto.SettlementResponse) {
	if payment.BalanceRevision != nil && *payment.BalanceRevision != revision {
		response.Warnings = append(response.Warnings, dto.SettlementWarning{
			Code:    "SETTLEMENT_BALANCE_CHANGED",
			Message: "the group balance or membership changed after this payment was registered",
		})
	}
}

func (s *settlementService) resolveURLs(ctx context.Context, userID int64, payments []*entity.Settlement, responses []*dto.SettlementResponse) error {
	ids := make([]int64, 0)
	for _, p := range payments {
		for _, att := range p.Attachments {
			ids = append(ids, att.ID)
		}
		for _, m := range []*entity.GroupMember{p.FromMember, p.ToMember} {
			if m != nil && m.User != nil && m.User.AvatarAttachmentID != nil {
				ids = append(ids, *m.User.AvatarAttachmentID)
			}
		}
	}
	urls, err := s.urls.ResolveAccessURLs(ctx, userID, ids)
	if err != nil {
		return err
	}
	for i, res := range responses {
		for j := range res.Attachments {
			if url := urls[res.Attachments[j].ID]; url != "" {
				res.Attachments[j].AccessURL = &url
			}
		}
		for _, pair := range []struct {
			source *entity.GroupMember
			target *dto.SettlementMemberResponse
		}{{payments[i].FromMember, &res.FromMember}, {payments[i].ToMember, &res.ToMember}} {
			if pair.source != nil && pair.source.User != nil && pair.source.User.AvatarAttachmentID != nil {
				if url := urls[*pair.source.User.AvatarAttachmentID]; url != "" {
					pair.target.AvatarURL = &url
				}
			}
		}
	}
	return nil
}
