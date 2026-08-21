package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
	"github.com/nhassl3/IpBuild-backend/pkg/mailer"
)

type PlansService struct {
	repo   postgres.Plan
	mailer mailer.Notifier
}

func NewPlansService(repo postgres.Plan, mailer mailer.Notifier) *PlansService {
	return &PlansService{repo: repo, mailer: mailer}
}

// CreatePlan saves the plan request to the DB and asynchronously notifies the
// owner by email. SMTP errors are logged but do not fail the request.
func (s *PlansService) CreatePlan(ctx context.Context, plan *domain.CreatePlanInput, userId *string) (*domain.Plan, error) {
	directionName := strconv.Itoa(int(plan.Direction))
	name, err := s.repo.GetDirection(ctx, plan.Direction)
	if name == "" {
		if err != nil {
			return nil, fmt.Errorf("plan_serivce.CreatePlan: failed to load direction: %w", err)
		}
		return nil, domain.ErrDirectionNotFound
	}
	directionName = name

	go func() {
		_ = s.mailer.NotifyNewPlan(ctx, &domain.CreatePlanInputEmail{
			FullName:        plan.FullName,
			TaskDescription: plan.TaskDescription,
			Direction:       directionName,
			EmailToFeedback: plan.EmailToFeedback,
		})
		_ = s.mailer.NotifyUserAboutPlan(ctx, plan.EmailToFeedback)
	}()

	var pgErr *pgconn.PgError

	result, err := s.repo.CreatePlan(ctx, plan)
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, domain.ErrPlanRequestAlreadyExists
			}
		}
		return nil, fmt.Errorf("plan_service.CreatePlan: %w", err)
	}

	if userId != nil && *userId != "" {
		if err := s.repo.CreateLinkRequest(ctx, *userId, result.UUID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrPlanRequestNotExists
			} else if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return nil, domain.ErrPlanRequestAlreadyExists
				}
				// TODO: add new errors with new code (constraint errors)
			}
			return nil, fmt.Errorf("plan_service.CreatePlan.CreateLinkRequest: %w", err)
		}
	}

	return result, nil
}

func (s *PlansService) GetPlan(ctx context.Context, planId string) (*domain.UserPlan, error) {
	result, err := s.repo.GetPlan(ctx, planId)
	if err != nil {
		return nil, fmt.Errorf("plan_service.GetPlan: %w", err)
	}
	return result, nil
}

func (s *PlansService) GetAllPlans(ctx context.Context) (*domain.Plans, error) {
	result, err := s.repo.GetAllPlans(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPlanRequestNotExists
		}
		return nil, fmt.Errorf("plan_service.GetAllPlans: %w", err)
	}
	return result, nil
}
