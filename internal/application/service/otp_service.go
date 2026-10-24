package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/Doratou/Backend/pkg/random"
)

type OTPServiceConfig struct {
	TTL         time.Duration
	VerifiedTTL time.Duration
	Digits      int
}

type otpService struct {
	repo      cache.OTPRepository
	redisTrx  redisInfra.TrxManager
	randomEng *random.RandomGeneratorEngine
	cfg       OTPServiceConfig
}

func NewOTPService(
	repo cache.OTPRepository,
	redisTrx redisInfra.TrxManager,
	cfg OTPServiceConfig,
	consts *bootstrap.Constant,
) adapter.OTPService {
	otpConst := consts.OTP

	if cfg.TTL <= 0 {
		log.Printf("[CONFIG NOTICE] OTP TTL is not configured or non-positive (%v), applying default: %v", cfg.TTL, otpConst.DefaultTTL)
		cfg.TTL = otpConst.DefaultTTL
	}
	if cfg.VerifiedTTL <= 0 {
		log.Printf("[CONFIG NOTICE] OTP VerifiedTTL is not configured or non-positive (%v), applying default: %v", cfg.VerifiedTTL, otpConst.DefaultVerifiedTTL)
		cfg.VerifiedTTL = otpConst.DefaultVerifiedTTL
	}
	if cfg.Digits <= 0 {
		log.Printf("[CONFIG NOTICE] OTP Digits is not configured or non-positive (%v), applying default: %d", cfg.Digits, otpConst.DefaultDigits)
		cfg.Digits = otpConst.DefaultDigits
	}

	return &otpService{
		repo:      repo,
		redisTrx:  redisTrx,
		randomEng: random.NewRandomGeneratorEngine(),
		cfg:       cfg,
	}
}

func (s *otpService) generateCode() string {
	num := s.randomEng.GeneratePositiveIntRandomNumber(s.cfg.Digits)
	return fmt.Sprintf("%0*d", s.cfg.Digits, num)
}

func (s *otpService) SendOTP(ctx context.Context, identifier string, otpType enum.OTPType) (*dto.SendOTPResponse, error) {
	if otpType != enum.OTPTypeEmail && otpType != enum.OTPTypePhone {
		return nil, exception.ErrOTPUnsupportedType
	}

	code := s.generateCode()

	if err := s.repo.SetOTP(ctx, otpType, identifier, code, s.cfg.TTL); err != nil {
		return nil, err
	}

	// In development, log the OTP code to stdout for immediate testing
	log.Printf("[OTP DEV LOG] Type: %s | Identifier: %s | Code: %s (Expires in %v)", otpType, identifier, code, s.cfg.TTL)

	return &dto.SendOTPResponse{
		ExpiresInSeconds: int64(s.cfg.TTL.Seconds()),
		Message:          fmt.Sprintf("verification code sent to %s", identifier),
	}, nil
}

func (s *otpService) VerifyOTP(ctx context.Context, identifier string, otpType enum.OTPType, code string) (*dto.VerifyOTPResponse, error) {
	if otpType != enum.OTPTypeEmail && otpType != enum.OTPTypePhone {
		return nil, exception.ErrOTPUnsupportedType
	}

	savedCode, err := s.repo.GetOTP(ctx, otpType, identifier)
	if err != nil {
		return nil, err
	}

	if savedCode != code {
		return nil, exception.ErrOTPInvalid
	}

	// Atomically delete the consumed OTP and set verified session flag via Redis TrxManager
	err = s.redisTrx.WithTransaction(ctx, func(trxCtx context.Context) error {
		if err := s.repo.DeleteOTP(trxCtx, otpType, identifier); err != nil {
			return err
		}
		return s.repo.SetVerified(trxCtx, otpType, identifier, s.cfg.VerifiedTTL)
	})
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}

	return &dto.VerifyOTPResponse{
		Status:  "verified",
		Message: "verification successful",
	}, nil
}

func (s *otpService) IsVerified(ctx context.Context, identifier string, otpType enum.OTPType) (bool, error) {
	return s.repo.IsVerified(ctx, otpType, identifier)
}

func (s *otpService) ClearVerified(ctx context.Context, identifier string, otpType enum.OTPType) error {
	return s.repo.ClearVerified(ctx, otpType, identifier)
}
