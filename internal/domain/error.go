package domain

import "errors"

var (
	ErrUserNotExists            = errors.New("user not exists")
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrPlanRequestAlreadyExists = errors.New("plan request already exists")
	ErrPlanRequestNotExists     = errors.New("plan request does not exists")
	ErrVacanciesNotExists       = errors.New("vacancies not exists")
	ErrVacancyNotExists         = errors.New("vacancy not exists")
	ErrVacanciesAlreadyExists   = errors.New("vacancies already exists")
	ErrVacanciesAlreadyRespond  = errors.New("vacancies already respond")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrRedisNotFound            = errors.New("redis not found")
)
