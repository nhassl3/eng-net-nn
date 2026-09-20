package postgres

import (
	"context"
	"fmt"

	"github.com/nhassl3/IpBuild-backend/internal/db"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

type PartnerRepo struct {
	db *db.Store
}

func NewPartnerRepo(db *db.Store) *PartnerRepo {
	return &PartnerRepo{
		db: db,
	}
}

func (repo *PartnerRepo) AddPartner(ctx context.Context, userUID string) error {
	id, err := string2UUID(userUID)
	if err != nil {
		return domain.ErrUserNotExists
	}

	if err := repo.db.AddPartner(ctx, id); err != nil {
		// TODO: catch created in trigger exception and linking with domain error
		return fmt.Errorf("add partner error: %w", err)
	}
	return nil
}
