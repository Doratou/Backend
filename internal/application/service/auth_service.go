package service

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/Doratou/Backend/pkg/hasher"
	"github.com/Doratou/Backend/pkg/random"
)

type AuthServiceConfig struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type authService struct {
	userRepo   domain_db.UserRepository
	otpService adapter.OTPService
	tokenRepo  cache.AuthnTokenRepository
	dbTrx      database.TrxManager
	redisTrx   redisInfra.TrxManager
	jwtMaker   jwt.TokenMaker
	hasher     hasher.Hasher
	randomEng  *random.RandomGeneratorEngine
	cfg        AuthServiceConfig
}

func NewAuthService(
	userRepo domain_db.UserRepository,
	otpService adapter.OTPService,
	tokenRepo cache.AuthnTokenRepository,
	dbTrx database.TrxManager,
	redisTrx redisInfra.TrxManager,
	jwtMaker jwt.TokenMaker,
	hasher hasher.Hasher,
	cfg AuthServiceConfig,
	consts *bootstrap.Constant,
) adapter.AuthService {
	if cfg.AccessTokenTTL <= 0 {
		defaultTTL := consts.Auth.DefaultAccessTokenTTL
		log.Printf("[CONFIG NOTICE] AccessTokenTTL is not configured or non-positive (%v), applying default: %v", cfg.AccessTokenTTL, defaultTTL)
		cfg.AccessTokenTTL = defaultTTL
	}
	if cfg.RefreshTokenTTL <= 0 {
		defaultTTL := consts.Auth.DefaultRefreshTokenTTL
		log.Printf("[CONFIG NOTICE] RefreshTokenTTL is not configured or non-positive (%v), applying default: %v", cfg.RefreshTokenTTL, defaultTTL)
		cfg.RefreshTokenTTL = defaultTTL
	}
	return &authService{
		userRepo:   userRepo,
		otpService: otpService,
		tokenRepo:  tokenRepo,
		dbTrx:      dbTrx,
		redisTrx:   redisTrx,
		jwtMaker:   jwtMaker,
		hasher:     hasher,
		randomEng:  random.NewRandomGeneratorEngine(),
		cfg:        cfg,
	}
}

func (s *authService) issueTokenPair(ctx context.Context, userID int64, role enum.UserRole) (accessToken string, accessClaims *jwt.Claims, refreshToken string, refreshClaims *jwt.Claims, err error) {
	accessToken, accessClaims, err = s.jwtMaker.CreateToken(userID, role, jwt.TokenTypeAccess, s.cfg.AccessTokenTTL)
	if err != nil {
		return "", nil, "", nil, err
	}

	refreshToken, refreshClaims, err = s.jwtMaker.CreateToken(userID, role, jwt.TokenTypeRefresh, s.cfg.RefreshTokenTTL)
	if err != nil {
		return "", nil, "", nil, err
	}

	userIDStr := strconv.FormatInt(userID, 10)
	if err := s.tokenRepo.SaveRefreshSession(ctx, userIDStr, refreshClaims.ID, refreshToken, s.cfg.RefreshTokenTTL); err != nil {
		return "", nil, "", nil, err
	}

	return accessToken, accessClaims, refreshToken, refreshClaims, nil
}

func (s *authService) SendOTP(ctx context.Context, req dto.SendOTPRequest) (*dto.SendOTPResponse, error) {
	return s.otpService.SendOTP(ctx, req.Identifier, req.Type)
}

func (s *authService) VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error) {
	return s.otpService.VerifyOTP(ctx, req.Identifier, req.Type, req.Code)
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	hasEmail := req.Email != nil && *req.Email != ""
	hasPhone := req.Phone != nil && *req.Phone != ""

	if !hasEmail && !hasPhone {
		return nil, exception.ErrEmptyEmailPhone
	}

	var emailVerified, phoneVerified bool
	var err error

	if hasEmail {
		emailVerified, err = s.otpService.IsVerified(ctx, *req.Email, enum.OTPTypeEmail)
		if err != nil {
			return nil, err
		}
	}

	if hasPhone {
		phoneVerified, err = s.otpService.IsVerified(ctx, *req.Phone, enum.OTPTypePhone)
		if err != nil {
			return nil, err
		}
	}

	// Condition: At least one of email or phone (or both) must be verified via OTP
	if !emailVerified && !phoneVerified {
		return nil, exception.ErrUnverifiedCredentials
	}

	// Uniqueness checks in database
	if hasEmail {
		exists, err := s.userRepo.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, exception.ErrUserEmailAlreadyExists
		}
	}

	if hasPhone {
		exists, err := s.userRepo.ExistsByPhone(ctx, *req.Phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, exception.ErrUserPhoneAlreadyExists
		}
	}

	// Password hashing
	hashedPassword, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}

	name := req.Name
	if name == "" {
		rand_name, err := s.randomEng.GenerateRandomBase58String(8)
		if err != nil {
			rand_name = "<|O__O|>"
		}
		name = rand_name
	}

	user := &entity.User{
		Name:               name,
		PasswordHash:       hashedPassword,
		Email:              req.Email,
		Phone:              req.Phone,
		IsEmailVerified:    emailVerified,
		IsPhoneVerified:    phoneVerified,
		AvatarAttachmentID: nil, // Nullable: client frontend renders default avatar placeholder
		SystemRole:         enum.UserRoleUser,
		Status:             enum.UserStatusActive,
		Wallet:             0,
		TelegramChatID:     nil, // Not collected during initial registration
		TelegramUserID:     nil,
	}

	// Atomic persistence within a PostgreSQL database transaction
	err = s.dbTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		return s.userRepo.Create(trxCtx, user)
	})
	if err != nil {
		return nil, err
	}

	// Clean up verified OTP session flags from Redis atomically using Redis TrxManager
	_ = s.redisTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		if emailVerified {
			_ = s.otpService.ClearVerified(trxCtx, *req.Email, enum.OTPTypeEmail)
		}
		if phoneVerified {
			_ = s.otpService.ClearVerified(trxCtx, *req.Phone, enum.OTPTypePhone)
		}
		return nil
	})

	// Issue token pair (access token + refresh token)
	accessToken, accessClaims, refreshToken, refreshClaims, err := s.issueTokenPair(ctx, user.ID, user.SystemRole)
	if err != nil {
		return nil, err
	}

	userResponse := dto.ToUserResponse(user)
	return &dto.AuthResponse{
		User:                  userResponse,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time,
		ExpiresAt:             accessClaims.ExpiresAt.Time,
	}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	var user *entity.User
	var err error

	// If identifier contains '@', lookup by email; otherwise by phone
	if strings.Contains(req.Identifier, "@") {
		user, err = s.userRepo.FindByEmail(ctx, req.Identifier)
	} else {
		user, err = s.userRepo.FindByPhone(ctx, req.Identifier)
	}

	if err != nil {
		if errors.Is(err, exception.ErrUserNotFound) {
			return nil, exception.ErrInvalidCredentials
		}
		return nil, err
	}

	if user.Status != enum.UserStatusActive {
		return nil, exception.ErrUserInactive
	}

	if err := s.hasher.Compare(req.Password, user.PasswordHash); err != nil {
		return nil, exception.ErrInvalidCredentials
	}

	accessToken, accessClaims, refreshToken, refreshClaims, err := s.issueTokenPair(ctx, user.ID, user.SystemRole)
	if err != nil {
		return nil, err
	}

	userResponse := dto.ToUserResponse(user)
	return &dto.AuthResponse{
		User:                  userResponse,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time,
		ExpiresAt:             accessClaims.ExpiresAt.Time,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.TokenPairResponse, error) {
	claims, err := s.jwtMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, exception.Wrap(exception.ErrRefreshTokenInvalid, err)
	}

	if claims.TokenType != jwt.TokenTypeRefresh {
		return nil, exception.ErrRefreshTokenInvalid
	}

	userIDStr := strconv.FormatInt(claims.UserID, 10)
	storedToken, err := s.tokenRepo.GetRefreshSession(ctx, userIDStr, claims.ID)
	if err != nil || storedToken != req.RefreshToken {
		return nil, exception.ErrRefreshTokenInvalid
	}

	// Rotate refresh token: purge used session before issuing new tokens
	_, _ = s.tokenRepo.DeleteRefreshSession(ctx, userIDStr, claims.ID)

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, exception.ErrUserNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, err
	}

	if user.Status != enum.UserStatusActive {
		return nil, exception.ErrUserInactive
	}

	accessToken, accessClaims, refreshToken, refreshClaims, err := s.issueTokenPair(ctx, user.ID, user.SystemRole)
	if err != nil {
		return nil, err
	}

	return &dto.TokenPairResponse{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time,
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.jwtMaker.VerifyToken(accessToken)
	if err != nil {
		return exception.Wrap(exception.ErrTokenInvalid, err)
	}

	// Blacklist the access token's jti for however long it would otherwise remain valid
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining > 0 {
		if err := s.tokenRepo.BlacklistAccessToken(ctx, claims.ID, remaining); err != nil {
			return err
		}
	}

	// Purge all refresh sessions for this user so logout terminates full session
	userIDStr := strconv.FormatInt(claims.UserID, 10)
	_ = s.tokenRepo.DeleteUserRefreshSessions(ctx, userIDStr)

	return nil
}

func (s *authService) LoginWithEmail(ctx context.Context, req dto.LoginWithEmailRequest) (*dto.AuthResponse, error) {
	verified, err := s.otpService.IsVerified(ctx, req.Email, enum.OTPTypeEmail)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, exception.ErrUnverifiedCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, exception.ErrUserNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, err
	}

	if user.Status != enum.UserStatusActive {
		return nil, exception.ErrUserInactive
	}

	if !user.IsEmailVerified {
		user.IsEmailVerified = true
		_ = s.userRepo.UpdateVerifiedStatus(ctx, user.ID, true, user.IsPhoneVerified)
	}

	// Clear verified flag atomically from Redis
	_ = s.redisTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		return s.otpService.ClearVerified(trxCtx, req.Email, enum.OTPTypeEmail)
	})

	accessToken, accessClaims, refreshToken, refreshClaims, err := s.issueTokenPair(ctx, user.ID, user.SystemRole)
	if err != nil {
		return nil, err
	}

	userResponse := dto.ToUserResponse(user)
	return &dto.AuthResponse{
		User:                  userResponse,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time,
		ExpiresAt:             accessClaims.ExpiresAt.Time,
	}, nil
}

func (s *authService) LoginWithPhone(ctx context.Context, req dto.LoginWithPhoneRequest) (*dto.AuthResponse, error) {
	verified, err := s.otpService.IsVerified(ctx, req.Phone, enum.OTPTypePhone)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, exception.ErrUnverifiedCredentials
	}

	user, err := s.userRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, exception.ErrUserNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, err
	}

	if user.Status != enum.UserStatusActive {
		return nil, exception.ErrUserInactive
	}

	if !user.IsPhoneVerified {
		user.IsPhoneVerified = true
		_ = s.userRepo.UpdateVerifiedStatus(ctx, user.ID, user.IsEmailVerified, true)
	}

	// Clear verified flag atomically from Redis
	_ = s.redisTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		return s.otpService.ClearVerified(trxCtx, req.Phone, enum.OTPTypePhone)
	})

	accessToken, accessClaims, refreshToken, refreshClaims, err := s.issueTokenPair(ctx, user.ID, user.SystemRole)
	if err != nil {
		return nil, err
	}

	userResponse := dto.ToUserResponse(user)
	return &dto.AuthResponse{
		User:                  userResponse,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time,
		ExpiresAt:             accessClaims.ExpiresAt.Time,
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID int64, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Status != enum.UserStatusActive {
		return exception.ErrUserInactive
	}

	if err := s.hasher.Compare(req.OldPassword, user.PasswordHash); err != nil {
		return exception.ErrInvalidCredentials
	}

	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}

	// Purge existing refresh sessions for security
	_ = s.tokenRepo.DeleteUserRefreshSessions(ctx, strconv.FormatInt(userID, 10))

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	verified, err := s.otpService.IsVerified(ctx, req.Identifier, req.Type)
	if err != nil {
		return err
	}
	if !verified {
		return exception.ErrUnverifiedCredentials
	}

	var user *entity.User
	if req.Type == enum.OTPTypeEmail {
		user, err = s.userRepo.FindByEmail(ctx, req.Identifier)
	} else {
		user, err = s.userRepo.FindByPhone(ctx, req.Identifier)
	}
	if err != nil {
		if errors.Is(err, exception.ErrUserNotFound) {
			return exception.ErrUserNotFound
		}
		return err
	}

	if user.Status != enum.UserStatusActive {
		return exception.ErrUserInactive
	}

	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, newHash); err != nil {
		return err
	}

	// Clear verified flag atomically from Redis
	_ = s.redisTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		return s.otpService.ClearVerified(trxCtx, req.Identifier, req.Type)
	})

	// Purge existing refresh sessions for security
	_ = s.tokenRepo.DeleteUserRefreshSessions(ctx, strconv.FormatInt(user.ID, 10))

	return nil
}
