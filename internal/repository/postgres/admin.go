package postgres

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/db"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

type AdminRepo struct {
	db *db.Store
}

func NewAdminRepo(db *db.Store) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) IsAdmin(ctx context.Context, userID string) (bool, error) {
	id, err := string2UUID(userID)
	if err != nil {
		return false, domain.ErrUserNotExists
	}

	ok, err := r.db.IsAdmin(ctx, id)
	if err != nil {
		return false, fmt.Errorf("admin_repo.IsAdmin: %w", err)
	}
	return ok, nil
}

func (r *AdminRepo) AddAdmin(ctx context.Context, userID string) error {
	id, err := string2UUID(userID)
	if err != nil {
		return domain.ErrUserNotExists
	}

	if err := r.db.AddAdmin(ctx, id); err != nil {
		return fmt.Errorf("admin_repo.AddAdmin: %w", err)
	}
	return nil
}
