package postgres

import (
	"context"

	"github.com/nhassl3/IpBuild-backend/internal/db"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

type Authorization interface {
	CreateUser(ctx context.Context, user domain.CreateUserInput) (*domain.User, error)
	GetUserForLogin(ctx context.Context, username string) (*domain.User, string, error)
	GetMe(ctx context.Context, params domain.GetMeParams) (*domain.User, error)
}

type Admin interface {
	IsAdmin(ctx context.Context, userID string) (bool, error)
	AddAdmin(ctx context.Context, userID string) error
}

type Vacancies interface {
	List(ctx context.Context) (*domain.Vacancies, error)
	GetVacancy(ctx context.Context, vacancyId string) (*domain.Vacancy, error)

	CreateVacancy(ctx context.Context, params *domain.CreateVacancyInput) (*domain.Vacancy, error)
	UpdateVacancy(ctx context.Context, vacancyId string, updVacancy *domain.UpdatedVacancyInput) error
	DeleteVacancy(ctx context.Context, vacancyId string) error
	
	RespondToVacancy(ctx context.Context, vacancyId string, applicantsForm *domain.ApplicantsFormInput) (string, error)
}

type Plan interface {
	CreatePlan(ctx context.Context, plan *domain.CreatePlanInput) (*domain.Plan, error)
	GetPlan(ctx context.Context, planId string) (*domain.UserPlan, error)
}

type Repository struct {
	Authorization
	Admin
	Vacancies
	Plan
}

func NewRepository(db *db.Store) *Repository {
	return &Repository{
		Authorization: NewAuthRepo(db),
		Admin:         NewAdminRepo(db),
		Vacancies:     NewVacanciesRepo(db),
		Plan:          NewPlanRepo(db),
	}
}
