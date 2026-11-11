package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

type luckyDrawService struct {
	drawRepo domain_db.LuckyDrawRepository
	userRepo domain_db.UserRepository
	dbTrx    database.TrxManager
	random   adapter.RandomGenerator
}

func NewLuckyDrawService(drawRepo domain_db.LuckyDrawRepository, userRepo domain_db.UserRepository, dbTrx database.TrxManager, random adapter.RandomGenerator) adapter.LuckyDrawService {
	return &luckyDrawService{drawRepo: drawRepo, userRepo: userRepo, dbTrx: dbTrx, random: random}
}

func (s *luckyDrawService) Create(ctx context.Context, userID, groupID int64, req dto.CreateLuckyDrawRequest) (*dto.LuckyDrawResponse, error) {
	title := strings.TrimSpace(req.Title)
	if !validLuckyDrawText(title, 200) {
		return nil, exception.ErrBadRequest
	}
	candidates, err := luckyDrawCandidates(req.Candidates)
	if err != nil {
		return nil, err
	}
	draw := &entity.LuckyDraw{Title: title, Candidates: candidates, Status: enum.LuckyDrawStatusDraft}
	err = s.withGroupAccess(ctx, userID, groupID, true, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		draw.GroupID = group.ID
		draw.CreatedByMemberID = member.ID
		return s.drawRepo.Create(trxCtx, draw)
	})
	if err != nil {
		return nil, err
	}
	response := dto.ToLuckyDrawResponse(draw, false)
	return &response, nil
}

func (s *luckyDrawService) List(ctx context.Context, userID, groupID int64, req dto.ListLuckyDrawsRequest) (*dto.PaginatedResponse[dto.LuckyDrawListItem], error) {
	page, limit, offset, err := luckyDrawPagination(req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	var status *enum.LuckyDrawStatus
	if req.Status != "" {
		if req.Status != enum.LuckyDrawStatusDraft && req.Status != enum.LuckyDrawStatusCompleted {
			return nil, exception.ErrBadRequest
		}
		status = &req.Status
	}
	var response dto.PaginatedResponse[dto.LuckyDrawListItem]
	err = s.withGroupAccess(ctx, userID, groupID, false, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		draws, total, err := s.drawRepo.FindByGroupID(trxCtx, group.ID, status, limit, offset)
		if err != nil {
			return err
		}
		items := make([]dto.LuckyDrawListItem, 0, len(draws))
		for _, draw := range draws {
			items = append(items, dto.ToLuckyDrawListItem(draw, group.Status != enum.GroupStatusActive))
		}
		response = dto.PaginatedResponse[dto.LuckyDrawListItem]{Items: items, Pagination: dto.PaginationResponse{Page: page, Limit: limit, Total: total}}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *luckyDrawService) Get(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error) {
	var response dto.LuckyDrawResponse
	err := s.withDrawAccess(ctx, userID, groupID, drawID, false, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, draw *entity.LuckyDraw) error {
		response = dto.ToLuckyDrawResponse(draw, group.Status != enum.GroupStatusActive)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *luckyDrawService) Update(ctx context.Context, userID, groupID, drawID int64, req dto.UpdateLuckyDrawRequest) (*dto.LuckyDrawResponse, error) {
	if req.Title == nil && req.Candidates == nil {
		return nil, exception.ErrBadRequest
	}
	var title *string
	if req.Title != nil {
		value := strings.TrimSpace(*req.Title)
		if !validLuckyDrawText(value, 200) {
			return nil, exception.ErrBadRequest
		}
		title = &value
	}
	var candidates []entity.LuckyDrawCandidate
	if req.Candidates != nil {
		var err error
		candidates, err = luckyDrawCandidates(*req.Candidates)
		if err != nil {
			return nil, err
		}
	}
	var response dto.LuckyDrawResponse
	err := s.withDrawAccess(ctx, userID, groupID, drawID, true, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, draw *entity.LuckyDraw) error {
		if draw.Status != enum.LuckyDrawStatusDraft {
			return exception.ErrLuckyDrawCompleted
		}
		if title != nil {
			draw.Title = *title
		}
		if req.Candidates != nil {
			draw.Candidates = candidates
		}
		if err := s.drawRepo.UpdateDraft(trxCtx, draw); err != nil {
			return err
		}
		response = dto.ToLuckyDrawResponse(draw, false)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *luckyDrawService) Run(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error) {
	var response dto.LuckyDrawResponse
	err := s.withDrawAccess(ctx, userID, groupID, drawID, true, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, draw *entity.LuckyDraw) error {
		if draw.Status == enum.LuckyDrawStatusCompleted {
			response = dto.ToLuckyDrawResponse(draw, false)
			return nil
		}
		if draw.Status != enum.LuckyDrawStatusDraft {
			return exception.Wrap(exception.ErrInternal, fmt.Errorf("unexpected lucky draw status"))
		}
		index, err := s.pickCandidate(draw.Candidates)
		if err != nil {
			return exception.Wrap(exception.ErrInternal, err)
		}
		draw.ResultIndex = &index
		if err := s.drawRepo.Complete(trxCtx, draw); err != nil {
			return err
		}
		response = dto.ToLuckyDrawResponse(draw, false)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *luckyDrawService) withGroupAccess(ctx context.Context, userID, groupID int64, write bool, fn func(context.Context, *entity.Group, *entity.GroupMember) error) error {
	if userID <= 0 || groupID <= 0 {
		return exception.ErrBadRequest
	}
	return s.dbTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		user, err := s.userRepo.FindByID(trxCtx, userID)
		if err != nil {
			return err
		}
		if user.Status != enum.UserStatusActive {
			return exception.ErrUserInactive
		}
		group, err := s.drawRepo.FindGroupByID(trxCtx, groupID)
		if err != nil {
			return err
		}
		member, err := s.drawRepo.FindActiveMemberByUserID(trxCtx, groupID, userID)
		if err != nil {
			return err
		}
		if member.Status != enum.GroupMemberStatusActive || member.LeftAt != nil {
			return exception.ErrLuckyDrawForbidden
		}
		if write && group.Status != enum.GroupStatusActive {
			return exception.ErrLuckyDrawGroupInactive
		}
		return fn(trxCtx, group, member)
	})
}

func (s *luckyDrawService) withDrawAccess(ctx context.Context, userID, groupID, drawID int64, write bool, fn func(context.Context, *entity.Group, *entity.GroupMember, *entity.LuckyDraw) error) error {
	if drawID <= 0 {
		return exception.ErrBadRequest
	}
	return s.withGroupAccess(ctx, userID, groupID, write, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		var draw *entity.LuckyDraw
		var err error
		if write {
			draw, err = s.drawRepo.FindByIDForUpdate(trxCtx, groupID, drawID)
		} else {
			draw, err = s.drawRepo.FindByID(trxCtx, groupID, drawID)
		}
		if err != nil {
			return err
		}
		if write && draw.CreatedByMemberID != member.ID {
			return exception.ErrLuckyDrawCreatorRequired
		}
		return fn(trxCtx, group, member, draw)
	})
}

func (s *luckyDrawService) pickCandidate(candidates []entity.LuckyDrawCandidate) (int, error) {
	if s.random == nil {
		return 0, fmt.Errorf("random generator is not configured")
	}
	if len(candidates) < 2 || len(candidates) > 100 {
		return 0, fmt.Errorf("invalid stored candidate count")
	}
	var total int64
	for _, candidate := range candidates {
		if candidate.Weight < 1 || candidate.Weight > 1000000 {
			return 0, fmt.Errorf("invalid stored candidate weight")
		}
		total += candidate.Weight
	}
	value, err := s.random.Int64n(total)
	if err != nil {
		return 0, err
	}
	if value < 0 || value >= total {
		return 0, fmt.Errorf("random generator returned an out-of-range value")
	}
	for i, candidate := range candidates {
		if value < candidate.Weight {
			return i, nil
		}
		value -= candidate.Weight
	}
	return 0, fmt.Errorf("could not select a lucky draw candidate")
}

func luckyDrawCandidates(input []dto.LuckyDrawCandidateInput) ([]entity.LuckyDrawCandidate, error) {
	if len(input) < 2 || len(input) > 100 {
		return nil, exception.ErrLuckyDrawInvalidCandidates
	}
	candidates := make([]entity.LuckyDrawCandidate, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, candidate := range input {
		title := strings.TrimSpace(candidate.Title)
		if !validLuckyDrawText(title, 255) || candidate.Weight < 1 || candidate.Weight > 1000000 {
			return nil, exception.ErrLuckyDrawInvalidCandidates
		}
		key := strings.ToLower(title)
		if _, exists := seen[key]; exists {
			return nil, exception.ErrLuckyDrawDuplicateCandidates
		}
		seen[key] = struct{}{}
		candidates = append(candidates, entity.LuckyDrawCandidate{Title: title, Weight: candidate.Weight})
	}
	return candidates, nil
}

func validLuckyDrawText(value string, maxLength int) bool {
	return value != "" && utf8.ValidString(value) && !strings.ContainsRune(value, '\x00') && utf8.RuneCountInString(value) <= maxLength
}

func luckyDrawPagination(page, limit int) (int, int, int, error) {
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
