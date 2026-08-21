package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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
		return nil, fmt.Errorf("plan_repository.CreatePlan: %w", err)
	}
	return new(mapPlan(plan)), nil
}

func (r *PlanRepo) GetPlan(ctx context.Context, planId string) (*domain.UserPlan, error) {
	userPlan, err := r.db.GetResponseFromRequest(ctx, string2UUID(planId))
	if err != nil {
		return nil, fmt.Errorf("plan_repository.GetPlan: %w", err)
	}

	user, err := r.db.GetUser(ctx, db.GetUserParams{
		ID: pgtype.UUID{Bytes: userPlan.UserID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("plan_repository.GetPlan: %w", err)
	}

	plan, err := r.db.GetPlan(ctx, string2UUID(planId))
	if err != nil {
		return nil, fmt.Errorf("plan_repository.GetPlan: %w", err)
	}

	return &domain.UserPlan{
		User: new(mapUser(user)),
		Plan: new(mapPlan(plan)),
	}, nil
}

func (r *PlanRepo) GetDirection(ctx context.Context, directionId int32) (string, error) {
	name, err := r.db.GetDirection(ctx, directionId)
	if err != nil {
		return "", fmt.Errorf("plan_repository.GetDirection: %w", err)
	}
	return name.String, nil
}

func (r *PlanRepo) CreateLinkRequest(ctx context.Context, userId, planId string) error {
	return r.db.CreateLinkRequest(ctx, db.CreateLinkRequestParams{
		UserID: string2UUID(userId),
		PlanID: string2UUID(planId),
	})
}

// GetAllPlans returns plans
// TODO: set pages
func (r *PlanRepo) GetAllPlans(ctx context.Context) (*domain.Plans, error) {
	allUsersPlans, err := r.db.GetAllPlans(ctx, db.GetAllPlansParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("plan_repository.GetAllPlans: %w", err)
	}
	return mapPlans(allUsersPlans), nil
}

func mapPlan(plan db.Plan) domain.Plan {
	return domain.Plan{
		UUID:            uuid2String(plan.ID),
		FullName:        plan.FullName.String,
		Direction:       plan.Direction,
		TaskDescription: plan.TaskDescription.String,
		EmailToFeedback: plan.Email,
		CreatedAt:       pgTimeTZ(plan.CreatedAt, time.UTC),
	}
}

func mapPlans(plans []db.Plan) *domain.Plans {
	domainPlan := make([]domain.Plan, len(plans))
	for i := range plans {
		domainPlan[i] = mapPlan(plans[i])
	}
	return &domain.Plans{
		Plans: domainPlan,
		Total: len(plans),
	}
}
