package service

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
)

var friendCodePattern = regexp.MustCompile(`^DR-[23456789A-HJ-NP-Z]{12}$`)

type userService struct {
	userRepo          domain_db.UserRepository
	attachmentService adapter.AttachmentService
	otpService        adapter.OTPService
	tokenRepo         cache.AuthnTokenRepository
	jwtMaker          jwt.TokenMaker
	redisTrx          redisInfra.TrxManager
}

func NewUserService(
	userRepo domain_db.UserRepository,
	attachmentService adapter.AttachmentService,
	otpService adapter.OTPService,
	tokenRepo cache.AuthnTokenRepository,
	jwtMaker jwt.TokenMaker,
	redisTrx redisInfra.TrxManager,
) adapter.UserService {
	return &userService{
		userRepo:          userRepo,
		attachmentService: attachmentService,
		otpService:        otpService,
		tokenRepo:         tokenRepo,
		jwtMaker:          jwtMaker,
		redisTrx:          redisTrx,
	}
}

func (s *userService) GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	user, err := s.activeUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	response := dto.ToUserResponse(user)
	return &response, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	if _, err := s.activeUser(ctx, userID); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if len(name) < 2 || len(name) > 100 {
		return nil, exception.ErrBadRequest
	}

	if err := s.userRepo.UpdateName(ctx, userID, name); err != nil {
		return nil, err
	}

	return s.GetMe(ctx, userID)
}

func (s *userService) ActivateUserStatus(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	if err := s.userRepo.UpdateStatus(ctx, userID, enum.UserStatusActive); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

func (s *userService) ChangeEmail(ctx context.Context, userID int64, req dto.ChangeEmailRequest) (*dto.UserResponse, error) {
	user, err := s.activeUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return nil, exception.ErrBadRequest
	}

	if user.Email != nil && *user.Email == email && user.IsEmailVerified {
		return s.GetMe(ctx, userID)
	}

	verified, err := s.otpService.IsVerified(ctx, email, enum.OTPTypeEmail)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, exception.ErrUnverifiedCredentials
	}

	// 1. Ensure email is not already taken by another user
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, exception.ErrUserEmailAlreadyExists
	}

	// 2. Update email and mark verified
	if err := s.userRepo.UpdateEmail(ctx, userID, email); err != nil {
		return nil, err
	}

	// 3. Clear verified flag from Redis
	_ = s.redisTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		return s.otpService.ClearVerified(trxCtx, email, enum.OTPTypeEmail)
	})

	return s.GetMe(ctx, userID)
}

func (s *userService) ChangePhone(ctx context.Context, userID int64, req dto.ChangePhoneRequest) (*dto.UserResponse, error) {
	user, err := s.activeUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	phone := strings.TrimSpace(req.Phone)
	if len(phone) < 10 || len(phone) > 20 {
		return nil, exception.ErrBadRequest
	}

	if user.Phone != nil && *user.Phone == phone && user.IsPhoneVerified {
		return s.GetMe(ctx, userID)
	}

	// 1. Verify OTP was confirmed in Redis
	verified, err := s.otpService.IsVerified(ctx, phone, enum.OTPTypePhone)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, exception.ErrUnverifiedCredentials
	}

	// 2. Ensure phone is not already taken by another user
	exists, err := s.userRepo.ExistsByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, exception.ErrUserPhoneAlreadyExists
	}

	// 3. Update phone and mark verified
	if err := s.userRepo.UpdatePhone(ctx, userID, phone); err != nil {
		return nil, err
	}

	// 4. Clear verified flag from Redis
	_ = s.redisTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		return s.otpService.ClearVerified(trxCtx, phone, enum.OTPTypePhone)
	})

	return s.GetMe(ctx, userID)
}

func (s *userService) DeleteAccount(ctx context.Context, userID int64, tokenString string) error {
	if _, err := s.activeUser(ctx, userID); err != nil {
		return err
	}

	_ = s.userRepo.UpdateStatus(ctx, userID, enum.UserStatusInactive)

	if err := s.userRepo.DeleteByID(ctx, userID); err != nil {
		return err
	}

	userIDStr := strconv.FormatInt(userID, 10)
	_ = s.tokenRepo.DeleteUserRefreshSessions(ctx, userIDStr)

	claims, err := s.jwtMaker.VerifyToken(tokenString)
	if err == nil && claims != nil {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > 0 {
			_ = s.tokenRepo.BlacklistAccessToken(ctx, claims.ID, remaining)
		}
	}

	return nil
}

func (s *userService) GetWallet(ctx context.Context, userID int64) (*dto.WalletResponse, error) {
	user, err := s.activeUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.WalletResponse{Wallet: user.Wallet}, nil
}

func (s *userService) GetUserByID(ctx context.Context, currentUserID, targetUserID int64) (*dto.UserPreviewResponse, error) {
	if _, err := s.activeUser(ctx, currentUserID); err != nil {
		return nil, err
	}

	targetUser, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser.Status != enum.UserStatusActive {
		return nil, exception.ErrUserNotFound
	}

	response := &dto.UserPreviewResponse{
		ID:                 targetUser.ID,
		Name:               targetUser.Name,
		AvatarAttachmentID: targetUser.AvatarAttachmentID,
	}

	if targetUser.AvatarAttachmentID != nil && targetUser.AvatarAttachment != nil && targetUser.AvatarAttachment.ObjectKey != nil {
		avatar, err := s.attachmentService.GetByID(ctx, *targetUser.AvatarAttachmentID, &currentUserID)
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
	return s.publicProfile(ctx, userID, user)
}

func (s *userService) GetProfile(ctx context.Context, callerUserID, targetUserID int64) (*dto.UserPreviewResponse, error) {
	if targetUserID <= 0 {
		return nil, exception.ErrBadRequest
	}
	if _, err := s.activeUser(ctx, callerUserID); err != nil {
		return nil, err
	}
	user, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if user.Status != enum.UserStatusActive {
		return nil, exception.ErrUserNotFound
	}
	return s.publicProfile(ctx, callerUserID, user)
}

func (s *userService) publicProfile(ctx context.Context, userID int64, user *entity.User) (*dto.UserPreviewResponse, error) {
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

func (s *userService) RegenerateFriendCode(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	if _, err := s.activeUser(ctx, userID); err != nil {
		return nil, err
	}
	newCode, err := s.userRepo.RegenerateFriendCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.FriendCode = newCode
	response := dto.ToUserResponse(user)
	return &response, nil
}

func (s *userService) AdminListUsers(ctx context.Context, filter dto.AdminUserFilterRequest, page, limit, offset int) (*dto.PaginatedResponse[dto.UserResponse], error) {
	repoFilter := domain_db.UserFilter{
		Role:   filter.Role,
		Status: filter.Status,
		Search: filter.Search,
	}

	users, total, err := s.userRepo.FindAll(ctx, repoFilter, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]dto.UserResponse, len(users))
	for i, u := range users {
		items[i] = dto.ToUserResponse(u)
	}

	return &dto.PaginatedResponse[dto.UserResponse]{
		Items: items,
		Pagination: dto.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func (s *userService) AdminGetUserByID(ctx context.Context, targetUserID int64) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByIDUnscoped(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	response := dto.ToUserResponse(user)
	return &response, nil
}

func (s *userService) AdminUpdateUserRole(ctx context.Context, adminUserID, targetUserID int64, req dto.ChangeUserRoleRequest) (*dto.UserResponse, error) {
	if adminUserID == targetUserID {
		return nil, exception.ErrAdminSelfRoleChange
	}

	if req.Role != enum.UserRoleUser && req.Role != enum.UserRoleAdmin {
		return nil, exception.ErrBadRequest
	}

	targetUser, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	if targetUser.SystemRole == req.Role {
		res := dto.ToUserResponse(targetUser)
		return &res, nil
	}

	if err := s.userRepo.UpdateRole(ctx, targetUserID, req.Role); err != nil {
		return nil, err
	}

	targetUser.SystemRole = req.Role
	res := dto.ToUserResponse(targetUser)
	return &res, nil
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
