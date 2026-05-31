package postgres

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/db"
)

type AdminRepo struct {
	db *db.Store
}

func NewAdminRepo(db *db.Store) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) IsAdmin(ctx context.Context, userID string) (bool, error) {
	ok, err := r.db.IsAdmin(ctx, string2UUID(userID))
	if err != nil {
		return false, fmt.Errorf("admin_repo.IsAdmin: %w", err)
	}
	return ok, nil
}

func (r *AdminRepo) AddAdmin(ctx context.Context, userID string) error {
	if err := r.db.AddAdmin(ctx, string2UUID(userID)); err != nil {
		return fmt.Errorf("admin_repo.AddAdmin: %w", err)
	}
	return nil
}
