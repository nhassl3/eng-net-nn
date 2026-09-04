package postgres

import (
	"context"

	"github.com/nhassl3/IpBuild-backend/internal/db"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

type Authorization interface {
	CreateUser(ctx context.Context, user domain.CreateUserInput) (*domain.User, error)
	GetUserForLogin(ctx context.Context, in *domain.SignInInput) (*domain.User, string, error)
	GetMe(ctx context.Context, params domain.GetMeParams) (*domain.User, error)
}

type Admin interface {
	IsAdmin(ctx context.Context, userID string) (bool, error)
	AddAdmin(ctx context.Context, userID string) error
}

type Vacancies interface {
	List(ctx context.Context, limit, offset int32) (*domain.VacanciesWithJd, error)
	GetVacancy(ctx context.Context, vacancyId string) (*domain.VacancyWithJd, error)

	CreateVacancy(ctx context.Context, params *domain.CreateVacancyInput) (*domain.Vacancy, error)
	UpdateVacancy(ctx context.Context, vacancyId string, updVacancy *domain.UpdatedVacancyInput) error
	DeleteVacancy(ctx context.Context, vacancyId string) error

	ListJd(ctx context.Context, limit, offset int32) (*domain.JobDirections, error)
	GetJd(ctx context.Context, jdId int64) (*domain.JobDirection, error)

	CreateJd(ctx context.Context, params *domain.CreateJobDirectionInput) (*domain.JobDirection, error)
	UpdateJd(ctx context.Context, jdId int64, params *domain.UpdateJobDirectionInput) (*domain.JobDirection, error)
	RemoveJd(ctx context.Context, jdId int64) error

	RespondToVacancy(ctx context.Context, vacancyId, objectName string, applicantsForm *domain.ApplicantsFormInput) (string, error)
	GetRespondVacancies(ctx context.Context) (*domain.RespondVacancies, error)
	GetRespondVacancy(ctx context.Context, respondVacancyId string) (*domain.RespondVacancy, error)
}

type Plan interface {
	CreatePlan(ctx context.Context, plan *domain.CreatePlanInput) (*domain.Plan, error)
	GetPlan(ctx context.Context, planId string) (*domain.UserPlan, error)
	GetDirection(ctx context.Context, directionId int32) (string, error)
	CreateLinkRequest(ctx context.Context, userId, planId string) error
	GetAllPlans(ctx context.Context) (*domain.Plans, error)
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
