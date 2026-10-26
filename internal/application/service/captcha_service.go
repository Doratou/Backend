package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	"github.com/Doratou/Backend/pkg/captcha"
)

// CaptchaServiceConfig encapsulates runtime tuning parameters for captcha challenges.
type CaptchaServiceConfig struct {
	TTL time.Duration
}

type captchaService struct {
	repo       cache.CaptchaRepository
	captchaMan captcha.CaptchaMan
	cfg        CaptchaServiceConfig
}

// NewCaptchaService creates an instance of CaptchaService with fallback default configuration.
func NewCaptchaService(
	repo cache.CaptchaRepository,
	captchaMan captcha.CaptchaMan,
	cfg CaptchaServiceConfig,
	consts *bootstrap.Constant,
) adapter.CaptchaService {
	if cfg.TTL <= 0 {
		if consts != nil && consts.Captcha.DefaultTTL > 0 {
			cfg.TTL = consts.Captcha.DefaultTTL
		} else {
			cfg.TTL = 2 * time.Minute
		}
		log.Printf("[CONFIG NOTICE] Captcha TTL is not configured or non-positive, applying default: %v", cfg.TTL)
	}

	if captchaMan == nil {
		captchaMan = captcha.NewStringCaptcha(nil)
	}

	return &captchaService{
		repo:       repo,
		captchaMan: captchaMan,
		cfg:        cfg,
	}
}

// Generate produces a new string captcha challenge, caches the solution, and returns the presentation DTO.
func (s *captchaService) Generate(ctx context.Context) (*dto.CaptchaResponse, error) {
	cid, b64, answer, err := s.captchaMan.Generate()
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}

	if err := s.repo.Save(ctx, cid, answer, s.cfg.TTL); err != nil {
		return nil, err
	}

	return &dto.CaptchaResponse{
		CaptchaID: cid,
		ImageB64:  b64,
		ExpiresAt: time.Now().Add(s.cfg.TTL),
	}, nil
}

// Verify validates the user submitted code and purges the key on success (single-use destruction).
func (s *captchaService) Verify(ctx context.Context, captchaID, code string) (bool, error) {
	captchaID = strings.TrimSpace(captchaID)
	code = strings.TrimSpace(code)

	if captchaID == "" || code == "" {
		return false, exception.ErrBadRequest
	}

	expectedCode, err := s.repo.Get(ctx, captchaID)
	if err != nil {
		return false, err
	}

	if !strings.EqualFold(expectedCode, code) {
		return false, exception.ErrCaptchaInvalid
	}

	// Single-use token destruction to prevent replay attacks
	_ = s.repo.Delete(ctx, captchaID)

	return true, nil
}
