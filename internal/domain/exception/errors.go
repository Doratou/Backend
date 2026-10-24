// Package exception defines the application's domain error vocabulary and uniform AppError envelopes.
package exception

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the standard error container throughout the application.
// Handlers translate AppErrors into JSON error envelopes without leaking internal details.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Wrap returns a new copy of sentinel with underlying err attached.
func Wrap(sentinel *AppError, err error) *AppError {
	if sentinel == nil {
		return nil
	}
	return &AppError{
		Code:       sentinel.Code,
		Message:    sentinel.Message,
		HTTPStatus: sentinel.HTTPStatus,
		Err:        err,
	}
}

// ==========================================
// 1. Raw Domain Sentinel Errors
// ==========================================
var (
	ErrDomainUserNotFound           = errors.New("user not found")
	ErrDomainUserAlreadyExists      = errors.New("user already exists")
	ErrDomainUserEmailAlreadyExists = errors.New("user email already exists")
	ErrDomainUserPhoneAlreadyExists = errors.New("user phone already exists")
	ErrDomainUserInactive           = errors.New("user account is inactive")

	ErrDomainInvalidCredentials   = errors.New("invalid credentials")
	ErrDomainTokenInvalid         = errors.New("token is invalid")
	ErrDomainTokenExpired         = errors.New("token has expired")
	ErrDomainTokenRevoked         = errors.New("token has been revoked")
	ErrDomainMissingToken         = errors.New("missing authorization token")
	ErrDomainRefreshTokenInvalid  = errors.New("invalid or expired refresh token")

	ErrDomainOTPInvalid           = errors.New("otp code invalid or expired")
	ErrDomainOTPNotFound          = errors.New("otp code not found or expired")
	ErrDomainOTPUnsupportedType   = errors.New("unsupported otp type")
	ErrDomainUnverifiedCredential = errors.New("credentials are not verified with otp")
	ErrDomainRateLimitExceeded    = errors.New("rate limit exceeded, please try again later")

	ErrDomainAttachmentNotFound   = errors.New("attachment not found")

	ErrDomainBadRequest           = errors.New("bad request")
	ErrDomainInternal             = errors.New("internal server error")
)

// ==========================================
// 2. Uniform Application Sentinel AppErrors (All messages in English)
// ==========================================
var (
	// User
	ErrUserNotFound              = &AppError{Code: "USER_NOT_FOUND", Message: "user not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainUserNotFound}
	ErrUserAlreadyExists         = &AppError{Code: "USER_ALREADY_EXISTS", Message: "user already exists", HTTPStatus: http.StatusConflict, Err: ErrDomainUserAlreadyExists}
	ErrUserEmailAlreadyExists    = &AppError{Code: "USER_EMAIL_ALREADY_EXISTS", Message: "email is already registered", HTTPStatus: http.StatusConflict, Err: ErrDomainUserEmailAlreadyExists}
	ErrUserPhoneAlreadyExists    = &AppError{Code: "USER_PHONE_ALREADY_EXISTS", Message: "phone number is already registered", HTTPStatus: http.StatusConflict, Err: ErrDomainUserPhoneAlreadyExists}
	ErrUserInactive              = &AppError{Code: "USER_INACTIVE", Message: "user account is deactivated", HTTPStatus: http.StatusForbidden, Err: ErrDomainUserInactive}

	// Auth & OTP
	ErrInvalidCredentials        = &AppError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials provided", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainInvalidCredentials}
	ErrTokenInvalid              = &AppError{Code: "TOKEN_INVALID", Message: "authorization token is invalid", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenInvalid}
	ErrTokenExpired              = &AppError{Code: "TOKEN_EXPIRED", Message: "authorization token has expired", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenExpired}
	ErrTokenRevoked              = &AppError{Code: "TOKEN_REVOKED", Message: "authorization token has been revoked", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenRevoked}
	ErrMissingToken              = &AppError{Code: "MISSING_TOKEN", Message: "authorization token is missing", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainMissingToken}
	ErrRefreshTokenInvalid       = &AppError{Code: "REFRESH_TOKEN_INVALID", Message: "refresh token is invalid or expired", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainRefreshTokenInvalid}

	ErrOTPInvalid                = &AppError{Code: "OTP_INVALID", Message: "invalid or expired verification code", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainOTPInvalid}
	ErrOTPNotFound               = &AppError{Code: "OTP_NOT_FOUND", Message: "verification code not found or expired", HTTPStatus: http.StatusNotFound, Err: ErrDomainOTPNotFound}
	ErrOTPUnsupportedType        = &AppError{Code: "OTP_UNSUPPORTED_TYPE", Message: "unsupported verification type", HTTPStatus: http.StatusBadRequest, Err: ErrDomainOTPUnsupportedType}
	ErrEmptyEmailPhone           = &AppError{Code: "EMPTY_PHONE_EMAIL", Message: "at least one of email or phone must be provided", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrUnverifiedCredentials     = &AppError{Code: "NOT_VERIFIED_CREDENTIALS", Message: "at least one credential (email or phone) must be verified via OTP", HTTPStatus: http.StatusForbidden, Err: ErrDomainUnverifiedCredential}
	ErrRateLimitExceeded         = &AppError{Code: "RATE_LIMIT_EXCEEDED", Message: "too many requests, please try again later", HTTPStatus: http.StatusTooManyRequests, Err: ErrDomainRateLimitExceeded}

	// Attachment
	ErrAttachmentNotFound        = &AppError{Code: "ATTACHMENT_NOT_FOUND", Message: "attachment not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainAttachmentNotFound}

	// General
	ErrBadRequest                = &AppError{Code: "BAD_REQUEST", Message: "bad request payload or parameters", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInternal                  = &AppError{Code: "INTERNAL_SERVER_ERROR", Message: "an internal server error occurred", HTTPStatus: http.StatusInternalServerError, Err: ErrDomainInternal}
)
