package services

import (
	"context"
	"fmt"

	"gopkg.in/gomail.v2"
)

type emailService struct {
	config SMTPConfig
	dialer *gomail.Dialer
}

func NewEmailService(config SMTPConfig) EmailService {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	return &emailService{
		config: config,
		dialer: dialer,
	}
}

func (s *emailService) SendConfirmationEmail(ctx context.Context, email, token string) error {
	confirmationLink := fmt.Sprintf("https://yourdomain.com/confirm?token=%s", token)

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Подтверждение аккаунта")
	m.SetBody("text/html", s.buildConfirmationTemplate(confirmationLink))

	return s.dialer.DialAndSend(m)
}

func (s *emailService) SendWelcomeEmail(ctx context.Context, email, name string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Добро пожаловать!")
	m.SetBody("text/html", s.buildWelcomeTemplate(name))

	return s.dialer.DialAndSend(m)
}

func (s *emailService) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	resetLink := fmt.Sprintf("https://yourdomain.com/reset-password?token=%s", token)

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Сброс пароля")
	m.SetBody("text/html", s.buildPasswordResetTemplate(resetLink))

	return s.dialer.DialAndSend(m)
}

func (s *emailService) buildConfirmationTemplate(link string) string {
	return fmt.Sprintf(`
        <h2>Подтверждение email</h2>
        <p>Для завершения регистрации перейдите по ссылке:</p>
        <a href="%s">Подтвердить email</a>
        <p>Ссылка действительна 24 часа</p>
    `, link)
}

func (s *emailService) buildWelcomeTemplate(link string) string {
	return fmt.Sprintf(`
        <h2>Подтверждение email</h2>
        <p>Для завершения регистрации перейдите по ссылке:</p>
        <a href="%s">Подтвердить email</a>
        <p>Ссылка действительна 24 часа</p>
    `, link)
}

func (s *emailService) buildPasswordResetTemplate(link string) string {
	return fmt.Sprintf(`
        <h2>Подтверждение email</h2>
        <p>Для завершения регистрации перейдите по ссылке:</p>
        <a href="%s">Подтвердить email</a>
        <p>Ссылка действительна 24 часа</p>
    `, link)
}
