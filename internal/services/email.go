// email.go: correos transaccionales con Resend. Si RESEND_API_KEY está vacío
// el envío se omite con un warning, así la app funciona sin configurarlo.
package services

import (
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v2"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/config"
)

const emailBrandColor = "#9D2C34" // burgundy Idílica (Brand Guidelines)

type EmailService struct {
	client      *resend.Client
	configured  bool
	fromEmail   string
	frontendURL string
	logger      *slog.Logger
}

func NewEmailService(cfg *config.Config, logger *slog.Logger) *EmailService {
	svc := &EmailService{
		configured:  cfg.ResendAPIKey != "",
		fromEmail:   cfg.ResendFromEmail,
		frontendURL: cfg.FrontendURL,
		logger:      logger,
	}
	if svc.configured {
		svc.client = resend.NewClient(cfg.ResendAPIKey)
	}
	return svc
}

func (s *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	resetURL := fmt.Sprintf("%s/restablecer?token=%s", s.frontendURL, resetToken)
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
			<h2 style="color: %[1]s;">Restablecer contraseña</h2>
			<p>Recibimos una solicitud para restablecer tu contraseña.</p>
			<a href="%[2]s" style="%[3]s">Restablecer contraseña</a>
			<p style="color: #666; font-size: 14px;">Este enlace expira en 1 hora.</p>
			<p style="color: #666; font-size: 14px;">Si no lo solicitaste, puedes ignorar este correo.</p>
		</div>`, emailBrandColor, resetURL, buttonStyle())

	if err := s.send(to, "Restablecer tu contraseña — Idílica", html); err != nil {
		return err
	}
	s.logger.Info("Password reset email sent", "to", to)
	return nil
}

func (s *EmailService) SendEmailVerification(to, verificationToken string) error {
	verifyURL := fmt.Sprintf("%s/verificar?token=%s", s.frontendURL, verificationToken)
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
			<h2 style="color: %[1]s;">Bienvenida a Idílica</h2>
			<p>Gracias por registrarte. Verifica tu correo electrónico:</p>
			<a href="%[2]s" style="%[3]s">Verificar correo</a>
			<p style="color: #666; font-size: 14px;">Este enlace expira en 24 horas.</p>
		</div>`, emailBrandColor, verifyURL, buttonStyle())

	if err := s.send(to, "Verifica tu correo electrónico — Idílica", html); err != nil {
		return err
	}
	s.logger.Info("Email verification sent", "to", to)
	return nil
}

func (s *EmailService) send(to, subject, html string) error {
	if !s.configured {
		s.logger.Warn("Resend not configured, skipping email", "to", to, "subject", subject)
		return nil
	}

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	})
	if err != nil {
		s.logger.Error("Failed to send email", "to", to, "error", err)
		return apperrors.NewBusinessRule("Failed to send email")
	}
	return nil
}

func buttonStyle() string {
	return fmt.Sprintf("display:inline-block;padding:12px 24px;background-color:%s;"+
		"color:#fff;text-decoration:none;border-radius:8px;margin:16px 0", emailBrandColor)
}
