package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
	"github.com/nhassl3/IpBuild-backend/pkg/mailer"
)

type VacanciesService struct {
	repo   postgres.Vacancies
	mailer mailer.Notifier
}

func NewVacanciesService(repo postgres.Vacancies, mailer mailer.Notifier) *VacanciesService {
	return &VacanciesService{repo: repo, mailer: mailer}
}

func (s *VacanciesService) List(ctx context.Context) (*domain.Vacancies, error) {
	vacancies, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.List: %w", err)
	}
	return vacancies, nil
}

func (s *VacanciesService) GetVacancy(ctx context.Context, vacancyId string) (*domain.Vacancy, error) {
	vacancy, err := s.repo.GetVacancy(ctx, vacancyId)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.GetVacancy: %w", err)
	}
	return vacancy, nil
}

func (s *VacanciesService) Create(ctx context.Context, params *domain.CreateVacancyInput) (*domain.Vacancy, error) {
	vacancy, err := s.repo.CreateVacancy(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, domain.ErrVacancyAlreadyExists
			}
		}
		return nil, fmt.Errorf("vacancies_service.Create: %w", err)
	}
	return vacancy, nil
}

func (s *VacanciesService) Update(ctx context.Context, vacancyId string, updVacancy *domain.UpdatedVacancyInput) (*domain.Vacancy, error) {
	if err := s.repo.UpdateVacancy(ctx, vacancyId, updVacancy); err != nil {
		return nil, fmt.Errorf("vacancies_service.Update: %w", err)
	}
	vacancy, err := s.repo.GetVacancy(ctx, vacancyId)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.Update: fetch after update: %w", err)
	}
	return vacancy, nil
}

func (s *VacanciesService) Delete(ctx context.Context, vacancyId string) error {
	if err := s.repo.DeleteVacancy(ctx, vacancyId); err != nil {
		return fmt.Errorf("vacancies_service.Delete: %w", err)
	}
	return nil
}

// Respond saves the applicant's form to the DB and asynchronously notifies the
// owner by email. SMTP errors are logged but do not fail the request.
func (s *VacanciesService) Respond(ctx context.Context, vacancyId string, applicantsForm *domain.ApplicantsFormInput) error {
	vacancy, err := s.repo.GetVacancy(ctx, vacancyId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrVacancyNotExists
		}
		return fmt.Errorf("vacancies_service.Respond: failed to get vacancy %w", err)
	}

	go func() {
		_ = s.mailer.NotifyNewApplicant(ctx, vacancy.Name, applicantsForm)
		_ = s.mailer.NotifyUserAboutVacancy(ctx, vacancy.Name, applicantsForm.Email)
	}()

	if _, err := s.repo.RespondToVacancy(ctx, vacancyId, applicantsForm); err != nil {
		return fmt.Errorf("vacancies_service.Respond: %w", err)
	}

	return nil
}
