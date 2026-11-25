package service

import (
	"context"
	"fmt"
	"log"
	netmail "net/mail"
	"strings"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/Doratou/Backend/pkg/random"
	mail "github.com/Doratou/Backend/pkg/wneessen-go-mail"
)

type OTPServiceConfig struct {
	TTL                  time.Duration
	VerifiedTTL          time.Duration
	Digits               int
	EmailDeliveryEnabled bool
	EmailDeliveryMode    string
}

type otpService struct {
	repo            cache.OTPRepository
	redisTrx        redisInfra.TrxManager
	randomEng       *random.RandomGeneratorEngine
	syncEmailSender mail.Sender
	emailRenderer   adapter.OTPEmailRenderer
	brokerPub       domain_broker.Publisher
	cfg             OTPServiceConfig
}

func NewOTPService(
	repo cache.OTPRepository,
	redisTrx redisInfra.TrxManager,
	syncEmailSender mail.Sender,
	emailRenderer adapter.OTPEmailRenderer,
	brokerPub domain_broker.Publisher,
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
		repo:            repo,
		redisTrx:        redisTrx,
		randomEng:       random.NewRandomGeneratorEngine(),
		syncEmailSender: syncEmailSender,
		emailRenderer:   emailRenderer,
		brokerPub:       brokerPub,
		cfg:             cfg,
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

	if otpType == enum.OTPTypeEmail {
		address, err := netmail.ParseAddress(identifier)
		if err != nil || address.Address != identifier {
			return nil, exception.ErrBadRequest
		}
	}
	if otpType == enum.OTPTypeEmail && s.cfg.EmailDeliveryEnabled {
		if !strings.EqualFold(strings.TrimSpace(s.cfg.EmailDeliveryMode), "SMTP") {
			return nil, exception.Wrap(exception.ErrInternal, fmt.Errorf("EMAIL_DELIVERY_MODE must be SMTP"))
		}
		if s.brokerPub == nil && s.syncEmailSender == nil {
			return nil, exception.Wrap(exception.ErrInternal, fmt.Errorf("neither broker publisher nor email sender is configured"))
		}
		if s.emailRenderer == nil {
			return nil, exception.Wrap(exception.ErrInternal, fmt.Errorf("OTP email renderer is not configured"))
		}
	}

	code := s.generateCode()

	if err := s.repo.SetOTP(ctx, otpType, identifier, code, s.cfg.TTL); err != nil {
		return nil, err
	}

	if otpType == enum.OTPTypeEmail && s.cfg.EmailDeliveryEnabled {
		message, err := s.emailRenderer.Render(code)
		if err != nil {
			log.Printf("failed to render OTP email: %v", err)
			return nil, exception.Wrap(exception.ErrInternal, err)
		}
		message.To = []string{identifier}

		// Asynchronous delivery via RabbitMQ broker if publisher is available
		if s.brokerPub != nil {
			attachments := make([]payload.InlineAttachment, len(message.InlineAttachments))
			for i, att := range message.InlineAttachments {
				attachments[i] = payload.InlineAttachment{
					Name:      att.Name,
					ContentID: att.ContentID,
					Data:      att.Data,
				}
			}
			emailPayload := payload.EmailPayload{
				To:                message.To,
				Subject:           message.Subject,
				PlainBody:         message.PlainBody,
				HTMLBody:          message.HTMLBody,
				InlineAttachments: attachments,
			}
			if err := s.brokerPub.PublishJSON(ctx, domain_broker.NotificationExchange, domain_broker.EmailRoutingKey, emailPayload); err != nil {
				log.Printf("failed to publish OTP email to rabbitmq: %v", err)
				return nil, exception.Wrap(exception.ErrInternal, err)
			}
		} else if s.syncEmailSender != nil {
			// Synchronous fallback if broker publisher is not configured
			if err := s.syncEmailSender.Send(ctx, message); err != nil {
				log.Printf("failed to send OTP email: %v", err)
				return nil, exception.Wrap(exception.ErrInternal, err)
			}
		}
	} else if otpType == enum.OTPTypePhone {
		// Asynchronous delivery for SMS queue
		smsPayload := payload.SMSPayload{
			Recipient:  identifier,
			Message:    fmt.Sprintf("Your Doratou verification code: %s", code),
			TemplateID: "otp_verify",
			Tokens:     map[string]string{"code": code},
		}
		if s.brokerPub != nil {
			if err := s.brokerPub.PublishJSON(ctx, domain_broker.NotificationExchange, domain_broker.SMSRoutingKey, smsPayload); err != nil {
				log.Printf("failed to publish OTP sms to rabbitmq: %v", err)
				return nil, exception.Wrap(exception.ErrInternal, err)
			} else {
				// use syncSmsSender
			}
		}
	} else {
		log.Printf("[OTP DEV LOG] Type: %s | Identifier: %s | Code: %s (Expires in %v)", otpType, identifier, code, s.cfg.TTL)
	}

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
