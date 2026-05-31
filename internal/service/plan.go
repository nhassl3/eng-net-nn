package service

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
)

type PlanService struct {
	repo postgres.Plan
}

func NewPlanService(repo postgres.Plan) *PlanService {
	return &PlanService{repo: repo}
}

func (s *PlanService) CreatePlan(ctx context.Context, plan *domain.CreatePlanInput) (*domain.Plan, error) {
	result, err := s.repo.CreatePlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("plan_service.CreatePlan: %w", err)
	}
	return result, nil
}

func (s *PlanService) GetPlan(ctx context.Context, planId string) (*domain.UserPlan, error) {
	result, err := s.repo.GetPlan(ctx, planId)
	if err != nil {
		return nil, fmt.Errorf("plan_service.GetPlan: %w", err)
	}
	return result, nil
}
