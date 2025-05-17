package email

import (
	"fmt"

	"go.uber.org/zap"

	"net/smtp"

	"os"
)

// Sender defines methods for sending various types of notification emails.
type Sender interface {
	SendMail(to, subject, body string) error
	SendConfirmation(to, token string) error
	SendResetPasswordLink(to, token string) error
	SendApprovalNotification(email string) error
	SendRejectionNotification(email string, errors map[string]string) error
}

// Mailer implements the Sender interface using SMTP.
type Mailer struct {
	host       string
	port       string
	user       string
	password   string
	from       string
	projectURL string
	logger     *zap.Logger
}

// NewMailer creates a new instance of Mailer with configuration from environment variables.
func NewMailer(logger *zap.Logger) *Mailer {
	return &Mailer{
		host:       os.Getenv("SMTP_HOST"),
		port:       os.Getenv("SMTP_PORT"),
		user:       os.Getenv("SMTP_USER"),
		password:   os.Getenv("SMTP_PASSWORD"),
		from:       os.Getenv("SMTP_FROM"),
		projectURL: os.Getenv("PROJECT_URL"),
		logger:     logger,
	}
}

// SendMail sends a raw email using SMTP with the given recipient, subject, and body.
func (m *Mailer) SendMail(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.from, to, subject, body))

	var auth smtp.Auth
	if m.user != "" && m.password != "" {
		auth = smtp.PlainAuth("", m.user, m.password, m.host)
	}

	m.logger.Info("Sending email",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("smtp", addr),
		zap.String("from", m.from),
	)

	err := smtp.SendMail(addr, auth, m.from, []string{to}, msg)
	if err != nil {
		m.logger.Error("Failed to send email",
			zap.String("to", to),
			zap.Error(err),
		)
		return err
	}

	m.logger.Info("Email sent successfully",
		zap.String("to", to),
	)

	return nil
}

// SendConfirmation sends a confirmation email with a tokenized confirmation link.
func (m *Mailer) SendConfirmation(to, token string) error {
	subject := "Email Confirmation"
	link := fmt.Sprintf("%s/api/v1/auth/confirm?token=%s", m.projectURL, token)
	body := fmt.Sprintf("Click the link to confirm your email: %s", link)

	m.logger.Info("Preparing confirmation email",
		zap.String("to", to),
		zap.String("confirmation_link", link),
	)

	return m.SendMail(to, subject, body)
}

// SendResetPasswordLink sends a password reset email with a secure reset token.
func (m *Mailer) SendResetPasswordLink(to, token string) error {
	subject := "Password Reset Request"
	link := fmt.Sprintf("%s/reset-password?token=%s", m.projectURL, token)
	body := fmt.Sprintf("To reset your password, click the link below:\n\n%s\n\nIf you did not request a password reset, please ignore this email.", link)

	m.logger.Info("Preparing reset password email",
		zap.String("to", to),
		zap.String("reset_link", link),
	)

	return m.SendMail(to, subject, body)
}

// SendApprovalNotification sends an email notifying the user that their account has been approved.
func (m *Mailer) SendApprovalNotification(email string) error {
	subject := "Account approved"
	body := "Your account has been approved"

	m.logger.Info("Account approved mail was sent",
		zap.String("email", email),
	)

	return m.SendMail(email, subject, body)
}

// SendRejectionNotification sends an email informing the user that their registration was rejected, along with the reasons.
func (m *Mailer) SendRejectionNotification(email string, errors map[string]string) error {
	subject := "Registration Rejected"

	body := "Unfortunately, your registration was rejected due to the following issues:\n\n"
	for field, reason := range errors {
		body += fmt.Sprintf("- %s: %s\n", field, reason)
	}
	body += "\nPlease correct these issues and try again."

	m.logger.Info("Account rejection mail was sent",
		zap.String("email", email),
		zap.Any("rejection_errors", errors),
	)

	return m.SendMail(email, subject, body)
}
