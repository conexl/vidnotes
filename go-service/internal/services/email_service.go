package services

import "context"

type EmailService interface {
	SendConfirmationEmail(ctx context.Context, email, token string) error
	SendWelcomeEmail(ctx context.Context, email, name string) error
	SendPasswordResetEmail(ctx context.Context, email, token string) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}
