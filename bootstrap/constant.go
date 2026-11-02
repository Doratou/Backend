package bootstrap

import "time"

type Constant struct {
	Database   DBConst
	Auth       AuthConst
	OTP        OTPConst
	Captcha    CaptchaConst
	Attachment AttachmentConst
}

type DBConst struct {
	MaxOpenDbConn int32
	MaxIdleDbConn time.Duration
	MaxDbLifeTime time.Duration
}

type AuthConst struct {
	DefaultAccessTokenTTL  time.Duration
	DefaultRefreshTokenTTL time.Duration
}

type OTPConst struct {
	DefaultTTL         time.Duration
	DefaultVerifiedTTL time.Duration
	DefaultDigits      int
}

type CaptchaConst struct {
	DefaultTTL     time.Duration
	DefaultEnabled bool
}

type AttachmentConst struct {
	MaxProfileImageSize int64
	MaxGroupCoverSize   int64
	MaxGalleryImageSize int64
	MaxGalleryVideoSize int64
	MaxReceiptSize      int64
	DefaultPresignedTTL time.Duration
}

func NewConstant() *Constant {
	return &Constant{
		Database: DBConst{
			MaxOpenDbConn: 10,
			MaxIdleDbConn: 5 * time.Minute,
			MaxDbLifeTime: 5 * time.Minute,
		},
		Auth: AuthConst{
			DefaultAccessTokenTTL:  24 * time.Hour,
			DefaultRefreshTokenTTL: 7 * 24 * time.Hour,
		},
		OTP: OTPConst{
			DefaultTTL:         3 * time.Minute,
			DefaultVerifiedTTL: 15 * time.Minute,
			DefaultDigits:      5,
		},
		Captcha: CaptchaConst{
			DefaultTTL:     2 * time.Minute,
			DefaultEnabled: true,
		},
		Attachment: AttachmentConst{
			MaxProfileImageSize: 5 * 1024 * 1024,   // 5 MB
			MaxGroupCoverSize:   10 * 1024 * 1024,  // 10 MB
			MaxGalleryImageSize: 20 * 1024 * 1024,  // 20 MB
			MaxGalleryVideoSize: 100 * 1024 * 1024, // 100 MB
			MaxReceiptSize:      10 * 1024 * 1024,  // 10 MB
			DefaultPresignedTTL: 15 * time.Minute,
		},
	}
}
