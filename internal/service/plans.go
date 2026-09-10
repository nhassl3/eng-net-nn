package service

import (
	"context"
	"fmt"
	"strconv"

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

// CreatePlan saves the plan request to the DB and notifies the owner by
// email. SMTP errors are ignored and do not fail the request.
func (s *PlansService) CreatePlan(ctx context.Context, plan *domain.CreatePlanInput, userId *string) (*domain.Plan, error) {
	result, err := s.repo.CreatePlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("plan_service.CreatePlan: %w", err)
	}

	if userId != nil && *userId != "" {
		if err := s.repo.CreateLinkRequest(ctx, *userId, result.UUID); err != nil {
			return nil, fmt.Errorf("plan_service.CreatePlan.CreateLinkRequest: %w", err)
		}
	}

	directionName := strconv.Itoa(int(plan.Direction))
	name, err := s.repo.GetDirection(ctx, plan.Direction)
	if name == "" {
		if err != nil {
			return nil, fmt.Errorf("plan_serivce.CreatePlan: failed to load direction: %w", err)
		}
		return nil, domain.ErrDirectionNotFound
	}
	directionName = name

	_ = s.mailer.NotifyNewPlan(ctx, &domain.CreatePlanInputEmail{
		FullName:        result.FullName,
		TaskDescription: result.TaskDescription,
		Direction:       directionName,
		EmailToFeedback: plan.EmailToFeedback,
	})
	_ = s.mailer.NotifyUserAboutPlan(ctx, plan.EmailToFeedback)

	return result, nil
}

func (s *PlansService) GetUserPlan(ctx context.Context, planUID, userUID string) (*domain.UserPlan, error) {
	result, err := s.repo.GetUserPlan(ctx, planUID, userUID)
	if err != nil {
		return nil, fmt.Errorf("plan_service.GetPlan: %w", err)
	}
	return result, nil
}

func (s *PlansService) GetPlan(ctx context.Context, planUID string) (*domain.UserPlan, error) {
	result, err := s.repo.GetPlan(ctx, planUID)
	if err != nil {
		return nil, fmt.Errorf("plan_service.GetPlan: %w", err)
	}
	return result, nil
}

func (s *PlansService) GetAllPlans(ctx context.Context, limit, offset int32) (*domain.Plans, error) {
	result, err := s.repo.GetAllPlans(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("plan_service.GetAllPlans: %w", err)
	}
	return result, nil
}
