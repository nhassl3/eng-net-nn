package postgres

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/db"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

type PlanRepo struct {
	db *db.Store
}

func NewPlanRepo(db *db.Store) *PlanRepo {
	return &PlanRepo{
		db: db,
	}
}

func (r *PlanRepo) CreatePlan(ctx context.Context, params *domain.CreatePlanInput) (*domain.Plan, error) {
	plan, err := r.db.RequestPlan(ctx, db.RequestPlanParams{
		FullName:        stringToNullable(params.FullName),
		Direction:       params.Direction,
		TaskDescription: stringToNullable(params.TaskDescription),
		Email:           params.EmailToFeedback,
	})
	if err != nil {
		if mapped, ok := mapConstraintErr(err, domain.ErrPlanRequestAlreadyExists); ok {
			return nil, mapped
		}
		return nil, fmt.Errorf("plan_repository.CreatePlan: %w", err)
	}
	return new(mapPlan(plan)), nil
}

func (r *PlanRepo) GetUserPlan(ctx context.Context, planUID, userUID string) (*domain.UserPlan, error) {
	userId := stringToNullable(userUID)
	if userId.Valid == false {
		return nil, domain.ErrInvalidToken
	}

	planID, err := string2UUID(planUID)
	if err != nil {
		return nil, domain.ErrPlanRequestNotExists
	}

	userPlan, err := r.db.GetUserPlan(ctx, db.GetUserPlanParams{
		UserID: userId,
		PlanID: planID,
	})
	if err != nil {
		if mapped, ok := mapNotFound(err, domain.ErrPlanRequestNotExists); ok {
			return nil, mapped
		}
		return nil, fmt.Errorf("plan_repository.GetPlan: %w", err)
	}

	return &domain.UserPlan{
		User: new(mapUser(userPlan.User)),
		Plan: new(mapPlan(userPlan.Plan)),
	}, nil
}

func (r *PlanRepo) GetPlan(ctx context.Context, planUID string) (*domain.UserPlan, error) {
	planID, err := string2UUID(planUID)
	if err != nil {
		return nil, domain.ErrPlanRequestNotExists
	}

	userPlan, err := r.db.GetUserPlan(ctx, db.GetUserPlanParams{
		PlanID: planID,
	})
	if err != nil {
		if mapped, ok := mapNotFound(err, domain.ErrPlanRequestNotExists); ok {
			return nil, mapped
		}
		return nil, fmt.Errorf("plan_repository.GetPlan: %w", err)
	}

	return &domain.UserPlan{
		User: new(mapUser(userPlan.User)),
		Plan: new(mapPlan(userPlan.Plan)),
	}, nil
}

func (r *PlanRepo) GetDirection(ctx context.Context, directionId int32) (string, error) {
	name, err := r.db.GetDirection(ctx, directionId)
	if err != nil {
		if mapped, ok := mapNotFound(err, domain.ErrDirectionNotFound); ok {
			return "", mapped
		}
		return "", fmt.Errorf("plan_repository.GetDirection: %w", err)
	}
	return name.String, nil
}

func (r *PlanRepo) CreateLinkRequest(ctx context.Context, userId, planId string) error {
	userID, err := string2UUID(userId)
	if err != nil {
		return domain.ErrUserNotExists
	}

	planID, err := string2UUID(planId)
	if err != nil {
		return domain.ErrPlanRequestNotExists
	}

	if err := r.db.CreateLinkRequest(ctx, db.CreateLinkRequestParams{
		UserID: userID,
		PlanID: planID,
	}); err != nil {
		if mapped, ok := mapConstraintErr(err, domain.ErrPlanRequestAlreadyExists); ok {
			return mapped
		}
		if mapped, ok := mapNotFound(err, domain.ErrPlanRequestNotExists); ok {
			return mapped
		}
		return fmt.Errorf("plan_repository.CreateLinkRequest: %w", err)
	}
	return nil
}

// GetAllPlans returns plans
func (r *PlanRepo) GetAllPlans(ctx context.Context, limit, offset int32) (*domain.Plans, error) {
	allUsersPlans, err := r.db.GetAllPlans(ctx, db.GetAllPlansParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		if mapped, ok := mapNotFound(err, domain.ErrPlanRequestNotExists); ok {
			return nil, mapped
		}
		return nil, fmt.Errorf("plan_repository.GetAllPlans: %w", err)
	}

	total, err := r.db.CountPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("plan_repository.GetAllPlans: %w", err)
	}

	return mapPlans(allUsersPlans, total), nil
}

func mapPlan(plan db.Plan) domain.Plan {
	return domain.Plan{
		UUID:            uuid2String(plan.ID),
		FullName:        plan.FullName.String,
		Direction:       plan.Direction,
		TaskDescription: plan.TaskDescription.String,
		EmailToFeedback: plan.Email,
		CreatedAt:       pgTimeTZ(plan.CreatedAt),
	}
}

func mapPlans(plans []db.Plan, total int64) *domain.Plans {
	domainPlan := make([]domain.Plan, len(plans))
	for i := range plans {
		domainPlan[i] = mapPlan(plans[i])
	}
	return &domain.Plans{
		Plans: domainPlan,
		Total: int(total),
	}
}
