package service

import (
	"context"
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
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/Doratou/Backend/pkg/hasher"
)

type AuthServiceConfig struct {
	AccessTokenTTL time.Duration
}

type authService struct {
	userRepo   domain_db.UserRepository
	otpService adapter.OTPService
	dbTrx      database.TrxManager
	redisTrx   redisInfra.TrxManager
	jwtMaker   jwt.TokenMaker
	hasher     hasher.Hasher
	cfg        AuthServiceConfig
}

func NewAuthService(
	userRepo domain_db.UserRepository,
	otpService adapter.OTPService,
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
	return &authService{
		userRepo:   userRepo,
		otpService: otpService,
		dbTrx:      dbTrx,
		redisTrx:   redisTrx,
		jwtMaker:   jwtMaker,
		hasher:     hasher,
		cfg:        cfg,
	}
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

	user := &entity.User{
		Name:               req.Name,
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

	// Issue JWT token
	token, claims, err := s.jwtMaker.CreateToken(user.ID, user.SystemRole, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	userResponse := dto.ToUserResponse(user)
	return &dto.AuthResponse{
		User:        userResponse,
		AccessToken: token,
		ExpiresAt:   claims.ExpiresAt.Time,
	}, nil
}
