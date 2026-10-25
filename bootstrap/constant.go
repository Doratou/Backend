package bootstrap

import "time"

type Constant struct {
	Database DBConst
	Auth     AuthConst
	OTP      OTPConst
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
	}
}
