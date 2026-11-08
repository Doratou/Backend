package adapter

import mail "github.com/Doratou/Backend/pkg/wneessen-go-mail"

type OTPEmailRenderer interface {
	Render(code string) (mail.SendOptions, error)
}
