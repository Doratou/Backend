package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

type expenseService struct {
	expenseRepo    domain_db.ExpenseRepository
	userRepo       domain_db.UserRepository
	dbTrx          database.TrxManager
	attachmentURLs adapter.AttachmentURLResolver
}

func NewExpenseService(expenseRepo domain_db.ExpenseRepository, userRepo domain_db.UserRepository, dbTrx database.TrxManager, attachmentURLs adapter.AttachmentURLResolver) adapter.ExpenseService {
	return &expenseService{expenseRepo: expenseRepo, userRepo: userRepo, dbTrx: dbTrx, attachmentURLs: attachmentURLs}
}

func (s *expenseService) Create(ctx context.Context, userID, groupID int64, req dto.SaveExpenseRequest) (*dto.ExpenseResponse, error) {
	expense, err := newExpense(req)
	if err != nil {
		return nil, err
	}
	var response dto.ExpenseResponse
	err = s.withGroupAccess(ctx, userID, groupID, true, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		if err := s.validateParticipants(txCtx, group.ID, expense.Shares); err != nil {
			return err
		}
		expense.GroupID = group.ID
		expense.CreatedByMemberID = member.ID
		if expense.OccurredAt == nil {
			now := time.Now().UTC()
			expense.OccurredAt = &now
		}
		if err := s.expenseRepo.CreateWithShares(txCtx, expense); err != nil {
			return err
		}
		if err := s.replaceAttachments(txCtx, userID, expense, req.AttachmentIDs); err != nil {
			return err
		}
		stored, err := s.expenseRepo.FindByID(txCtx, group.ID, expense.ID)
		if err != nil {
			return err
		}
		response = dto.ToExpenseResponse(stored, group, member)
		return s.resolveExpenseURLs(txCtx, userID, []*dto.ExpenseResponse{&response})
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *expenseService) List(ctx context.Context, userID, groupID int64, req dto.ListExpensesRequest) (*dto.ExpenseListResponse, error) {
	page, limit, offset, err := expensePagination(req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	if req.Scope == "" {
		req.Scope = enum.ExpenseListScopeAll
	}
	if !req.Scope.IsValid() {
		return nil, exception.ErrBadRequest
	}
	var response dto.ExpenseListResponse
	err = s.withGroupAccess(ctx, userID, groupID, false, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		expenses, total, err := s.expenseRepo.FindByGroupID(txCtx, group.ID, member.ID, req.Scope, limit, offset)
		if err != nil {
			return err
		}
		items := make([]dto.ExpenseResponse, 0, len(expenses))
		for _, expense := range expenses {
			items = append(items, dto.ToExpenseResponse(expense, group, member))
		}
		pointers := make([]*dto.ExpenseResponse, len(items))
		for i := range items {
			pointers[i] = &items[i]
		}
		if err := s.resolveExpenseURLs(txCtx, userID, pointers); err != nil {
			return err
		}
		response = dto.ExpenseListResponse{Items: items, Pagination: dto.ExpensePaginationResponse{Page: page, Limit: limit, Total: total}}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *expenseService) Get(ctx context.Context, userID, groupID, expenseID int64) (*dto.ExpenseResponse, error) {
	var response dto.ExpenseResponse
	err := s.withExpenseAccess(ctx, userID, groupID, expenseID, false, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember, expense *entity.Expense) error {
		response = dto.ToExpenseResponse(expense, group, member)
		return s.resolveExpenseURLs(txCtx, userID, []*dto.ExpenseResponse{&response})
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *expenseService) Update(ctx context.Context, userID, groupID, expenseID int64, req dto.SaveExpenseRequest) (*dto.ExpenseResponse, error) {
	replacement, err := newExpense(req)
	if err != nil {
		return nil, err
	}
	var response dto.ExpenseResponse
	err = s.withExpenseAccess(ctx, userID, groupID, expenseID, true, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember, expense *entity.Expense) error {
		var err error
		response, err = s.updateExpense(txCtx, userID, group, member, expense, replacement, req.AttachmentIDs)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *expenseService) Patch(ctx context.Context, userID, groupID, expenseID int64, patch dto.PatchExpenseRequest) (*dto.ExpenseResponse, error) {
	if !patch.Title.Present && !patch.Description.Present && !patch.TotalAmountToman.Present && !patch.SplitType.Present && !patch.OccurredAt.Present && !patch.Shares.Present && !patch.AttachmentIDs.Present {
		return nil, exception.ErrBadRequest
	}
	if patch.Title.Null || patch.TotalAmountToman.Null || patch.SplitType.Null || patch.OccurredAt.Null || patch.Shares.Null || patch.AttachmentIDs.Null {
		return nil, exception.ErrBadRequest
	}
	var response dto.ExpenseResponse
	err := s.withExpenseAccess(ctx, userID, groupID, expenseID, true, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember, expense *entity.Expense) error {
		if expense.CreatedByMemberID != member.ID && group.CreatedByUserID != userID {
			return exception.ErrExpenseEditorRequired
		}
		req := dto.SaveExpenseRequest{Title: expense.Title, Description: expense.Description, TotalAmountToman: expense.TotalAmountToman, SplitType: expense.SplitType, OccurredAt: expense.OccurredAt, Shares: make([]dto.ExpenseShareRequest, 0, len(expense.Shares))}
		for _, share := range expense.Shares {
			req.Shares = append(req.Shares, dto.ExpenseShareRequest{MemberID: share.MemberID, Value: json.RawMessage(dto.ExpenseSplitJSONValue(expense.SplitType, share.SplitValue))})
		}
		if patch.Title.Present {
			req.Title = patch.Title.Value
		}
		if patch.Description.Present {
			req.Description = nil
			if !patch.Description.Null {
				req.Description = &patch.Description.Value
			}
		}
		if patch.TotalAmountToman.Present {
			req.TotalAmountToman = patch.TotalAmountToman.Value
		}
		if patch.SplitType.Present {
			if patch.SplitType.Value != expense.SplitType && !patch.Shares.Present {
				return exception.ErrExpenseInvalidSplit
			}
			req.SplitType = patch.SplitType.Value
		}
		if patch.OccurredAt.Present {
			req.OccurredAt = &patch.OccurredAt.Value
		}
		if patch.Shares.Present {
			req.Shares = patch.Shares.Value
		}
		if patch.AttachmentIDs.Present {
			req.AttachmentIDs = &patch.AttachmentIDs.Value
		}
		replacement, err := newExpense(req)
		if err != nil {
			return err
		}
		response, err = s.updateExpense(txCtx, userID, group, member, expense, replacement, req.AttachmentIDs)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *expenseService) updateExpense(ctx context.Context, userID int64, group *entity.Group, member *entity.GroupMember, expense, replacement *entity.Expense, attachmentIDs *[]int64) (dto.ExpenseResponse, error) {
	if expense.CreatedByMemberID != member.ID && group.CreatedByUserID != userID {
		return dto.ExpenseResponse{}, exception.ErrExpenseEditorRequired
	}
	if err := s.validateParticipants(ctx, group.ID, replacement.Shares); err != nil {
		return dto.ExpenseResponse{}, err
	}
	expense.Title = replacement.Title
	expense.Description = replacement.Description
	expense.TotalAmountToman = replacement.TotalAmountToman
	expense.SplitType = replacement.SplitType
	expense.Shares = replacement.Shares
	if replacement.OccurredAt != nil {
		expense.OccurredAt = replacement.OccurredAt
	}
	if err := s.expenseRepo.UpdateWithShares(ctx, expense); err != nil {
		return dto.ExpenseResponse{}, err
	}
	if err := s.replaceAttachments(ctx, userID, expense, attachmentIDs); err != nil {
		return dto.ExpenseResponse{}, err
	}
	stored, err := s.expenseRepo.FindByID(ctx, group.ID, expense.ID)
	if err != nil {
		return dto.ExpenseResponse{}, err
	}
	response := dto.ToExpenseResponse(stored, group, member)
	if err := s.resolveExpenseURLs(ctx, userID, []*dto.ExpenseResponse{&response}); err != nil {
		return dto.ExpenseResponse{}, err
	}
	return response, nil
}

func (s *expenseService) Delete(ctx context.Context, userID, groupID, expenseID int64) error {
	return s.withExpenseAccess(ctx, userID, groupID, expenseID, true, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember, expense *entity.Expense) error {
		if expense.CreatedByMemberID != member.ID && group.CreatedByUserID != userID {
			return exception.ErrExpenseEditorRequired
		}
		return s.expenseRepo.DeleteByID(txCtx, group.ID, expense.ID)
	})
}

func (s *expenseService) ListParticipants(ctx context.Context, userID, groupID int64, req dto.ExpensePaginationRequest) (*dto.ExpenseParticipantListResponse, error) {
	page, limit, offset, err := expensePagination(req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	var response dto.ExpenseParticipantListResponse
	err = s.withGroupAccess(ctx, userID, groupID, false, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		members, total, err := s.expenseRepo.FindParticipants(txCtx, group.ID, limit, offset)
		if err != nil {
			return err
		}
		items := make([]dto.ExpenseParticipantResponse, 0, len(members))
		for _, participant := range members {
			items = append(items, dto.ToExpenseParticipantResponse(participant))
		}
		avatarIDs := make([]int64, 0, len(items))
		for _, item := range items {
			if item.AvatarAttachmentID != nil {
				avatarIDs = append(avatarIDs, *item.AvatarAttachmentID)
			}
		}
		if len(avatarIDs) > 0 {
			urls, err := s.attachmentURLs.ResolveAccessURLs(txCtx, userID, avatarIDs)
			if err != nil {
				return err
			}
			for i := range items {
				if items[i].AvatarAttachmentID != nil {
					if url := urls[*items[i].AvatarAttachmentID]; url != "" {
						items[i].AvatarURL = &url
					}
				}
			}
		}
		response = dto.ExpenseParticipantListResponse{Items: items, Pagination: dto.ExpensePaginationResponse{Page: page, Limit: limit, Total: total}}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *expenseService) replaceAttachments(ctx context.Context, userID int64, expense *entity.Expense, ids *[]int64) error {
	if ids == nil {
		return nil
	}
	seen := make(map[int64]struct{}, len(*ids))
	for _, id := range *ids {
		if id <= 0 {
			return exception.ErrBadRequest
		}
		if _, exists := seen[id]; exists {
			return exception.ErrExpenseDuplicateAttachments
		}
		seen[id] = struct{}{}
	}
	attachments, err := s.expenseRepo.FindAttachmentsByIDs(ctx, *ids)
	if err != nil {
		return err
	}
	if len(attachments) != len(*ids) {
		return exception.ErrAttachmentNotFound
	}
	for _, attachment := range attachments {
		if attachment.Type != enum.AttachmentTypeExpenseReceipt {
			return exception.ErrInvalidFileType
		}
		if attachment.GroupID != nil && *attachment.GroupID != expense.GroupID {
			return exception.ErrAttachmentForbidden
		}
		if attachment.ExpenseID != nil {
			if *attachment.ExpenseID != expense.ID {
				return exception.ErrAttachmentForbidden
			}
		} else if attachment.UploadedByUserID == nil || *attachment.UploadedByUserID != userID {
			return exception.ErrAttachmentForbidden
		}
	}
	return s.expenseRepo.ReplaceAttachments(ctx, expense.GroupID, expense.ID, *ids)
}

func (s *expenseService) resolveExpenseURLs(ctx context.Context, userID int64, items []*dto.ExpenseResponse) error {
	ids := make([]int64, 0)
	for _, item := range items {
		ids = append(ids, item.AttachmentIDs...)
	}
	if len(ids) == 0 {
		return nil
	}
	urls, err := s.attachmentURLs.ResolveAccessURLs(ctx, userID, ids)
	if err != nil {
		return err
	}
	for _, item := range items {
		for i := range item.Attachments {
			if url := urls[item.Attachments[i].ID]; url != "" {
				item.Attachments[i].AccessURL = &url
			}
		}
	}
	return nil
}

func (s *expenseService) withGroupAccess(ctx context.Context, userID, groupID int64, write bool, fn func(context.Context, *entity.Group, *entity.GroupMember) error) error {
	if userID <= 0 || groupID <= 0 {
		return exception.ErrBadRequest
	}
	return s.dbTrx.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.userRepo.FindByID(txCtx, userID)
		if err != nil {
			return err
		}
		if user.Status != enum.UserStatusActive {
			return exception.ErrUserInactive
		}
		group, err := s.expenseRepo.FindGroupByID(txCtx, groupID)
		if err != nil {
			return err
		}
		member, err := s.expenseRepo.FindActiveMemberByUserID(txCtx, groupID, userID)
		if errors.Is(err, exception.ErrGroupMemberNotFound) {
			return exception.ErrExpenseForbidden
		}
		if err != nil {
			return err
		}
		if member.Status != enum.GroupMemberStatusActive || member.LeftAt != nil {
			return exception.ErrExpenseForbidden
		}
		if write && group.Status != enum.GroupStatusActive {
			return exception.ErrExpenseGroupInactive
		}
		return fn(txCtx, group, member)
	})
}

func (s *expenseService) withExpenseAccess(ctx context.Context, userID, groupID, expenseID int64, write bool, fn func(context.Context, *entity.Group, *entity.GroupMember, *entity.Expense) error) error {
	if expenseID <= 0 {
		return exception.ErrBadRequest
	}
	return s.withGroupAccess(ctx, userID, groupID, write, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		expense, err := s.expenseRepo.FindByID(txCtx, groupID, expenseID)
		if err != nil {
			return err
		}
		return fn(txCtx, group, member, expense)
	})
}

func (s *expenseService) validateParticipants(ctx context.Context, groupID int64, shares []*entity.ExpenseShare) error {
	ids := make([]int64, 0, len(shares))
	for _, share := range shares {
		ids = append(ids, share.MemberID)
	}
	members, err := s.expenseRepo.FindMembersByIDs(ctx, groupID, ids)
	if err != nil {
		return err
	}
	if len(members) != len(ids) {
		return exception.ErrExpenseParticipantNotFound
	}
	return nil
}

func newExpense(req dto.SaveExpenseRequest) (*entity.Expense, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" || !utf8.ValidString(title) || strings.ContainsRune(title, '\x00') || utf8.RuneCountInString(title) > 200 || req.TotalAmountToman <= 0 {
		return nil, exception.ErrBadRequest
	}
	if req.Description != nil && (!utf8.ValidString(*req.Description) || strings.ContainsRune(*req.Description, '\x00')) {
		return nil, exception.ErrBadRequest
	}
	if req.OccurredAt != nil && (req.OccurredAt.Year() < 1 || req.OccurredAt.Year() > 9999) {
		return nil, exception.ErrBadRequest
	}
	if !req.SplitType.IsValid() || len(req.Shares) == 0 {
		return nil, exception.ErrExpenseInvalidSplit
	}
	shares := make([]*entity.ExpenseShare, 0, len(req.Shares))
	seen := make(map[int64]struct{}, len(req.Shares))
	sum := new(big.Int)
	for _, input := range req.Shares {
		if input.MemberID <= 0 {
			return nil, exception.ErrBadRequest
		}
		if _, ok := seen[input.MemberID]; ok {
			return nil, exception.ErrExpenseDuplicateMembers
		}
		seen[input.MemberID] = struct{}{}
		value, err := expenseSplitValue(req.SplitType, input.Value)
		if err != nil {
			return nil, err
		}
		shares = append(shares, &entity.ExpenseShare{MemberID: input.MemberID, SplitValue: value})
		sum.Add(sum, big.NewInt(value))
	}
	sort.Slice(shares, func(i, j int) bool { return shares[i].MemberID < shares[j].MemberID })
	if req.SplitType == enum.ExpenseSplitTypeExact {
		if sum.Cmp(big.NewInt(req.TotalAmountToman)) != 0 {
			return nil, exception.ErrExpenseSplitTotalMismatch
		}
		for _, share := range shares {
			share.ShareAmountToman = share.SplitValue
		}
	} else {
		if req.SplitType == enum.ExpenseSplitTypePercentage && sum.Cmp(big.NewInt(10000)) != 0 {
			return nil, exception.ErrExpenseSplitTotalMismatch
		}
		if sum.Sign() <= 0 {
			return nil, exception.ErrExpenseInvalidSplit
		}
		type allocation struct {
			index     int
			remainder *big.Int
		}
		remainders := make([]allocation, 0, len(shares))
		var allocated int64
		for i, share := range shares {
			numerator := new(big.Int).Mul(big.NewInt(req.TotalAmountToman), big.NewInt(share.SplitValue))
			quotient, remainder := new(big.Int), new(big.Int)
			quotient.QuoRem(numerator, sum, remainder)
			share.ShareAmountToman = quotient.Int64()
			allocated += share.ShareAmountToman
			remainders = append(remainders, allocation{index: i, remainder: remainder})
		}
		sort.Slice(remainders, func(i, j int) bool {
			if comparison := remainders[i].remainder.Cmp(remainders[j].remainder); comparison != 0 {
				return comparison > 0
			}
			return shares[remainders[i].index].MemberID < shares[remainders[j].index].MemberID
		})
		for i := int64(0); i < req.TotalAmountToman-allocated; i++ {
			shares[remainders[i].index].ShareAmountToman++
		}
	}
	return &entity.Expense{Title: title, Description: req.Description, TotalAmountToman: req.TotalAmountToman, SplitType: req.SplitType, ExpenseType: enum.ExpenseTypeNormal, OccurredAt: req.OccurredAt, Shares: shares}, nil
}

func expenseSplitValue(splitType enum.ExpenseSplitType, raw []byte) (int64, error) {
	raw = bytes.TrimSpace(raw)
	if splitType == enum.ExpenseSplitTypeEqual && (len(raw) == 0 || bytes.Equal(raw, []byte("null"))) {
		return 1, nil
	}
	if len(raw) == 0 || len(raw) > 32 {
		return 0, exception.ErrExpenseInvalidSplit
	}
	if splitType == enum.ExpenseSplitTypePercentage {
		for _, c := range raw {
			if (c < '0' || c > '9') && c != '.' {
				return 0, exception.ErrExpenseInvalidSplit
			}
		}
		value, ok := new(big.Rat).SetString(string(raw))
		if !ok || value.Sign() < 0 || value.Cmp(big.NewRat(100, 1)) > 0 {
			return 0, exception.ErrExpenseInvalidSplit
		}
		value.Mul(value, big.NewRat(100, 1))
		if !value.IsInt() {
			return 0, exception.ErrExpenseInvalidSplit
		}
		return value.Num().Int64(), nil
	}
	value, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil || value < 0 {
		return 0, exception.ErrExpenseInvalidSplit
	}
	if splitType == enum.ExpenseSplitTypeEqual && value != 1 {
		return 0, exception.ErrExpenseInvalidSplit
	}
	if splitType == enum.ExpenseSplitTypeWeighted && value == 0 {
		return 0, exception.ErrExpenseInvalidSplit
	}
	return value, nil
}

func expensePagination(page, limit int) (int, int, int, error) {
	if page < 0 || limit < 0 || limit > 100 {
		return 0, 0, 0, exception.ErrBadRequest
	}
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 20
	}
	if page-1 > int(^uint(0)>>1)/limit {
		return 0, 0, 0, exception.ErrBadRequest
	}
	return page, limit, (page - 1) * limit, nil
}
