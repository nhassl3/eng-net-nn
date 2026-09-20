package mailer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
	"github.com/wneessen/go-mail"
)

type Notifier interface {
	NotifyNewApplicant(ctx context.Context, vacancyName, resumeUrl string, form *domain.ApplicantsFormInput) error
	NotifyNewPlan(ctx context.Context, plan *domain.CreatePlanInputEmail) error
	NotifyUserAboutVacancy(ctx context.Context, vacancyName, userEmail string) error
	NotifyUserAboutPlan(ctx context.Context, userEmail string) error
	NotifyUserAboutCode(ctx context.Context, code string) error
	NotifyUserAboutResponseToPlan(ctx context.Context, userEmail, message string) error
	Close(ctx context.Context) error
}

type NewApplicantsFormInput struct {
	VacancyName,
	Name,
	Email,
	Phone,
	City,
	Experience,
	AboutSelf,
	Link string
}

type NewPlanFormInput struct {
	Name,
	Email,
	Direction,
	Description string
}

// NoopNotifier logs notifications without sending emails.
// Used when SMTP host is not configured (e.g. local dev).
type NoopNotifier struct {
	log logger.Logger
}

func NewNoopNotifier(log logger.Logger) *NoopNotifier {
	return &NoopNotifier{log: log}
}

func (n *NoopNotifier) NotifyUserAboutVacancy(_ context.Context, vacancyName, userEmail string) error {
	n.log.Info("noop: notify user about vacancy", logger.String("vacancy", vacancyName), logger.Email("email", userEmail))
	return nil
}

func (n *NoopNotifier) NotifyUserAboutPlan(_ context.Context, userEmail string) error {
	n.log.Info("noop: notify user about plan", logger.Email("email", userEmail))
	return nil
}

func (n *NoopNotifier) NotifyNewApplicant(_ context.Context, vacancyName, resumeUrl string, form *domain.ApplicantsFormInput) error {
	n.log.Info("noop: notify new applicant", logger.String("vacancy", vacancyName), logger.Email("email", form.Email))
	return nil
}

func (n *NoopNotifier) NotifyNewPlan(_ context.Context, plan *domain.CreatePlanInputEmail) error {
	n.log.Info("noop: notify new plan", logger.String("name", plan.FullName), logger.Email("email", plan.EmailToFeedback))
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
	queueSize      = 100
	numWorkers     = 2
	sendTimeout    = 15 * time.Second
	enqueueTimeout = 2 * time.Second
)

type SMTPMailer struct {
	smtpClient *mail.Client
	from       string
	ownerEmail string
	queue      chan job
	closeCh    chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
	log        logger.Logger
}

func NewSMTPMailer(host, username, password, from, ownerEmail string, port int, log logger.Logger) (*SMTPMailer, error) {
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
		closeCh:    make(chan struct{}),
		log:        log,
	}

	for range numWorkers {
		m.wg.Add(1)
		go m.worker()
	}

	return m, nil
}

func (m *SMTPMailer) worker() {
	defer m.wg.Done()
	for {
		select {
		case j := <-m.queue:
			m.handle(j)
		case <-m.closeCh:
			// Drain whatever is already queued, then exit. The queue channel
			// itself is never closed, so this is the only shutdown signal.
			for {
				select {
				case j := <-m.queue:
					m.handle(j)
				default:
					return
				}
			}
		}
	}
}

func (m *SMTPMailer) handle(j job) {
	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	if j.toUser {
		if err := m.sendToUser(ctx, j.subject, j.body, j.replyTo); err != nil {
			m.log.Error("send to user failed", logger.Op("SMTPMailer.worker"), logger.Err(err))
		} else {
			m.log.Info("email sent", logger.Op("SMTPMailer.worker"), logger.String("subject", j.subject), logger.Bool("to_user", true))
		}
	} else {
		if err := m.sendToOwner(ctx, j.subject, j.body, j.replyTo); err != nil {
			m.log.Error("send to owner failed", logger.Op("SMTPMailer.worker"), logger.Err(err))
		} else {
			m.log.Info("email sent", logger.Op("SMTPMailer.worker"), logger.String("subject", j.subject), logger.Bool("to_user", false))
		}
	}
}

// enqueue submits j to the worker queue, waiting for space up to
// enqueueTimeout (or until ctx is done, if sooner). It fails fast if the
// mailer is already closed. This bounds how long a request handler can be
// blocked by a full queue, since Notify* is called synchronously.
func (m *SMTPMailer) enqueue(ctx context.Context, j job) error {
	ctx, cancel := context.WithTimeout(ctx, enqueueTimeout)
	defer cancel()

	select {
	case m.queue <- j:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("mailer.enqueue: %w", ctx.Err())
	case <-m.closeCh:
		return fmt.Errorf("mailer.enqueue: mailer is closed")
	}
}

// Close stops accepting new work and drains the worker queue within the
// given context deadline. The queue channel is never closed directly, so a
// concurrent enqueue can never panic on a send to a closed channel.
func (m *SMTPMailer) Close(ctx context.Context) error {
	m.closeOnce.Do(func() { close(m.closeCh) })
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

func (m *SMTPMailer) NotifyNewApplicant(ctx context.Context, vacancyName string, resumeUrl string, form *domain.ApplicantsFormInput) error {
	subject := fmt.Sprintf("Новый отклик на вакансию: %s", vacancyName)
	body, err := Render(NewApplicant, NewApplicantsFormInput{
		VacancyName: vacancyName,
		Name:        form.FullName,
		Email:       form.Email,
		Phone:       form.PhoneNumber,
		City:        form.City,
		Experience:  form.Exp,
		AboutSelf:   form.Description,
		Link:        resumeUrl,
	})
	if err != nil {
		return fmt.Errorf("mailer.NotifyNewApplicant: %w", err)
	}
	return m.enqueue(ctx, job{subject: subject, body: body, replyTo: form.Email})
}

func (m *SMTPMailer) NotifyNewPlan(ctx context.Context, plan *domain.CreatePlanInputEmail) error {
	body, err := Render(NewPlan, NewPlanFormInput{
		Name:        plan.FullName,
		Email:       plan.EmailToFeedback,
		Direction:   plan.Direction,
		Description: plan.TaskDescription,
	})
	if err != nil {
		return fmt.Errorf("mailer.NotifyNewPlan: %w", err)
	}
	return m.enqueue(ctx, job{subject: "Новая заявка на разработку плана", body: body, replyTo: plan.EmailToFeedback})
}

func (m *SMTPMailer) NotifyUserAboutVacancy(ctx context.Context, vacancyName, userEmail string) error {
	body, err := Render(NotifyAboutVacancy, nil)
	if err != nil {
		return fmt.Errorf("mailer.NotifyUserAboutVacancy: %w", err)
	}
	return m.enqueue(ctx, job{subject: fmt.Sprintf("Отклик на вакансию %s", vacancyName), body: body, replyTo: userEmail, toUser: true})
}

func (m *SMTPMailer) NotifyUserAboutPlan(ctx context.Context, userEmail string) error {
	body, err := Render(NotifyAboutPlan, nil)
	if err != nil {
		return fmt.Errorf("mailer.NotifyUserAboutPlan: %w", err)
	}
	return m.enqueue(ctx, job{subject: "Рассмотрение Вашего плана", body: body, replyTo: userEmail, toUser: true})
}

func (m *SMTPMailer) NotifyUserAboutCode(ctx context.Context, code string) error {
	return nil
}

func (m *SMTPMailer) NotifyUserAboutResponseToPlan(ctx context.Context, userEmail, message string) error {
	return nil
}
