package service

import (
	"context"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
	"github.com/nhassl3/IpBuild-backend/internal/repository/redis"
	"github.com/nhassl3/IpBuild-backend/pkg/auth"
	"github.com/nhassl3/IpBuild-backend/pkg/mailer"
)

// Authorization service (create/login user, token lifecycle)
type Authorization interface {
	CreateUser(ctx context.Context, user *domain.CreateUserInput) (*domain.User, *domain.TokenPair, error)
	SignIn(ctx context.Context, username, password string) (*domain.User, error)
	GenerateToken(ctx context.Context, user *domain.User) (*domain.TokenPair, error)
	ParseToken(ctx context.Context, token string) (*domain.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, token string) error
}

// Vacancies service — vacancy CRUD and applicant responses
type Vacancies interface {
	List(ctx context.Context) (*domain.Vacancies, error)
	GetVacancy(ctx context.Context, vacancyId string) (*domain.Vacancy, error)

	Create(ctx context.Context, params *domain.CreateVacancyInput) (*domain.Vacancy, error)
	Update(ctx context.Context, vacancyId string, updVacancy *domain.UpdatedVacancyInput) (*domain.Vacancy, error)
	Delete(ctx context.Context, vacancyId string) error

	Respond(ctx context.Context, vacancyId string, applicantsForm *domain.ApplicantsFormInput) error
}

// Plan service — individual plan requests
type Plan interface {
	CreatePlan(ctx context.Context, plan *domain.CreatePlanInput, userId *string) (*domain.Plan, error)
	GetPlan(ctx context.Context, planId string) (*domain.UserPlan, error)
}

type Service struct {
	Authorization
	Vacancies
	Plan
}

func NewService(
	repos *postgres.Repository,
	authRedis redis.AuthRedis,
	accessMaker auth.TokenManager,
	refreshMaker auth.TokenManager,
	blacklist auth.TokenBlacklist,
	mailer mailer.Notifier,
) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization, repos.Admin, authRedis, accessMaker, refreshMaker, blacklist),
		Vacancies:     NewVacanciesService(repos.Vacancies, mailer),
		Plan:          NewPlanService(repos.Plan, mailer),
	}
}
