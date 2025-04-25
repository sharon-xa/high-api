package auth

import (
	"github.com/refine-software/high-api/internal/config"
	"gopkg.in/gomail.v2"
)

func SendVerificationEmail(userEmail, OTP string, env *config.Env) error {
	htmlBody := generateTemplate(OTP)

	m := gomail.NewMessage()
	m.SetHeader("From", env.Email)
	m.SetHeader("To", userEmail)
	m.SetHeader("Subject", "High Account Verification")
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(
		"smtp.hostinger.com",
		465,
		env.Email,
		env.Password,
	)
	d.SSL = true
	return d.DialAndSend(m)
}
