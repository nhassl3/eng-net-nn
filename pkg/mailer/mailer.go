package mailer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/wneessen/go-mail"
)

type Notifier interface {
	NotifyNewApplicant(ctx context.Context, vacancyName string, form *domain.ApplicantsFormInput) error
	NotifyNewPlan(ctx context.Context, plan *domain.CreatePlanInput) error
}

// NoopNotifier logs incoming notifications without sending emails.
// Used when SMTP host is not configured (e.g. local dev).
type NoopNotifier struct{}

func (n *NoopNotifier) NotifyNewApplicant(_ context.Context, vacancyName string, form *domain.ApplicantsFormInput) error {
	slog.Info("mailer: noop — new applicant", slog.String("vacancy", vacancyName), slog.String("email", form.Email))
	return nil
}

func (n *NoopNotifier) NotifyNewPlan(_ context.Context, plan *domain.CreatePlanInput) error {
	slog.Info("mailer: noop — new plan request", slog.String("name", plan.FullName), slog.String("email", plan.EmailToFeedback))
	return nil
}

type SMTPMailer struct {
	smtpClient *mail.Client
	from       string
	ownerEmail string
}

func NewSMTPMailer(host, username, password, from, ownerEmail string, port int) (*SMTPMailer, error) {
	if from == "" {
		from = username
	}
	client, err := mail.NewClient(host,
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTLSPolicy(mail.DefaultPortTLS),
	)
	if err != nil {
		return nil, fmt.Errorf("mailer.NewSMTPMailer: %w", err)
	}
	return &SMTPMailer{
		smtpClient: client,
		from:       from,
		ownerEmail: ownerEmail,
	}, nil
}

func (m *SMTPMailer) send(ctx context.Context, subject, htmlBody, replyTo string) error {
	msg := mail.NewMsg()
	if err := msg.From(m.from); err != nil {
		return fmt.Errorf("mailer.send: from: %w", err)
	}
	if err := msg.To(m.ownerEmail); err != nil {
		return fmt.Errorf("mailer.send: to: %w", err)
	}
	if replyTo != "" {
		if err := msg.ReplyTo(replyTo); err != nil {
			return fmt.Errorf("mailer.send: reply-to: %w", err)
		}
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, htmlBody)
	if err := m.smtpClient.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("mailer.send: %w", err)
	}
	return nil
}

func (m *SMTPMailer) NotifyNewApplicant(ctx context.Context, vacancyName string, form *domain.ApplicantsFormInput) error {
	subject := fmt.Sprintf("Новый отклик на вакансию: %s", vacancyName)
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;color:#222;max-width:640px;margin:0 auto">
  <h2 style="color:#2563eb">Новый отклик на вакансию</h2>
  <p style="color:#6b7280">Вакансия: <strong>%s</strong></p>
  <table style="border-collapse:collapse;width:100%%">
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Имя</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Email</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Телефон</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Город</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Опыт</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">О себе</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Резюме</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb"><a href="%s" style="color:#2563eb">Открыть резюме</a></td></tr>
  </table>
</body>
</html>`,
		vacancyName,
		form.FullName, form.Email, form.PhoneNumber,
		form.City, form.Exp, form.Description, form.Resume,
	)
	return m.send(ctx, subject, body, form.Email)
}

func (m *SMTPMailer) NotifyNewPlan(ctx context.Context, plan *domain.CreatePlanInput) error {
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;color:#222;max-width:640px;margin:0 auto">
  <h2 style="color:#2563eb">Новая заявка на разработку плана</h2>
  <table style="border-collapse:collapse;width:100%%">
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Имя</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Email для связи</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Направление</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%d</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Описание задачи</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
  </table>
</body>
</html>`,
		plan.FullName, plan.EmailToFeedback, plan.Direction, plan.TaskDescription,
	)
	return m.send(ctx, "Новая заявка на разработку плана", body, plan.EmailToFeedback)
}
