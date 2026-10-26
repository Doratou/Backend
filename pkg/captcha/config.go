package captcha

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
)

type DriverString struct {
	Height          int
	Width           int
	NoiseCount      int
	ShowLineOptions int
	Length          int
	Source          string
	BgColor         *color.RGBA
	FontsStorage    base64Captcha.FontsStorage
	Fonts           []string
}

type DriverMath struct {
	Height          int
	Width           int
	NoiseCount      int
	ShowLineOptions int
	BgColor         *color.RGBA
	FontsStorage    base64Captcha.FontsStorage
	Fonts           []string
}

type DriverAudio struct {
	Length   int
	Language string
}

type DriverDigit struct {
	Height   int
	Width    int
	Length   int
	MaxSkew  float64
	DotCount int
}

func DefaultDriverDigit() *DriverDigit {
	return &DriverDigit{
		Height:   80,
		Width:    175,
		Length:   5,
		MaxSkew:  0.85,
		DotCount: 5,
	}
}

func DefaultDriverString() *DriverString {
	return &DriverString{
		Height:          60,
		Width:           245,
		NoiseCount:      70,
		ShowLineOptions: base64Captcha.OptionShowHollowLine,
		Length:          7,
		Source:          "0123456789bcdfghjklmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ",
		BgColor:         &color.RGBA{R: 93, G: 178, B: 178, A: 200},
		FontsStorage:    nil,
		Fonts:           []string{},
	}
}

func DefaultDriverMath() *DriverMath {
	return &DriverMath{
		Height:          60,
		Width:           175,
		NoiseCount:      55,
		ShowLineOptions: 2,
		BgColor:         &color.RGBA{R: 93, G: 178, B: 178, A: 200},
		FontsStorage:    nil,
		Fonts:           []string{},
	}
}

func DefaultDriverAudio() *DriverAudio {
	return &DriverAudio{
		Length:   5,
		Language: "en",
	}
}
