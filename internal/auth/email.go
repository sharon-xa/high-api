package auth

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/sharon-xa/high-api/internal/config"
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

func SendResetPasswordEmail(email, token string, env *config.Env) error {
	resetURL := fmt.Sprintf("%s/reset-password/confirm?token=%s", env.FrontendUrl, token)

	m := gomail.NewMessage()
	m.SetHeader("From", env.Email)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Reset Your Password")
	m.SetBody("text/html", fmt.Sprintf(
		`<p>You requested a password reset. Click the link below to reset your password:</p>
		<p><a href='%s'>Reset Password</a></p>
		<p>If you did not request this, please ignore this email.</p>`,
		resetURL,
	))

	d := gomail.NewDialer(
		"smtp.hostinger.com",
		587,
		env.Email,
		env.Password,
	)
	d.SSL = false

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // sometimes needed for self-signed certs

	if err := d.DialAndSend(m); err != nil {
		log.Println("SMTP Connection Error:", err)
		return err
	}

	return nil
}
