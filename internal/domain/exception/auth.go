package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainInvalidCredentials   = errors.New("invalid credentials")
	ErrDomainTokenInvalid         = errors.New("token is invalid")
	ErrDomainTokenExpired         = errors.New("token has expired")
	ErrDomainTokenRevoked         = errors.New("token has been revoked")
	ErrDomainTokenNotFound        = errors.New("token session not found or revoked")
	ErrDomainMissingToken         = errors.New("missing authorization token")
	ErrDomainRefreshTokenInvalid  = errors.New("invalid or expired refresh token")
	ErrDomainOAuthStateInvalid    = errors.New("oauth state is invalid or expired")
	ErrDomainOAuthOnlyAccount     = errors.New("this account uses social login; set a password in profile settings")
	ErrDomainOAuthEmailUnverified = errors.New("google account email is not verified")

	ErrDomainOTPInvalid           = errors.New("otp code invalid or expired")
	ErrDomainOTPNotFound          = errors.New("otp code not found or expired")
	ErrDomainOTPUnsupportedType   = errors.New("unsupported otp type")
	ErrDomainUnverifiedCredential = errors.New("credentials are not verified with otp")
	ErrDomainRateLimitExceeded    = errors.New("rate limit exceeded, please try again later")
)

// Auth & OTP
var (
	ErrInvalidCredentials  = &AppError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials provided", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainInvalidCredentials}
	ErrTokenInvalid        = &AppError{Code: "TOKEN_INVALID", Message: "authorization token is invalid", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenInvalid}
	ErrTokenExpired        = &AppError{Code: "TOKEN_EXPIRED", Message: "authorization token has expired", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenExpired}
	ErrTokenRevoked        = &AppError{Code: "TOKEN_REVOKED", Message: "authorization token has been revoked", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenRevoked}
	ErrTokenNotFound       = &AppError{Code: "TOKEN_NOT_FOUND", Message: "token session not found or revoked", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainTokenNotFound}
	ErrMissingToken        = &AppError{Code: "MISSING_TOKEN", Message: "authorization token is missing", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainMissingToken}
	ErrRefreshTokenInvalid = &AppError{Code: "REFRESH_TOKEN_INVALID", Message: "refresh token is invalid or expired", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainRefreshTokenInvalid}

	ErrOTPInvalid            = &AppError{Code: "OTP_INVALID", Message: "invalid or expired verification code", HTTPStatus: http.StatusUnauthorized, Err: ErrDomainOTPInvalid}
	ErrOTPNotFound           = &AppError{Code: "OTP_NOT_FOUND", Message: "verification code not found or expired", HTTPStatus: http.StatusNotFound, Err: ErrDomainOTPNotFound}
	ErrOTPUnsupportedType    = &AppError{Code: "OTP_UNSUPPORTED_TYPE", Message: "unsupported verification type", HTTPStatus: http.StatusBadRequest, Err: ErrDomainOTPUnsupportedType}
	ErrEmptyEmailPhone       = &AppError{Code: "EMPTY_PHONE_EMAIL", Message: "at least one of email or phone must be provided", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrUnverifiedCredentials = &AppError{Code: "NOT_VERIFIED_CREDENTIALS", Message: "at least one credential (email or phone) must be verified via OTP", HTTPStatus: http.StatusForbidden, Err: ErrDomainUnverifiedCredential}
	ErrOAuthStateInvalid     = &AppError{Code: "OAUTH_STATE_INVALID", Message: "oauth state is missing or expired", HTTPStatus: http.StatusBadRequest, Err: ErrDomainOAuthStateInvalid}
	ErrOAuthOnlyAccount      = &AppError{Code: "OAUTH_ONLY_ACCOUNT", Message: "this account uses social login; set a password in profile settings", HTTPStatus: http.StatusBadRequest, Err: ErrDomainOAuthOnlyAccount}
	ErrOAuthEmailUnverified  = &AppError{Code: "OAUTH_EMAIL_UNVERIFIED", Message: "google account email is not verified", HTTPStatus: http.StatusForbidden, Err: ErrDomainOAuthEmailUnverified}
	ErrRateLimitExceeded     = &AppError{Code: "RATE_LIMIT_EXCEEDED", Message: "too many requests, please try again later", HTTPStatus: http.StatusTooManyRequests, Err: ErrDomainRateLimitExceeded}
)
