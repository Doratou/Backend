package captcha

import (
	"github.com/mojocn/base64Captcha"
)

type CaptchaMan interface {
	Generate() (cid string, b64 string, answer string, err error)
}

type captchaMan struct {
	captcha *base64Captcha.Captcha
}

func NewStringCaptcha(cfg *DriverString) CaptchaMan {
	if cfg == nil {
		cfg = DefaultDriverString()
	}

	driver := base64Captcha.NewDriverString(
		cfg.Height,
		cfg.Width,
		cfg.NoiseCount,
		cfg.ShowLineOptions,
		cfg.Length,
		cfg.Source,
		cfg.BgColor,
		cfg.FontsStorage,
		cfg.Fonts,
	)

	return &captchaMan{
		captcha: base64Captcha.NewCaptcha(driver, &dummyStore{}),
	}
}

func NewMathCaptcha(cfg *DriverMath) CaptchaMan {
	if cfg == nil {
		cfg = DefaultDriverMath()
	}

	driver := base64Captcha.NewDriverMath(
		cfg.Height,
		cfg.Width,
		cfg.NoiseCount,
		cfg.ShowLineOptions,
		cfg.BgColor,
		cfg.FontsStorage,
		cfg.Fonts,
	)

	return &captchaMan{
		captcha: base64Captcha.NewCaptcha(driver, &dummyStore{}),
	}
}

func NewAudioCaptcha(cfg *DriverAudio) CaptchaMan {
	if cfg == nil {
		cfg = DefaultDriverAudio()
	}

	driver := base64Captcha.NewDriverAudio(
		cfg.Length,
		cfg.Language,
	)

	return &captchaMan{
		captcha: base64Captcha.NewCaptcha(driver, &dummyStore{}),
	}
}

func NewDigitCaptcha(cfg *DriverDigit) CaptchaMan {
	if cfg == nil {
		cfg = DefaultDriverDigit()
	}

	driver := base64Captcha.NewDriverDigit(
		cfg.Height,
		cfg.Width,
		cfg.Length,
		cfg.MaxSkew,
		cfg.DotCount,
	)

	return &captchaMan{
		captcha: base64Captcha.NewCaptcha(driver, &dummyStore{}),
	}
}

func (cm *captchaMan) Generate() (cid string, b64 string, answer string, err error) {
	return cm.captcha.Generate()
}
