package mailer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/pkg/logger/sl"
	"github.com/wneessen/go-mail"
)

type Notifier interface {
	NotifyNewApplicant(ctx context.Context, vacancyName string, form *domain.ApplicantsFormInput) error
	NotifyNewPlan(ctx context.Context, plan *domain.CreatePlanInputEmail) error
	NotifyUserAboutVacancy(ctx context.Context, vacancyName, userEmail string) error
	NotifyUserAboutPlan(ctx context.Context, userEmail string) error
	Close(ctx context.Context) error
}

// NoopNotifier logs notifications without sending emails.
// Used when SMTP host is not configured (e.g. local dev).
type NoopNotifier struct{}

func (n *NoopNotifier) NotifyUserAboutVacancy(_ context.Context, vacancyName, userEmail string) error {
	slog.Info("mailer: noop - for user new applicant", slog.String("vacancy", vacancyName), slog.String("email", userEmail))
	return nil
}

func (n *NoopNotifier) NotifyUserAboutPlan(_ context.Context, userEmail string) error {
	slog.Info("mailer: noop - for user new plan request", slog.String("email", userEmail))
	return nil
}

func (n *NoopNotifier) NotifyNewApplicant(_ context.Context, vacancyName string, form *domain.ApplicantsFormInput) error {
	slog.Info("mailer: noop — new applicant", slog.String("vacancy", vacancyName), slog.String("email", form.Email))
	return nil
}

func (n *NoopNotifier) NotifyNewPlan(_ context.Context, plan *domain.CreatePlanInputEmail) error {
	slog.Info("mailer: noop — new plan request", slog.String("name", plan.FullName), slog.String("email", plan.EmailToFeedback))
	return nil
}

func (n *NoopNotifier) Close(_ context.Context) error { return nil }

type job struct {
	subject string
	body    string
	replyTo string
	toUser  bool
}

const (
	queueSize   = 100
	numWorkers  = 2
	sendTimeout = 15 * time.Second
)

type SMTPMailer struct {
	smtpClient *mail.Client
	from       string
	ownerEmail string
	queue      chan job
	wg         sync.WaitGroup
	logger     *slog.Logger
}

func NewSMTPMailer(host, username, password, from, ownerEmail string, port int, logger *slog.Logger) (*SMTPMailer, error) {
	if from == "" {
		from = username
	}

	opts := []mail.Option{
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTimeout(10 * time.Second),
	}
	if port == 465 {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		return nil, fmt.Errorf("mailer.NewSMTPMailer: %w", err)
	}

	m := &SMTPMailer{
		smtpClient: client,
		from:       from,
		ownerEmail: ownerEmail,
		queue:      make(chan job, queueSize),
		logger:     logger,
	}

	for range numWorkers {
		m.wg.Add(1)
		go m.worker()
	}

	return m, nil
}

func (m *SMTPMailer) worker() {
	defer m.wg.Done()
	for j := range m.queue {
		ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
		if j.toUser {
			if err := m.sendToUser(ctx, j.subject, j.body, j.replyTo); err != nil {
				m.logger.Error("mailer: send to user failed", sl.ErrLog(err))
			}
		} else {
			if err := m.sendToOwner(ctx, j.subject, j.body, j.replyTo); err != nil {
				m.logger.Error("mailer: send to owner failed", sl.ErrLog(err))
			}
		}
		cancel()
	}
}

func (m *SMTPMailer) enqueue(j job) {
	select {
	case m.queue <- j:
	default:
		m.logger.Error("mailer: queue full, notification dropped",
			slog.String("subject", j.subject))
	}
}

// Close drains the worker queue within the given context deadline.
func (m *SMTPMailer) Close(ctx context.Context) error {
	close(m.queue)
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("mailer.Close: drain timed out: %w", ctx.Err())
	}
}

func (m *SMTPMailer) sendToOwner(ctx context.Context, subject, htmlBody, replyTo string) error {
	msg := mail.NewMsg()
	if err := msg.To(m.ownerEmail); err != nil {
		return fmt.Errorf("mailer.send: to: %w", err)
	}
	if replyTo != "" {
		if err := msg.ReplyTo(replyTo); err != nil {
			return fmt.Errorf("mailer.send: reply-to: %w", err)
		}
	}
	return m.send(ctx, msg, subject, htmlBody)
}

func (m *SMTPMailer) sendToUser(ctx context.Context, subject, htmlBody, userEmail string) error {
	msg := mail.NewMsg()
	if err := msg.To(userEmail); err != nil {
		return fmt.Errorf("mailer.send: to: %w", err)
	}
	return m.send(ctx, msg, subject, htmlBody)
}

func (m *SMTPMailer) send(ctx context.Context, msg *mail.Msg, subject, htmlBody string) error {
	if err := msg.From(m.from); err != nil {
		return fmt.Errorf("mailer.send: from: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, htmlBody)
	if err := m.smtpClient.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("mailer.send: %w", err)
	}
	return nil
}

func (m *SMTPMailer) NotifyNewApplicant(_ context.Context, vacancyName string, form *domain.ApplicantsFormInput) error {
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
	m.enqueue(job{subject: subject, body: body, replyTo: form.Email})
	return nil
}

func (m *SMTPMailer) NotifyNewPlan(_ context.Context, plan *domain.CreatePlanInputEmail) error {
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
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
    <tr><td style="padding:8px 12px;border:1px solid #e5e7eb;background:#f9fafb;font-weight:600">Описание задачи</td>
        <td style="padding:8px 12px;border:1px solid #e5e7eb">%s</td></tr>
  </table>
</body>
</html>`,
		plan.FullName, plan.EmailToFeedback, plan.Direction, plan.TaskDescription,
	)
	m.enqueue(job{subject: "Новая заявка на разработку плана", body: body, replyTo: plan.EmailToFeedback})
	return nil
}

func (m *SMTPMailer) NotifyUserAboutVacancy(_ context.Context, vacancyName, userEmail string) error {
	body := fmt.Sprintf(`<!DOCTYPE html>
	<html>
	<head><meta charset="UTF-8"></head>
	<body style="font-family:sans-serif;color:#222;max-width:640px;margin:0 auto">
	  <h2 style="color:#2563eb">Рассмотрение Вашего отклика началось!</h2>
	</body>
	</html>
	`)
	m.enqueue(job{subject: fmt.Sprintf("Отклик на вакансию %s", vacancyName), body: body, replyTo: userEmail, toUser: true})
	return nil
}

func (m *SMTPMailer) NotifyUserAboutPlan(_ context.Context, userEmail string) error {
	body := fmt.Sprintf(`<!DOCTYPE html>
	<html>
	<head><meta charset="UTF-8"></head>
	<body style="font-family:sans-serif;color:#222;max-width:640px;margin:0 auto">
	  <h2 style="color:#2563eb">Рассмотрение Вашего плана началось! Спасибо, что выбираете IpBuild Unet!</h2>
	</body>
	</html>
	`)
	m.enqueue(job{subject: "Рассмотрение Вашего плана", body: body, replyTo: userEmail, toUser: true})
	return nil
}
