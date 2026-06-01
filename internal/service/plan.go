package service

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
	"github.com/nhassl3/IpBuild-backend/pkg/mailer"
)

type PlanService struct {
	repo   postgres.Plan
	mailer mailer.Notifier
}

func NewPlanService(repo postgres.Plan, mailer mailer.Notifier) *PlanService {
	return &PlanService{repo: repo, mailer: mailer}
}

// CreatePlan saves the plan request to the DB and asynchronously notifies the
// owner by email. SMTP errors are logged but do not fail the request.
func (s *PlanService) CreatePlan(ctx context.Context, plan *domain.CreatePlanInput) (*domain.Plan, error) {
	result, err := s.repo.CreatePlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("plan_service.CreatePlan: %w", err)
	}

	_ = s.mailer.NotifyNewPlan(ctx, plan)
	return result, nil
}

func (s *PlanService) GetPlan(ctx context.Context, planId string) (*domain.UserPlan, error) {
	result, err := s.repo.GetPlan(ctx, planId)
	if err != nil {
		return nil, fmt.Errorf("plan_service.GetPlan: %w", err)
	}
	return result, nil
}
