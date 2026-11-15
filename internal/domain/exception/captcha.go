package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainCaptchaNotFound = errors.New("captcha not found or expired")
	ErrDomainCaptchaInvalid  = errors.New("captcha code is invalid")
	ErrDomainCaptchaRequired = errors.New("captcha verification is required")

	// PoW two-layer rate limiter
	ErrDomainPoWChallengeRequired = errors.New("proof of work challenge required")
	ErrDomainPoWChallengeInvalid  = errors.New("proof of work solution is invalid or expired")
)

// Captcha
var (
	ErrCaptchaNotFound = &AppError{Code: "CAPTCHA_NOT_FOUND", Message: "captcha not found or expired", HTTPStatus: http.StatusNotFound, Err: ErrDomainCaptchaNotFound}
	ErrCaptchaInvalid  = &AppError{Code: "CAPTCHA_INVALID", Message: "captcha verification failed", HTTPStatus: http.StatusBadRequest, Err: ErrDomainCaptchaInvalid}
	ErrCaptchaRequired = &AppError{Code: "CAPTCHA_REQUIRED", Message: "captcha headers X-Captcha-ID and X-Captcha-Code are required", HTTPStatus: http.StatusBadRequest, Err: ErrDomainCaptchaRequired}
)

// PoW two-layer rate limiter
//
// ErrPoWChallengeRequired is returned when a client has been blocked by the
// L2 rate limiter and must solve a PoW puzzle to continue.  The response body
// will carry a "challenge" field with everything the client needs.
var ErrPoWChallengeRequired = &AppError{
	Code:       "POW_CHALLENGE_REQUIRED",
	Message:    "rate limit exceeded, solve the proof of work challenge to continue",
	HTTPStatus: http.StatusLocked, // 423
	Err:        ErrDomainPoWChallengeRequired,
}

// ErrPoWChallengeInvalid is returned when the client supplied PoW headers
// but the solution was wrong or the challenge has expired.  A fresh
// challenge is included in the response body so the client can retry.
var ErrPoWChallengeInvalid = &AppError{
	Code:       "POW_CHALLENGE_INVALID",
	Message:    "proof of work solution is invalid or the challenge has expired",
	HTTPStatus: http.StatusUnprocessableEntity, // 422
	Err:        ErrDomainPoWChallengeInvalid,
}
