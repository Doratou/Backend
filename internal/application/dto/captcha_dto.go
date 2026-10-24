package dto

import "time"

// CaptchaResponse represents the payload returned after generating a new captcha challenge.
type CaptchaResponse struct {
	CaptchaID string    `json:"captcha_id"`
	ImageB64  string    `json:"captcha_image_base64"`
	ExpiresAt time.Time `json:"expires_at"`
}

// VerifyCaptchaRequest represents the payload required to verify a captcha answer.
type VerifyCaptchaRequest struct {
	CaptchaID string `json:"captcha_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// VerifyCaptchaResponse represents the verification outcome.
type VerifyCaptchaResponse struct {
	Valid bool `json:"valid"`
}
