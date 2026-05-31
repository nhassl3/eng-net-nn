package service

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
)

type VacanciesService struct {
	repo postgres.Vacancies
}

func NewVacanciesService(repo postgres.Vacancies) *VacanciesService {
	return &VacanciesService{repo: repo}
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

func (s *VacanciesService) Respond(ctx context.Context, vacancyId string, applicantsForm *domain.ApplicantsFormInput) error {
	_, err := s.repo.RespondToVacancy(ctx, vacancyId, applicantsForm)
	if err != nil {
		return fmt.Errorf("vacancies_service.Respond: %w", err)
	}
	return nil
}
