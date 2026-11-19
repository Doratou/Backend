package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

type FriendshipServiceConfig struct {
	RequestCooldown time.Duration
}

type friendshipService struct {
	friendshipRepo domain_db.FriendshipRepository
	userRepo       domain_db.UserRepository
	dbTrx          database.TrxManager
	readTrx        database.TrxManager
	attachmentURLs adapter.AttachmentURLResolver
	cfg            FriendshipServiceConfig
}

func NewFriendshipService(
	friendshipRepo domain_db.FriendshipRepository,
	userRepo domain_db.UserRepository,
	dbTrx, readTrx database.TrxManager,
	attachmentURLs adapter.AttachmentURLResolver,
	cfg FriendshipServiceConfig,
	consts *bootstrap.Constant,
) adapter.FriendshipService {
	if cfg.RequestCooldown <= 0 {
		defaultCooldown := consts.Friendship.DefaultRequestCooldown
		log.Printf("[CONFIG NOTICE] Friendship RequestCooldown is not configured or non-positive (%v), applying default: %v", cfg.RequestCooldown, defaultCooldown)
		cfg.RequestCooldown = defaultCooldown
	}
	return &friendshipService{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
		dbTrx:          dbTrx,
		readTrx:        readTrx,
		attachmentURLs: attachmentURLs,
		cfg:            cfg,
	}
}

func (s *friendshipService) SendRequest(ctx context.Context, userID int64, req dto.CreateFriendshipRequest) (*dto.FriendRequestResponse, error) {
	if userID <= 0 || req.ReceiverUserID <= 0 {
		return nil, exception.ErrBadRequest
	}
	if userID == req.ReceiverUserID {
		return nil, exception.ErrSelfFriendRequest
	}
	if err := s.ensureActiveUser(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.ensureActiveUser(ctx, req.ReceiverUserID); err != nil {
		return nil, err
	}

	var friendship *entity.Friendship
	err := s.dbTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		existing, err := s.friendshipRepo.FindByUserPair(trxCtx, userID, req.ReceiverUserID)
		if err != nil && !errors.Is(err, exception.ErrFriendshipNotFound) {
			return err
		}
		if err == nil && !existing.DeletedAt.Valid {
			switch existing.Status {
			case enum.FriendshipStatusPending:
				return exception.ErrFriendRequestAlreadyExists
			case enum.FriendshipStatusAccepted:
				return exception.ErrAlreadyFriends
			case enum.FriendshipStatusRejected:
			default:
				return exception.ErrFriendshipInvalidStatus
			}
		}

		now := time.Now().UTC()
		if err == nil && s.requestCooldownActive(existing, userID, now) {
			return exception.ErrFriendRequestCooldown
		}

		friendship = &entity.Friendship{
			RequesterUserID: userID,
			ReceiverUserID:  req.ReceiverUserID,
			Status:          enum.FriendshipStatusPending,
			RequestedAt:     now,
		}
		if err == nil {
			friendship.BaseEntity = existing.BaseEntity
			return s.friendshipRepo.ReopenRequest(trxCtx, friendship)
		}
		return s.friendshipRepo.Create(trxCtx, friendship)
	})
	if err != nil {
		return nil, err
	}

	response := dto.ToFriendRequestResponse(friendship)
	return &response, nil
}

func (s *friendshipService) ListRequests(ctx context.Context, userID int64, req dto.ListFriendRequestsRequest) (*dto.FriendRequestListResponse, error) {
	if req.Direction != "incoming" && req.Direction != "outgoing" {
		return nil, exception.ErrBadRequest
	}
	page, limit, offset, err := friendshipPagination(req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	if err := s.ensureActiveUser(ctx, userID); err != nil {
		return nil, err
	}

	var friendships []*entity.Friendship
	var total int64
	if req.Direction == "incoming" {
		friendships, total, err = s.friendshipRepo.FindIncomingRequests(ctx, userID, limit, offset)
	} else {
		friendships, total, err = s.friendshipRepo.FindOutgoingRequests(ctx, userID, limit, offset)
	}
	if err != nil {
		return nil, err
	}

	avatarURLs, err := s.resolveAvatarURLs(ctx, userID, friendships)
	if err != nil {
		return nil, err
	}
	items := make([]dto.FriendRequestListItem, 0, len(friendships))
	for _, friendship := range friendships {
		items = append(items, dto.ToFriendRequestListItem(friendship, avatarURLs))
	}
	return &dto.FriendRequestListResponse{
		Items: items,
		Pagination: dto.FriendshipPaginationResponse{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func (s *friendshipService) AcceptRequest(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error) {
	return s.respondToRequest(ctx, userID, requestID, enum.FriendshipStatusAccepted)
}

func (s *friendshipService) RejectRequest(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error) {
	return s.respondToRequest(ctx, userID, requestID, enum.FriendshipStatusRejected)
}

func (s *friendshipService) CancelRequest(ctx context.Context, userID, requestID int64) error {
	return s.deleteFriendship(ctx, userID, requestID, enum.FriendshipStatusPending)
}

func (s *friendshipService) ListFriends(ctx context.Context, userID int64, req dto.ListFriendsRequest) (*dto.FriendListResponse, error) {
	page, limit, offset, err := friendshipPagination(req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	if err := s.ensureActiveUser(ctx, userID); err != nil {
		return nil, err
	}

	friendships, total, err := s.friendshipRepo.FindFriends(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	avatarURLs, err := s.resolveAvatarURLs(ctx, userID, friendships)
	if err != nil {
		return nil, err
	}
	items := make([]dto.FriendResponse, 0, len(friendships))
	friendIDs := make([]int64, 0, len(friendships))
	for _, friendship := range friendships {
		item := dto.ToFriendResponse(friendship, userID, avatarURLs)
		items = append(items, item)
		friendIDs = append(friendIDs, item.User.ID)
	}
	groupCounts, err := s.friendshipRepo.CountMutualGroups(ctx, userID, friendIDs)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].MutualGroupCount = groupCounts[items[i].User.ID]
	}
	return &dto.FriendListResponse{
		Items: items,
		Pagination: dto.FriendshipPaginationResponse{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func (s *friendshipService) ListMutualGroups(ctx context.Context, userID, friendshipID int64, req dto.ListFriendsRequest) (*dto.PaginatedResponse[dto.MutualGroupResponse], error) {
	page, limit, offset, err := friendshipPagination(req.Page, req.Limit)
	if err != nil || friendshipID <= 0 {
		return nil, exception.ErrBadRequest
	}

	var groups []domain_db.MutualGroup
	var total int64
	err = s.readTrx.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.ensureActiveUser(txCtx, userID); err != nil {
			return err
		}
		friendship, err := s.friendshipRepo.FindByID(txCtx, friendshipID, false)
		if err != nil {
			return err
		}

		var friendID int64
		switch userID {
		case friendship.RequesterUserID:
			friendID = friendship.ReceiverUserID
		case friendship.ReceiverUserID:
			friendID = friendship.RequesterUserID
		default:
			return exception.ErrFriendshipForbidden
		}
		if friendship.Status != enum.FriendshipStatusAccepted {
			return exception.ErrFriendshipInvalidStatus
		}

		groups, total, err = s.friendshipRepo.FindMutualGroups(txCtx, userID, friendID, limit, offset)
		return err
	})
	if err != nil {
		return nil, err
	}

	coverIDs := make([]int64, 0, len(groups))
	for _, group := range groups {
		if group.CoverAttachmentID != nil {
			coverIDs = append(coverIDs, *group.CoverAttachmentID)
		}
	}
	coverURLs, err := s.resolveAttachmentURLs(ctx, userID, coverIDs)
	if err != nil {
		return nil, err
	}

	items := make([]dto.MutualGroupResponse, 0, len(groups))
	for _, group := range groups {
		item := dto.MutualGroupResponse{ID: group.ID, Title: group.Title, MemberCount: group.MemberCount}
		if group.CoverAttachmentID != nil {
			if url := coverURLs[*group.CoverAttachmentID]; url != "" {
				item.CoverURL = &url
			}
		}
		items = append(items, item)
	}
	return &dto.PaginatedResponse[dto.MutualGroupResponse]{
		Items:      items,
		Pagination: dto.PaginationResponse{Page: page, Limit: limit, Total: total},
	}, nil
}

func (s *friendshipService) resolveAvatarURLs(ctx context.Context, userID int64, friendships []*entity.Friendship) (map[int64]string, error) {
	ids := make([]int64, 0, len(friendships))
	seen := make(map[int64]struct{})
	for _, friendship := range friendships {
		for _, user := range []*entity.User{friendship.RequesterUser, friendship.ReceiverUser} {
			if user == nil || user.Status != enum.UserStatusActive || user.AvatarAttachmentID == nil {
				continue
			}
			id := *user.AvatarAttachmentID
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	return s.resolveAttachmentURLs(ctx, userID, ids)
}

func (s *friendshipService) resolveAttachmentURLs(ctx context.Context, userID int64, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	urls, err := s.attachmentURLs.ResolveAccessURLs(ctx, userID, ids)
	var appErr *exception.AppError
	if errors.As(err, &appErr) && appErr.Code == exception.ErrStorageUnavailable.Code {
		return nil, nil
	}
	return urls, err
}

func (s *friendshipService) RemoveFriend(ctx context.Context, userID, friendshipID int64) error {
	return s.deleteFriendship(ctx, userID, friendshipID, enum.FriendshipStatusAccepted)
}

func (s *friendshipService) respondToRequest(ctx context.Context, userID, requestID int64, status enum.FriendshipStatus) (*dto.FriendRequestStatusResponse, error) {
	if requestID <= 0 {
		return nil, exception.ErrBadRequest
	}
	if err := s.ensureActiveUser(ctx, userID); err != nil {
		return nil, err
	}

	var friendship *entity.Friendship
	err := s.dbTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		var err error
		friendship, err = s.friendshipRepo.FindByID(trxCtx, requestID, true)
		if err != nil {
			return err
		}
		if friendship.ReceiverUserID != userID {
			return exception.ErrFriendshipForbidden
		}
		if friendship.Status != enum.FriendshipStatusPending {
			return exception.ErrFriendshipInvalidStatus
		}

		respondedAt := time.Now().UTC()
		friendship.Status = status
		friendship.RespondedAt = &respondedAt
		return s.friendshipRepo.Update(trxCtx, friendship, enum.FriendshipStatusPending)
	})
	if err != nil {
		return nil, err
	}

	response := dto.ToFriendRequestStatusResponse(friendship)
	return &response, nil
}

func (s *friendshipService) deleteFriendship(ctx context.Context, userID, friendshipID int64, expectedStatus enum.FriendshipStatus) error {
	if friendshipID <= 0 {
		return exception.ErrBadRequest
	}
	if err := s.ensureActiveUser(ctx, userID); err != nil {
		return err
	}

	return s.dbTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		friendship, err := s.friendshipRepo.FindByID(trxCtx, friendshipID, true)
		if err != nil {
			return err
		}
		if expectedStatus == enum.FriendshipStatusPending {
			if friendship.RequesterUserID != userID {
				return exception.ErrFriendshipForbidden
			}
		} else if friendship.RequesterUserID != userID && friendship.ReceiverUserID != userID {
			return exception.ErrFriendshipForbidden
		}
		if friendship.Status != expectedStatus {
			return exception.ErrFriendshipInvalidStatus
		}
		return s.friendshipRepo.DeleteByID(trxCtx, friendshipID, expectedStatus)
	})
}

func (s *friendshipService) ensureActiveUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return exception.ErrBadRequest
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

func (s *friendshipService) requestCooldownActive(friendship *entity.Friendship, userID int64, now time.Time) bool {
	if friendship.RequesterUserID != userID {
		return false
	}
	if friendship.Status == enum.FriendshipStatusRejected {
		rejectedAt := friendship.UpdatedAt
		if friendship.RespondedAt != nil {
			rejectedAt = *friendship.RespondedAt
		}
		return now.Before(rejectedAt.Add(s.cfg.RequestCooldown))
	}
	if friendship.Status == enum.FriendshipStatusPending && friendship.DeletedAt.Valid {
		return now.Before(friendship.RequestedAt.Add(s.cfg.RequestCooldown))
	}
	return false
}

func friendshipPagination(page, limit int) (int, int, int, error) {
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
