package service

import (
	"context"
	"errors"
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

type pollService struct {
	pollRepo domain_db.PollRepository
	userRepo domain_db.UserRepository
	dbTrx    database.TrxManager
}

func NewPollService(pollRepo domain_db.PollRepository, userRepo domain_db.UserRepository, dbTrx database.TrxManager) adapter.PollService {
	return &pollService{pollRepo: pollRepo, userRepo: userRepo, dbTrx: dbTrx}
}

func (s *pollService) Create(ctx context.Context, userID, groupID int64, req dto.CreatePollRequest) (*dto.PollResponse, error) {
	poll, err := newPoll(req)
	if err != nil {
		return nil, err
	}
	var response dto.PollResponse
	err = s.withGroupAccess(ctx, userID, groupID, true, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		poll.GroupID = group.ID
		poll.CreatedByMemberID = member.ID
		if err := s.pollRepo.CreateWithOptions(trxCtx, poll); err != nil {
			return err
		}
		response = dto.ToPollResponse(poll, nil, nil, false)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *pollService) List(ctx context.Context, userID, groupID int64, req dto.ListPollsRequest) (*dto.PollListResponse, error) {
	page, limit, offset, err := pollPagination(req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	var status *enum.PollStatus
	if req.Status != "" {
		parsed := enum.PollStatus(req.Status)
		if parsed != enum.PollStatusActive && parsed != enum.PollStatusInactive {
			return nil, exception.ErrBadRequest
		}
		status = &parsed
	}
	var response dto.PollListResponse
	err = s.withGroupAccess(ctx, userID, groupID, false, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		polls, total, err := s.pollRepo.FindByGroupID(trxCtx, group.ID, status, limit, offset)
		if err != nil {
			return err
		}
		var summaries map[int64]domain_db.PollVoteSummary
		if len(polls) > 0 {
			pollIDs := make([]int64, 0, len(polls))
			for _, poll := range polls {
				pollIDs = append(pollIDs, poll.ID)
			}
			summaries, err = s.pollRepo.FindVoteSummaries(trxCtx, group.ID, member.ID, pollIDs)
			if err != nil {
				return err
			}
		}
		items := make([]dto.PollListItem, 0, len(polls))
		for _, poll := range polls {
			summary := summaries[poll.ID]
			items = append(items, dto.ToPollListItem(poll, summary.OptionCounts, summary.MyOptionID, group.Status != enum.GroupStatusActive))
		}
		response = dto.PollListResponse{
			Items: items, Pagination: dto.PollPaginationResponse{Page: page, Limit: limit, Total: total},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *pollService) Get(ctx context.Context, userID, groupID, pollID int64) (*dto.PollResponse, error) {
	var response *dto.PollResponse
	err := s.withPollAccess(ctx, userID, groupID, pollID, false, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, poll *entity.Poll) error {
		var err error
		response, err = s.pollResponse(trxCtx, group, member, poll)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *pollService) Vote(ctx context.Context, userID, groupID, pollID int64, req dto.VotePollRequest) (*dto.PollVoteResponse, error) {
	if req.OptionID <= 0 {
		return nil, exception.ErrBadRequest
	}
	var response dto.PollVoteResponse
	err := s.withPollAccess(ctx, userID, groupID, pollID, true, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, poll *entity.Poll) error {
		if poll.Status != enum.PollStatusActive {
			return exception.ErrPollClosed
		}
		if _, err := s.pollRepo.FindOptionByID(trxCtx, poll.ID, req.OptionID); err != nil {
			return err
		}
		vote := &entity.PollVote{PollID: poll.ID, MemberID: member.ID, OptionID: req.OptionID}
		if err := s.pollRepo.UpsertVote(trxCtx, vote); err != nil {
			return err
		}
		response = dto.ToPollVoteResponse(vote)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *pollService) WithdrawVote(ctx context.Context, userID, groupID, pollID int64) error {
	return s.withPollAccess(ctx, userID, groupID, pollID, true, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, poll *entity.Poll) error {
		if poll.Status != enum.PollStatusActive {
			return exception.ErrPollClosed
		}
		return s.pollRepo.DeleteVote(trxCtx, poll.ID, member.ID)
	})
}

func (s *pollService) ListVoters(ctx context.Context, userID, groupID, pollID int64, req dto.ListPollVotersRequest) (*dto.PollVoterListResponse, error) {
	page, limit, offset, err := pollPagination(req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	var response dto.PollVoterListResponse
	err = s.withPollAccess(ctx, userID, groupID, pollID, false, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, poll *entity.Poll) error {
		if poll.Visibility != enum.PollVisibilityPublic {
			return exception.ErrPollPrivate
		}
		votes, total, err := s.pollRepo.FindVotesByPollID(trxCtx, poll.ID, limit, offset)
		if err != nil {
			return err
		}
		items := make([]dto.PollVoterResponse, 0, len(votes))
		for _, vote := range votes {
			items = append(items, dto.ToPollVoterResponse(vote))
		}
		response = dto.PollVoterListResponse{
			Items: items, Pagination: dto.PollPaginationResponse{Page: page, Limit: limit, Total: total},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *pollService) Close(ctx context.Context, userID, groupID, pollID int64) (*dto.PollStatusResponse, error) {
	var response dto.PollStatusResponse
	err := s.withPollAccess(ctx, userID, groupID, pollID, true, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, poll *entity.Poll) error {
		if poll.CreatedByMemberID != member.ID {
			return exception.ErrPollCreatorRequired
		}
		if poll.Status != enum.PollStatusActive {
			return exception.ErrPollClosed
		}
		if err := s.pollRepo.Close(trxCtx, poll, enum.PollStatusActive); err != nil {
			return err
		}
		response = dto.ToPollStatusResponse(poll)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *pollService) Delete(ctx context.Context, userID, groupID, pollID int64) error {
	return s.withPollAccess(ctx, userID, groupID, pollID, true, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember, poll *entity.Poll) error {
		if poll.CreatedByMemberID != member.ID {
			return exception.ErrPollCreatorRequired
		}
		return s.pollRepo.DeleteByID(trxCtx, group.ID, poll.ID)
	})
}

func (s *pollService) withGroupAccess(ctx context.Context, userID, groupID int64, write bool, fn func(context.Context, *entity.Group, *entity.GroupMember) error) error {
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
		group, err := s.pollRepo.FindGroupByID(trxCtx, groupID)
		if err != nil {
			return err
		}
		member, err := s.pollRepo.FindActiveMemberByUserID(trxCtx, groupID, userID)
		if errors.Is(err, exception.ErrGroupMemberNotFound) {
			return exception.ErrPollForbidden
		}
		if err != nil {
			return err
		}
		if member.Status != enum.GroupMemberStatusActive || member.LeftAt != nil {
			return exception.ErrPollForbidden
		}
		if write && group.Status != enum.GroupStatusActive {
			return exception.ErrPollGroupInactive
		}
		return fn(trxCtx, group, member)
	})
}

func (s *pollService) withPollAccess(ctx context.Context, userID, groupID, pollID int64, write bool, fn func(context.Context, *entity.Group, *entity.GroupMember, *entity.Poll) error) error {
	if pollID <= 0 {
		return exception.ErrBadRequest
	}
	return s.withGroupAccess(ctx, userID, groupID, write, func(trxCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		poll, err := s.pollRepo.FindByID(trxCtx, group.ID, pollID)
		if err != nil {
			return err
		}
		return fn(trxCtx, group, member, poll)
	})
}

func (s *pollService) pollResponse(ctx context.Context, group *entity.Group, member *entity.GroupMember, poll *entity.Poll) (*dto.PollResponse, error) {
	counts, err := s.pollRepo.CountVotesByOption(ctx, poll.ID)
	if err != nil {
		return nil, err
	}
	vote, err := s.pollRepo.FindVoteByMemberID(ctx, poll.ID, member.ID)
	if err != nil && !errors.Is(err, exception.ErrPollVoteNotFound) {
		return nil, err
	}
	response := dto.ToPollResponse(poll, counts, vote, group.Status != enum.GroupStatusActive)
	return &response, nil
}

func newPoll(req dto.CreatePollRequest) (*entity.Poll, error) {
	question := strings.TrimSpace(req.Question)
	if !validPollText(question, 500) {
		return nil, exception.ErrBadRequest
	}
	if req.Visibility != enum.PollVisibilityPublic && req.Visibility != enum.PollVisibilityPrivate {
		return nil, exception.ErrBadRequest
	}
	if len(req.Options) < 2 {
		return nil, exception.ErrPollInvalidOptions
	}
	options := make([]*entity.PollOption, 0, len(req.Options))
	seen := make(map[string]struct{}, len(req.Options))
	for _, value := range req.Options {
		title := strings.TrimSpace(value)
		if !validPollText(title, 255) {
			return nil, exception.ErrPollInvalidOptions
		}
		key := strings.ToLower(title)
		if _, exists := seen[key]; exists {
			return nil, exception.ErrPollDuplicateOptions
		}
		seen[key] = struct{}{}
		options = append(options, &entity.PollOption{Title: title})
	}
	return &entity.Poll{Question: question, Visibility: req.Visibility, Status: enum.PollStatusActive, Options: options}, nil
}

func validPollText(value string, maxLength int) bool {
	return value != "" && utf8.ValidString(value) && !strings.ContainsRune(value, '\x00') && utf8.RuneCountInString(value) <= maxLength
}

func pollPagination(page, limit int) (int, int, int, error) {
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
