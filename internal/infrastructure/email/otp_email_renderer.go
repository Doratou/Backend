package email

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"io/fs"
	"net/url"
	"regexp"
	"strings"
	texttemplate "text/template"
	"unicode"

	"github.com/Doratou/Backend/internal/application/adapter"
	mail "github.com/Doratou/Backend/pkg/wneessen-go-mail"
)

//go:embed templates/otp/email.html templates/otp/email.txt templates/otp/subject.txt templates/otp/assets/email-header.png templates/otp/assets/fonts/*.woff2
var otpEmailFiles embed.FS

type otpEmailRenderer struct {
	html        *htmltemplate.Template
	text        *texttemplate.Template
	subject     string
	header      []byte
	outlook     []htmltemplate.HTML
	fontBaseURL string
}

func NewOTPEmailRenderer(fontBaseURL string) (adapter.OTPEmailRenderer, error) {
	fontBaseURL = strings.TrimSpace(fontBaseURL)
	if fontBaseURL != "" {
		fontURL, err := url.Parse(fontBaseURL)
		if err != nil || fontURL.Scheme != "https" || fontURL.Hostname() == "" || fontURL.User != nil || fontURL.RawQuery != "" || fontURL.ForceQuery || fontURL.Fragment != "" || strings.ContainsAny(fontBaseURL, "\"'<>\\") || strings.ContainsFunc(fontBaseURL, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
			return nil, fmt.Errorf("EMAIL_FONT_BASE_URL must be an absolute HTTPS URL without credentials, query or fragment")
		}
		fontBaseURL = strings.TrimRight(fontURL.String(), "/")
	}
	htmlBytes, err := otpEmailFiles.ReadFile("templates/otp/email.html")
	if err != nil {
		return nil, fmt.Errorf("read OTP HTML template: %w", err)
	}
	var outlook []htmltemplate.HTML
	outlookPattern := regexp.MustCompile("(?s)<!--\\[if mso\\]>.*?<!\\[endif\\]-->")
	htmlSource := outlookPattern.ReplaceAllStringFunc(string(htmlBytes), func(block string) string {
		outlook = append(outlook, htmltemplate.HTML(block))
		return fmt.Sprintf("{{index .Outlook %d}}", len(outlook)-1)
	})
	html, err := htmltemplate.New("email.html").Option("missingkey=error").Parse(htmlSource)
	if err != nil {
		return nil, fmt.Errorf("parse OTP HTML template: %w", err)
	}
	text, err := texttemplate.New("email.txt").Option("missingkey=error").ParseFS(otpEmailFiles, "templates/otp/email.txt")
	if err != nil {
		return nil, fmt.Errorf("parse OTP text template: %w", err)
	}
	subjectBytes, err := otpEmailFiles.ReadFile("templates/otp/subject.txt")
	if err != nil {
		return nil, fmt.Errorf("read OTP email subject: %w", err)
	}
	subject := strings.TrimSpace(string(subjectBytes))
	if subject == "" || strings.ContainsAny(subject, "\r\n") {
		return nil, fmt.Errorf("OTP email subject must be a non-empty single line")
	}
	header, err := otpEmailFiles.ReadFile("templates/otp/assets/email-header.png")
	if err != nil {
		return nil, fmt.Errorf("read OTP email header: %w", err)
	}
	return &otpEmailRenderer{html: html, text: text, subject: subject, header: header, outlook: outlook, fontBaseURL: fontBaseURL}, nil
}

func OTPEmailFontFiles() (fs.FS, error) {
	return fs.Sub(otpEmailFiles, "templates/otp/assets/fonts")
}

func (r *otpEmailRenderer) Render(code string) (mail.SendOptions, error) {
	if code == "" {
		return mail.SendOptions{}, fmt.Errorf("OTP email code must contain ASCII digits")
	}
	for _, digit := range code {
		if digit < '0' || digit > '9' {
			return mail.SendOptions{}, fmt.Errorf("OTP email code must contain ASCII digits")
		}
	}
	data := struct {
		OTPCode     string
		Outlook     []htmltemplate.HTML
		FontBaseURL string
	}{OTPCode: code, Outlook: r.outlook, FontBaseURL: r.fontBaseURL}
	var html, text bytes.Buffer
	if err := r.html.Execute(&html, data); err != nil {
		return mail.SendOptions{}, fmt.Errorf("render OTP HTML email: %w", err)
	}
	if err := r.text.Execute(&text, data); err != nil {
		return mail.SendOptions{}, fmt.Errorf("render OTP text email: %w", err)
	}
	return mail.SendOptions{
		Subject:   r.subject,
		PlainBody: text.String(),
		HTMLBody:  html.String(),
		InlineAttachments: []mail.InlineAttachment{{
			Name:      "email-header.png",
			ContentID: "dorato-email-header",
			Data:      bytes.Clone(r.header),
		}},
	}, nil
}
