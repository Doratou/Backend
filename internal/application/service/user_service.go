package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

var friendCodePattern = regexp.MustCompile(`^DR-[23456789A-HJ-NP-Z]{12}$`)

type userService struct {
	userRepo          domain_db.UserRepository
	attachmentService adapter.AttachmentService
}

func NewUserService(userRepo domain_db.UserRepository, attachmentService adapter.AttachmentService) adapter.UserService {
	return &userService{userRepo: userRepo, attachmentService: attachmentService}
}

func (s *userService) GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	user, err := s.activeUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	response := dto.ToUserResponse(user)
	return &response, nil
}

func (s *userService) FindByFriendCode(ctx context.Context, userID int64, friendCode string) (*dto.UserPreviewResponse, error) {
	if _, err := s.activeUser(ctx, userID); err != nil {
		return nil, err
	}
	friendCode = strings.ToUpper(strings.TrimSpace(friendCode))
	if !friendCodePattern.MatchString(friendCode) {
		return nil, exception.ErrFriendCodeInvalid
	}
	user, err := s.userRepo.FindByFriendCode(ctx, friendCode)
	if err != nil {
		return nil, err
	}
	response := &dto.UserPreviewResponse{
		ID:                 user.ID,
		Name:               user.Name,
		AvatarAttachmentID: user.AvatarAttachmentID,
	}
	if user.AvatarAttachmentID != nil && user.AvatarAttachment != nil && user.AvatarAttachment.ObjectKey != nil {
		avatar, err := s.attachmentService.GetByID(ctx, *user.AvatarAttachmentID, &userID)
		if err != nil {
			if errors.Is(err, exception.ErrDomainAttachmentNotFound) || errors.Is(err, exception.ErrDomainAttachmentForbidden) {
				return response, nil
			}
			return nil, err
		}
		if avatar.AccessURL != "" {
			response.AvatarURL = &avatar.AccessURL
		}
	}
	return response, nil
}

func (s *userService) activeUser(ctx context.Context, userID int64) (*entity.User, error) {
	if userID <= 0 {
		return nil, exception.ErrBadRequest
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Status != enum.UserStatusActive {
		return nil, exception.ErrUserInactive
	}
	return user, nil
}
